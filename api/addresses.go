package api

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
	"math/big"
)

type AddressesResponse struct {
	Address          types.Address     `json:"address"`
	Balance          map[string]string `json:"balance"`
	TransactionCount uint64            `json:"transaction_count"`
}

func Addresses(addresses []types.Address, height int) (*[]AddressesResponse, error) {
	cState, err := GetStateForHeight(height)
	if err != nil {
		return nil, err
	}

	response := make([]AddressesResponse, len(addresses))

	for i, address := range addresses {
		data := AddressesResponse{
			Address:          address,
			Balance:          make(map[string]string),
			TransactionCount: cState.GetNonce(address),
		}

		balances := cState.GetBalances(address)
		for k, v := range balances.Data {
			data.Balance[k.String()] = v.String()
		}

		if _, exists := data.Balance[types.GetBaseCoin().String()]; !exists {
			data.Balance[types.GetBaseCoin().String()] = big.NewInt(0).String()
		}

		response[i] = data
	}

	return &response, nil
}
