package websocketcontroller

// banny 284

import (
	"context"
	"fmt"
	"net/url"
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

	SubscribeMessage func(...types.CurrencyPair) []interface{}

	Name string

	WebsocketController struct {
		parentCtx    context.Context
		websocketCtx context.Context

		websocketCancelFunc context.CancelFunc
		providerName        Name
		websocketUrl        url.URL
		pair                []types.CurrencyPair
		messageHandler      MessageHandler
		subscribeMessage    SubscribeMessage

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
	subscribeMessage SubscribeMessage,
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
		subscribeMessage: subscribeMessage,
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
