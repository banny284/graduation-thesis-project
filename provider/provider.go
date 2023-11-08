package provider

import (
	"context"
	"net/http"
	"net/url"
	"price-feed-oracle/types"
	"sync"
	"time"

	Wsc "price-feed-oracle/provider/websocket_controller"

	"github.com/rs/zerolog"
)

const (
	defaultTimeout       = 10 * time.Second
	staleTickersCutoff   = 1 * time.Minute
	providerCandlePeriod = 10 * time.Minute

	ProviderOkx          Name = "okx"
	ProviderBinance      Name = "binance"
	ProviderCoinbase     Name = "coinbase"
	ProviderUniswapV3    Name = "uniswapv3"
	ProviderPancakeV3Bsc Name = "pancakev3_bsc"
)

type (
	// variable
	Name string

	// provider => asset => ticker price
	AggregatedTickerPrice map[string]map[string]types.TickerPrice

	// interface
	Provider interface {
		GetTickerPrice(...types.CurrencyPair) (map[string]types.TickerPrice, error)

		GetAvailableCurrencyPairs() (map[string]struct{}, error)
	}

	// struct
	provider struct {
		ctx       context.Context
		endpoints Endpoint
		httpBase  string
		http      *http.Client
		logger    zerolog.Logger
		mtx       sync.RWMutex
		pairs     map[string]types.CurrencyPair
		inverse   map[string]types.CurrencyPair
		tickers   map[string]types.TickerPrice
		contracts map[string]string
		websocket *Wsc.WebsocketController
	}

	Endpoint struct {
		Name            Name
		Urls            []string
		Websocket       string
		WebsocketPath   string
		PollInterval    time.Duration
		PingDuration    time.Duration
		PingType        uint
		PingMessage     string
		ContractAddress map[string]string
	}
)

func newDefaultHttpClient() *http.Client {
	return &http.Client{
		Timeout: defaultTimeout,
	}
}

func (p *provider) Init(
	ctx context.Context,
	endpoints Endpoint,
	logger zerolog.Logger,
	pairs []types.CurrencyPair,
	websocketMessageHandler Wsc.MessageHandler,
	websocketSubscribeHandler Wsc.SubscribeHandler,
) {
	p.ctx = ctx
	p.endpoints = endpoints

	p.logger = logger.With().Str("provider", string(p.endpoints.Name)).Logger()

	p.http = newDefaultHttpClient()
	p.httpBase = p.endpoints.Urls[0]
	p.tickers = map[string]types.TickerPrice{}
	p.contracts = p.endpoints.ContractAddress

	if p.endpoints.Websocket != "" {
		wsUrl := url.URL{
			Scheme: "wss",
			Host:   p.endpoints.Websocket,
			Path:   p.endpoints.WebsocketPath,
		}

		p.websocket = Wsc.NewWebsocketController(
			ctx,
			Wsc.Name(p.endpoints.Name),
			wsUrl,
			pairs,
			websocketMessageHandler,
			websocketSubscribeHandler,
			p.endpoints.PingDuration,
			p.endpoints.PingMessage,
			p.endpoints.PingType,
			p.logger,
		)

		go p.websocket.Start()

	}
}

func (p *provider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	tickers := make(
		map[string]types.TickerPrice,
		len(pairs),
	)

	for _, pair := range pairs {
		symbol := pair.String()
		ticker, ok := p.tickers[symbol]
		if !ok {
			p.logger.Warn().Str("pair", symbol).Msg("Missing ticker price for pair.")
		} else {
			if ticker.Price.IsZero() {
				p.logger.Warn().Str("pair", symbol).Msg("Ticker price is zero.")
				continue
			}
			if time.Since(ticker.Time) > staleTickersCutoff {
				p.logger.Warn().Str("pair", symbol).Time("time", ticker.Time).Msg("Ticker price is stale.")

			} else {
				tickers[symbol] = ticker
			}
		}
	}

	return tickers, nil
}
