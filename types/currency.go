package types

import "strings"

type CurrencyPair struct {
	Base  string `json:"base"`
	Quote string `json:"quote"`
}

func (cp CurrencyPair) String() string {
	return strings.ToUpper(cp.Base) + strings.ToUpper(cp.Quote)
}

func (cp CurrencyPair) Join(separator string) string {
	return strings.ToUpper(cp.Base) + separator + strings.ToUpper(cp.Quote)
}

func (cp CurrencyPair) Reverse() CurrencyPair {
	return CurrencyPair{cp.Quote, cp.Base}
}

func ConvertMapCpPairToCpPair(m map[string]CurrencyPair) []CurrencyPair {

	currencyPair := make([]CurrencyPair, 0, len(m))

	for _, v := range m {
		currencyPair = append(currencyPair, v)

	}

	return currencyPair
}
