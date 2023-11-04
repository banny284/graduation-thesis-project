package websocketcontroller

// banny 284

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"price-feed-oracle/provider/telemetry"
	"price-feed-oracle/types"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	defaultReadNewWsMessage   = 50 * time.Microsecond
	defaultMaxConnectionTime  = 23 * time.Hour
	defaultPingDuration       = 15 * time.Second
	disablePingDuration       = time.Duration(0)
	startingReconnectDuration = 5 * time.Second
	maxRetryMultiplier        = 25
)

type (
	MessageHandler func(int, []byte)

	SubscribeHandler func(...types.CurrencyPair) []interface{}

	Name string

	WebsocketController struct {
		parentCtx    context.Context
		websocketCtx context.Context

		websocketCancelFunc context.CancelFunc
		providerName        Name
		websocketUrl        url.URL
		pair                []types.CurrencyPair
		messageHandler      MessageHandler
		subscribeHandler    SubscribeHandler

		pingDuration    time.Duration
		pingMessage     string
		pingMessageType uint
		logger          zerolog.Logger

		mtx              sync.RWMutex
		client           *websocket.Conn
		reconnectCounter uint
	}
)

// 				//
// server side 	//
// 				//

// create new websocket controller
func NewWebsocketController(
	parentCtx context.Context,
	providerName Name,
	websocketUrl url.URL,
	pair []types.CurrencyPair,
	messageHandler MessageHandler,
	subscribeMessage SubscribeHandler,
	pingDuration time.Duration,
	pingMessage string,
	pingMessageType uint,
	logger zerolog.Logger,
) *WebsocketController {
	return &WebsocketController{
		parentCtx:        parentCtx,
		providerName:     providerName,
		websocketUrl:     websocketUrl,
		pair:             pair,
		messageHandler:   messageHandler,
		subscribeHandler: subscribeMessage,
		pingDuration:     pingDuration,
		pingMessage:      pingMessage,
		pingMessageType:  pingMessageType,
		logger:           logger,
	}
}

func (wsc *WebsocketController) ping() error {
	wsc.mtx.Lock()
	defer wsc.mtx.Unlock()

	if wsc.client == nil {
		return fmt.Errorf("don't have a websocket connection")
	}

	if err := wsc.client.WriteMessage(int(wsc.pingMessageType), []byte(wsc.pingMessage)); err != nil {
		wsc.logger.Err(fmt.Errorf(types.ErrWebSocketSend.Error(), wsc.providerName, err)).Send()
	}

	return nil
}

func pingLoop(ctx context.Context, pingDuration time.Duration, ping func() error) {

	if pingDuration == disablePingDuration {
		return
	}

	ticker := time.NewTicker(pingDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := ping(); err != nil {
				return
			}
		}
	}
}

func (wsc *WebsocketController) interateRetryCounter() time.Duration {
	if wsc.reconnectCounter < 25 {
		wsc.reconnectCounter++
	}

	multiplier := math.Pow(float64(wsc.reconnectCounter), 2)

	return time.Duration(multiplier) * startingReconnectDuration
}

// 				//
// client side	//
// 				//

// connect to websocket
func (wsc *WebsocketController) connect() error {
	wsc.mtx.Lock()
	defer wsc.mtx.Unlock()

	wsc.logger.Debug().Msg("Connecting to websocket")
	conn, reps, err := websocket.DefaultDialer.Dial(wsc.websocketUrl.String(), nil)
	if err != nil {
		return fmt.Errorf(types.ErrWebSocketDial.Error(), wsc.providerName, err)
	}

	defer reps.Body.Close()

	wsc.client = conn
	wsc.websocketCtx, wsc.websocketCancelFunc = context.WithCancel(wsc.parentCtx)
	wsc.client.SetPingHandler(wsc.pingHandler)
	wsc.reconnectCounter = 0

	return nil
}

// when sv sends a ping then client send a pong back
func (wsc *WebsocketController) pingHandler(appData string) error {
	if err := wsc.client.WriteMessage(websocket.PongMessage, []byte("pong")); err != nil {
		wsc.logger.Error().Err(err).Msg("error sending pong")
	}
	return nil
}

// subscribe to websocket
func (wsc *WebsocketController) SendJSON(msg interface{}) error {
	wsc.mtx.Lock()
	defer wsc.mtx.Unlock()

	if wsc.client == nil {
		return fmt.Errorf("unable to send JSON on a closed connection")
	}

	wsc.logger.Debug().
		Interface("msg", msg).
		Msg("sending websocket message")

	if err := wsc.client.WriteJSON(msg); err != nil {
		return fmt.Errorf(types.ErrWebSocketSend.Error(), wsc.providerName, err)
	}
	return nil
}

func (wsc *WebsocketController) subscribe(
	msgs []interface{},
) error {
	telemetry.TelemetryWebsocketSubscribeCurrencyPairs(telemetry.Name(wsc.providerName), len(wsc.pair))

	for _, msg := range msgs {
		if err := wsc.SendJSON(msg); err != nil {
			return fmt.Errorf(types.ErrWebSocketSend.Error(), wsc.providerName, err)
		}
	}

	return nil
}

func (wsc *WebsocketController) AddSubscriptionMsgs(msgs []interface{}) error {
	return wsc.subscribe(msgs)
}

func (wsc *WebsocketController) AddPairs(pairs []types.CurrencyPair) error {
	return wsc.subscribe(wsc.subscribeHandler(pairs...))
}
