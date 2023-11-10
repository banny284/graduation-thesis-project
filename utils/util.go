package utils

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	cosmosType "github.com/cosmos/cosmos-sdk/types"
)

const (
	defaultTimeout = 10 * time.Second
)

// preventRedirect avoid any redirect in the http.Client the request call
// will not return an error, but a valid response with redirect response code.
func preventRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func newDefaultHTTPClient() *http.Client {
	return newHTTPClientWithTimeout(defaultTimeout)
}

func newHTTPClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: preventRedirect,
	}
}

// PastUnixTime returns a millisecond timestamp that represents the unix time
// minus t.
func PastUnixTime(t time.Duration) int64 {
	return time.Now().Add(t*-1).Unix() * int64(time.Second/time.Millisecond)
}

// SecondsToMilli converts seconds to milliseconds for our unix timestamps.
func SecondsToMilli(t int64) int64 {
	return t * int64(time.Second/time.Millisecond)
}

func checkHTTPStatus(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func StrToDec(str string) cosmosType.Dec {
	if strings.Contains(str, ".") {
		split := strings.Split(str, ".")
		if len(split[1]) > 18 {
			// sdk.MustNewDecFromStr will panic if decimal precision is greater than 18
			str = split[0] + "." + split[1][0:18]
		}
	}
	dec, err := cosmosType.NewDecFromStr(str)
	if err != nil {
		dec = cosmosType.Dec{}
	}

	return dec
}

func floatToDec(f float64) cosmosType.Dec {
	return StrToDec(strconv.FormatFloat(f, 'f', -1, 64))
}

func InvertDec(d cosmosType.Dec) cosmosType.Dec {
	if d.IsZero() || d.IsNil() {
		return cosmosType.ZeroDec()
	}
	return cosmosType.NewDec(1).Quo(d)
}
