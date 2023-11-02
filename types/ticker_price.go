package types

import (
	"fmt"
	"time"

	cosmosType "github.com/cosmos/cosmos-sdk/types"
)

type TickerPrice struct {
	Price  cosmosType.Dec `json:"price"`
	Volume cosmosType.Dec `json:"volume"`
	Time   time.Time      `json:"time"`
}

func NewTickerPrice(
	price string,
	volume string,
	time time.Time,
) (TickerPrice, error) {
	priceDec, err := cosmosType.NewDecFromStr(price)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse ticker price: %v", err)
	}

	volumeDec, err := cosmosType.NewDecFromStr(volume)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse ticker volume: %v", err)
	}

	return TickerPrice{
		Price:  priceDec,
		Volume: volumeDec,
		Time:   time,
	}, nil

}
