package types

import (
	"fmt"

	cosmosType "github.com/cosmos/cosmos-sdk/types"
)

type CandlePrice struct {
	Price     cosmosType.Dec `json:"price"`
	Volume    cosmosType.Dec `json:"volume"`
	TimeStamp int64          `json:"time_stamp"`
}

func NewCandlePrice(
	provider string,
	symbol string,
	lastPrice string,
	volume string,
	timeStamp int64,
) (CandlePrice, error) {
	priceDec, err := cosmosType.NewDecFromStr(lastPrice)
	if err != nil {
		return CandlePrice{}, fmt.Errorf("failed to parse %s price for %s of provider %s: %v", lastPrice, symbol, provider, err)
	}

	volumeDec, err := cosmosType.NewDecFromStr(volume)
	if err != nil {
		return CandlePrice{}, fmt.Errorf("failed to parse %s volume for %s of provider %s: %v", volume, symbol, provider, err)
	}

	return CandlePrice{
		Price:     priceDec,
		Volume:    volumeDec,
		TimeStamp: timeStamp,
	}, nil
}
