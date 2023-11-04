package provider

// import (
// 	"context"
// 	"net/http"
// 	"price-feed-oracle/types"
// 	"sync"
// 	"time"

// 	"github.com/rs/zerolog"
// )

// const (
// 	defaultTimeout       = 10 * time.Second
// 	staleTickersCutoff   = 1 * time.Minute
// 	providerCandlePeriod = 10 * time.Minute

// 	ProviderOkx          Name = "okx"
// 	ProviderBinance      Name = "binance"
// 	ProviderCoinbase     Name = "coinbase"
// 	ProviderUniswapV3    Name = "uniswapv3"
// 	ProviderPancakeV3Bsc Name = "pancakev3_bsc"
// )

// type (
// 	Name string

// 	provider struct {
// 		ctx       context.Context
// 		endpoints Endpoint
// 		httpBase  string
// 		http      *http.Client
// 		logger    zerolog.Logger
// 		mtx       sync.RWMutex
// 		pairs     map[string]types.CurrencyPair
// 		inverse   map[string]types.CurrencyPair
// 		tickers   map[string]types.TickerPrice
// 		contracts map[string]string
// 		websocket *WebsocketController
// 	}
// )
