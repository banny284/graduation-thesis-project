package provider

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	telemetry "price-feed-oracle/provider/telemetry"
	Wsc "price-feed-oracle/provider/websocket_controller"
	"price-feed-oracle/types"
	util "price-feed-oracle/utils"
	"sync"
	"time"

	cosmosType "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
)

const (
	defaultTimeout       = 10 * time.Second
	staleTickersCutoff   = 1 * time.Minute
	providerCandlePeriod = 10 * time.Minute

	ProviderOkx          Name = "okx"
	ProviderBinance      Name = "binance"
	ProviderBinanceUS    Name = "binanceus"
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

		SubscribeCurrencyPairs(...types.CurrencyPair) error

		CurrencyPairToProviderPair(types.CurrencyPair) string
	}

	PollingProvider interface {
		Poll() error
	}

	CurrencyPairToProviderSymbol func(types.CurrencyPair) string

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
	p.endpoints.SetDefaultEnpoint()
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

func (p *provider) GetTickerPrice(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
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

func (p *provider) CurrencyPairToProviderPair(pair types.CurrencyPair) string {
	return pair.String()
}

func (p *provider) SubscribeCurrencyPairs(pairs ...types.CurrencyPair) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	newPair := p.addPairs(pairs...)
	if p.endpoints.Websocket != "" {
		return nil
	}

	return p.websocket.AddPairs(newPair)
}

func (p *provider) addPairs(pairs ...types.CurrencyPair) []types.CurrencyPair {
	newPairs := []types.CurrencyPair{}
	for _, pair := range pairs {
		_, ok := p.pairs[pair.String()]
		if !ok {
			newPairs = append(newPairs, pair)
		}
	}
	return newPairs
}

