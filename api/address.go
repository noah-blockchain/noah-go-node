package api

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
)

type AddressResponse struct {
	Balance          map[string]string `json:"balance"`
	TransactionCount uint64            `json:"transaction_count"`
}

func Address(address types.Address, height int) (*AddressResponse, error) {
	cState, err := GetStateForHeight(height)
	if err != nil {
		return nil, err
	}

	cState.RLock()
	defer cState.RUnlock()

	balances := cState.Accounts().GetBalances(address)

	response := AddressResponse{
		Balance:          make([]BalanceItem, len(balances)),
		TransactionCount: cState.Accounts().GetNonce(address),
	}

	isBaseCoinExists := false
	for k, b := range balances {
		response.Balance[k] = BalanceItem{
			CoinID: b.Coin.ID.Uint32(),
			Symbol: b.Coin.GetFullSymbol(),
			Value:  b.Value.String(),
		}

		if b.Coin.ID.IsBaseCoin() {
			isBaseCoinExists = true
		}
	}

	if !isBaseCoinExists {
		response.Balance = append(response.Balance, BalanceItem{
			CoinID: types.GetBaseCoinID().Uint32(),
			Symbol: types.GetBaseCoin().String(),
			Value:  "0",
		})
	}

	return &response, nil
}
