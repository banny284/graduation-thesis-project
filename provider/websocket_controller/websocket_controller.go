package websocketcontroller

import (
	"context"
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

// create
