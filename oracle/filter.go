package oracle

import (
	"fmt"
	"price-feed-oracle/provider"
	"price-feed-oracle/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
)

// defaultDeviationThreshold defines how many 𝜎 a provider can be away
// from the mean without being considered faulty. This can be overridden
// in the config.
var defaultDeviationThreshold = sdk.MustNewDecFromStr("1.0")

// FilterTickerDeviations finds the standard deviations of the prices of
// all assets, and filters out any providers that are not within 2𝜎 of the mean.

func isBetween(p, mean, margin sdk.Dec) bool {
	return p.GTE(mean.Sub(margin)) &&
		p.LTE(mean.Add(margin))
}

func FilterTickerDeviations(
	logger zerolog.Logger,
	symbol string,
	tickerPrices map[provider.Name]types.TickerPrice,
	deviationThreshold sdk.Dec,
) (map[provider.Name]types.TickerPrice, error) {
	if deviationThreshold.IsNil() {
		deviationThreshold = defaultDeviationThreshold
	}

	prices := []sdk.Dec{}
	for _, tickerPrice := range tickerPrices {
		prices = append(prices, tickerPrice.Price)
	}

	deviation, mean, err := StandardDeviation(prices)
	if err != nil {
		return tickerPrices, err
	}

	// We accept any prices that are within (2 * T)𝜎, or for which we couldn't get 𝜎.
	// T is defined as the deviation threshold, either set by the config
	// or defaulted to 1.
	filteredPrices := map[provider.Name]types.TickerPrice{}
	for providerName, tickerPrice := range tickerPrices {
		if isBetween(tickerPrice.Price, mean, deviation.Mul(deviationThreshold)) {
			filteredPrices[providerName] = tickerPrice
		} else {
			telemetry.IncrCounter(1, "failure", "provider", "type", "ticker")
			logger.Debug().
				Str("symbol", symbol).
				Str("provider", providerName.String()).
				Str("price", tickerPrice.Price.String()).
				Str("mean", mean.String()).
				Str("margin", deviation.Mul(deviationThreshold).String()).
				Msg("deviating price")
		}
	}

	return filteredPrices, nil
}

func ComputeVWAP(tickers []types.TickerPrice) (sdk.Dec, error) {
	if len(tickers) == 0 {
		return sdk.Dec{}, fmt.Errorf("no tickers supplied")
	}

	volumeSum := sdk.ZeroDec()

	for _, tp := range tickers {
		volumeSum = volumeSum.Add(tp.Volume)
	}

	weightedPrice := sdk.ZeroDec()

	for _, tp := range tickers {
		volume := tp.Volume
		if volumeSum.Equal(sdk.ZeroDec()) {
			volume = sdk.NewDec(1)
		}

		// weightedPrice = Σ {P * V} for all TickerPrice
		weightedPrice = weightedPrice.Add(tp.Price.Mul(volume))
	}

	if volumeSum.Equal(sdk.ZeroDec()) {
		volumeSum = sdk.NewDec(int64(len(tickers)))
	}

	return weightedPrice.Quo(volumeSum), nil
}

// StandardDeviation returns standard deviation and mean of assets.
// Will skip calculating for an asset if there are less than 3 prices.
func StandardDeviation(prices []sdk.Dec) (sdk.Dec, sdk.Dec, error) {
	// Skip if standard deviation would not be meaningful
	if len(prices) < 3 {
		err := fmt.Errorf("not enough values to calculate deviation")
		return sdk.Dec{}, sdk.Dec{}, err
	}

	sum := sdk.ZeroDec()

	for _, price := range prices {
		sum = sum.Add(price)
	}

	numPrices := int64(len(prices))
	mean := sum.QuoInt64(numPrices)
	varianceSum := sdk.ZeroDec()

	for _, price := range prices {
		deviation := price.Sub(mean)
		varianceSum = varianceSum.Add(deviation.Mul(deviation))
	}

	variance := varianceSum.QuoInt64(numPrices)

	deviation, err := variance.ApproxSqrt()
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}

	return deviation, mean, nil
}
