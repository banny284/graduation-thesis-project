package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cosmosType "github.com/cosmos/cosmos-sdk/types"
)

func BroadcastTx(clientCtx client.Context, txf tx.Factory, msgs ...cosmosType.Msg) (*cosmosType.TxResponse, error) {
	txf, err := prepareFactory(clientCtx, txf)
	if err != nil {
		return nil, err
	}

	_, adjusted, err := tx.CalculateGas(clientCtx, txf, msgs...)
	if err != nil {
		return nil, err
	}

	txf = txf.WithGas(adjusted)

	unsignedTx, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	unsignedTx.SetFeeGranter(clientCtx.GetFeeGranterAddress())

	if err = tx.Sign(txf, clientCtx.GetFromName(), unsignedTx, true); err != nil {
		return nil, err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(unsignedTx.GetTx())
	if err != nil {
		return nil, err
	}

	return clientCtx.BroadcastTx(txBytes)
}

func prepareFactory(clientCtx client.Context, txf tx.Factory) (tx.Factory, error) {
	from := clientCtx.GetFromAddress()

	if err := txf.AccountRetriever().EnsureExists(clientCtx, from); err != nil {
		return txf, err
	}

	initNum, initSeq := txf.AccountNumber(), txf.Sequence()
	if initNum == 0 || initSeq == 0 {
		num, seq, err := txf.AccountRetriever().GetAccountNumberSequence(clientCtx, from)
		if err != nil {
			return txf, err
		}

		if initNum == 0 {
			txf = txf.WithAccountNumber(num)
		}

		if initSeq == 0 {
			txf = txf.WithSequence(seq)
		}
	}

	return txf, nil
}
