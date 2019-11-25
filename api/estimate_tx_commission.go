package api

import (
	"fmt"
	"math/big"

	"github.com/noah-blockchain/noah-go-node/core/transaction"
	"github.com/noah-blockchain/noah-go-node/formula"
	"github.com/noah-blockchain/noah-go-node/rpc/lib/types"
)

type TxCommissionResponse struct {
	Commission string `json:"commission"`
}

func EstimateTxCommission(tx []byte, height int) (*TxCommissionResponse, error) {
	cState, err := GetStateForHeight(height)
	if err != nil {
		return nil, err
	}

	decodedTx, err := transaction.TxDecoder.DecodeFromBytes(tx)
	if err != nil {
		return nil, rpctypes.RPCError{Code: 400, Message: "Cannot decode transaction", Data: err.Error()}
	}

	commissionInBaseCoin := decodedTx.CommissionInBaseCoin()
	commission := big.NewInt(0).Set(commissionInBaseCoin)

	if !decodedTx.GasCoin.IsBaseCoin() {
		coin := cState.GetStateCoin(decodedTx.GasCoin)

		if coin.ReserveBalance().Cmp(commissionInBaseCoin) < 0 {
			return nil, rpctypes.RPCError{Code: 400, Message: fmt.Sprintf("Coin reserve balance is not sufficient for transaction. Has: %s, required %s",
				coin.ReserveBalance().String(), commissionInBaseCoin.String())}
		}

		commission = formula.CalculateSaleAmount(coin.Volume(), coin.ReserveBalance(), coin.Data().Crr, commissionInBaseCoin)
	}

	return &TxCommissionResponse{
		Commission: commission.String(),
	}, nil
}