func (p *provider) makeHttpRequest(url string, method string, body []byte, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	res, err := p.http.Do(req)
	if err != nil {
		p.logger.Warn().
			Err(err).
			Msg("http request failed")
		return nil, err
	}

	if res.StatusCode != 200 {
		p.logger.Warn().
			Int("code", res.StatusCode).
			Msg("http request returned invalid status")
		if res.StatusCode == 429 || res.StatusCode == 418 {
			p.logger.Warn().
				Str("url", url).
				Str("retry_after", res.Header.Get("Retry-After")).
				Msg("http ratelimited")
		}
		return nil, fmt.Errorf("http request returned invalid status")
	}
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (p *provider) httpRequest(path string, method string, body []byte, headers map[string]string) ([]byte, error) {
	res, err := p.makeHttpRequest(p.httpBase+path, method, body, headers)
	if err != nil {
		p.logger.Warn().
			Str("endpoint", p.httpBase).
			Str("path", path).
			Msg("trying alternate http endpoints")
		for _, endpoint := range p.endpoints.Urls {
			if endpoint == p.httpBase {
				continue
			}
			res, err = p.makeHttpRequest(endpoint+path, method, body, headers)
			if err == nil {
				p.logger.Info().Str("endpoint", endpoint).Msg("selected alternate http endpoint")
				p.httpBase = endpoint
				break
			}
		}
	}
	return res, err
}

func (p *provider) httpGet(path string) ([]byte, error) {
	return p.httpRequest(path, "GET", nil, nil)
}

func (p *provider) httpPost(path string, body []byte) ([]byte, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	return p.httpRequest(path, "POST", body, headers)
}

func (e *Endpoint) SetDefaultEnpoint() {
	var defaults Endpoint

	switch e.Name {

	case ProviderBinance:
		defaults = binanceDefaultEndpoints

	case ProviderOkx:
		defaults = okxDefaultEndpoints

	case ProviderCoinbase:
		defaults = coinbaseDefaultEndpoints
	case ProviderUniswapV3:
		defaults = uniswapv3DefaultEndpoints

	default:
		return
	}

	if e.Urls == nil {
		urls := defaults.Urls
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(
			len(urls),
			func(i, j int) { urls[i], urls[j] = urls[j], urls[i] },
		)
		e.Urls = urls
	}

	if e.Websocket == "" && defaults.Websocket != "" { // don't enable websockets for providers that don't support them
		e.Websocket = defaults.Websocket
	}
	if e.WebsocketPath == "" {
		e.WebsocketPath = defaults.WebsocketPath
	}
	if e.PollInterval == time.Duration(0) {
		e.PollInterval = defaults.PollInterval
	}
	if e.PingDuration == time.Duration(0) {
		e.PingDuration = defaults.PingDuration
	}
	if e.PingType == 0 {
		e.PingType = defaults.PingType
	}
	if e.PingMessage == "" {
		if defaults.PingMessage != "" {
			e.PingMessage = defaults.PingMessage
		} else {
			e.PingMessage = "ping"
		}
	}

	// add default contract addresses, if not already defined
	for symbol, address := range defaults.ContractAddress {
		_, found := e.ContractAddress[symbol]
		if found {
			continue
		}
		e.ContractAddress[symbol] = address
	}
}

func startPolling(p PollingProvider,
	interval time.Duration,
	logger zerolog.Logger,
) {
	logger.Debug().Dur("interval", interval).Msg("Starting polling loop")

	for {
		err := p.Poll()
		if err != nil {
			logger.Error().Err(err).Msg("Polling failed")
		}
		time.Sleep(interval)
	}
}

func (p *provider) setPairs(
	pairs []types.CurrencyPair,
	availablePairs map[string]struct{},
	toProviderSymbol CurrencyPairToProviderSymbol,
) error {
	p.pairs = map[string]types.CurrencyPair{}
	p.inverse = map[string]types.CurrencyPair{}

	if toProviderSymbol == nil {
		toProviderSymbol = func(cp types.CurrencyPair) string {
			return cp.String()
		}
	}

	if availablePairs == nil {
		p.logger.Warn().Msg("availablePairs is not provided")

		for _, pair := range pairs {

			invertedPair := pair.Reverse()

			p.pairs[toProviderSymbol(pair)] = pair
			p.pairs[toProviderSymbol(invertedPair)] = invertedPair
		}

		return nil
	}

	for _, pair := range pairs {
		invertedPair := pair.Reverse()

		providerSymbol := toProviderSymbol(invertedPair)

		// find the pair in availablePairs
		_, found := availablePairs[providerSymbol]
		if found {
			p.inverse[providerSymbol] = pair
			continue
		}

		symbol := toProviderSymbol(pair)
		_, found = availablePairs[providerSymbol]
		if found {
			p.pairs[symbol] = pair
			continue
		}

		p.logger.Error().Str("pair", pair.String()).Msgf("%s not supported by this provider", symbol)
	}

	return nil
}

func (p *provider) setTickerPrice(symbol string, price cosmosType.Dec, volume cosmosType.Dec, timestamp time.Time) {
	if price.IsNil() || price.LTE(cosmosType.ZeroDec()) {
		p.logger.Warn().
			Str("symbol", symbol).
			Msgf("price is %s", price)
		return
	}

	if volume.IsZero() {
		p.logger.Debug().
			Str("symbol", symbol).
			Msg("volume is zero")
	}

	// check if price needs to be inverted
	pair, inverse := p.inverse[symbol]
	if inverse {
		volume = volume.Mul(price)
		price = util.InvertDec(price)

		p.tickers[pair.String()] = types.TickerPrice{
			Price:  price,
			Volume: volume,
			Time:   timestamp,
		}

		telemetry.TelemetryProviderPrice(
			telemetry.Name(p.endpoints.Name),
			pair.String(),
			float32(price.MustFloat64()),
			float32(volume.MustFloat64()),
		)

		return
	}

	pair, found := p.pairs[symbol]
	if !found {
		p.logger.Error().
			Str("symbol", symbol).
			Msg("symbol not found")
		return
	}

	p.tickers[pair.String()] = types.TickerPrice{
		Price:  price,
		Volume: volume,
		Time:   timestamp,
	}

	telemetry.TelemetryProviderPrice(
		telemetry.Name(p.endpoints.Name),
		pair.String(),
		float32(price.MustFloat64()),
		float32(volume.MustFloat64()),
	)
}

func (p *provider) isPair(symbol string) bool {
	if _, found := p.pairs[symbol]; found {
		return true
	}

	if _, found := p.inverse[symbol]; found {
		return true
	}

	return false
}

func (p *provider) getAllPairs() map[string]types.CurrencyPair {
	pairs := map[string]types.CurrencyPair{}

	for symbol, pair := range p.pairs {
		pairs[symbol] = pair
	}

	for symbol, pair := range p.inverse {
		pairs[symbol] = pair
	}

	return pairs
}

func (p *provider) getAvailablePairsFromContracts() (map[string]struct{}, error) {
	symbols := map[string]struct{}{}
	for symbol := range p.contracts {
		symbols[symbol] = struct{}{}
	}
	return symbols, nil
}

func (p *provider) getContractAddress(pair types.CurrencyPair) (string, error) {
	address, found := p.contracts[pair.String()]
	if found {
		return address, nil
	}

	address, found = p.contracts[pair.Quote+pair.Base]
	if found {
		return address, nil
	}

	err := fmt.Errorf("no contract address found")

	p.logger.Error().
		Str("pair", pair.String()).
		Err(err)

	return "", err
}
