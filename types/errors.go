package types

import (
	cosmosSdkError "github.com/cosmos/cosmos-sdk/types/errors"
)

const ModuleName = "Price-Feed"

var (
	ErrProviderConnection  = cosmosSdkError.Register(ModuleName, 1, "Provider connection error.")
	ErrMissingExchangeRate = cosmosSdkError.Register(ModuleName, 2, "Missing exchange rate for %s .")
	ErrTickerNotFound      = cosmosSdkError.Register(ModuleName, 3, "Failed to get ticker price for %s .")
	ErrCandleNotFound      = cosmosSdkError.Register(ModuleName, 4, "Failed to get candle price for %s .")

	ErrWebSocketDial  = cosmosSdkError.Register(ModuleName, 5, "Error connecting to %s Websocket: %w .")
	ErrWebSocketRead  = cosmosSdkError.Register(ModuleName, 6, "Error reading from %s Websocket: %w .")
	ErrWebSocketSend  = cosmosSdkError.Register(ModuleName, 7, "Error sending to %s Websocket: %w .")
	ErrWebSocketClose = cosmosSdkError.Register(ModuleName, 8, "Error closing %s Websocket: %w .")
)
