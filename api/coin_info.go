package api

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/rpc/lib/types"
)

type CoinInfoResponse struct {
	Name           string `json:"name"`
	Symbol         string `json:"symbol"`
	Volume         string `json:"volume"`
	Crr            uint   `json:"crr"`
	ReserveBalance string `json:"reserve_balance"`
	MaxSupply      string `json:"max_supply"`
}

func CoinInfo(coinSymbol string, height int) (*CoinInfoResponse, error) {
	cState, err := GetStateForHeight(height)
	if err != nil {
		return nil, err
	}

	cState.RLock()
	defer cState.RUnlock()

	if coinSymbol == nil && id == nil {
		return nil, rpctypes.RPCError{Code: 404, Message: "Coin not found"}
	}

	if coinSymbol != nil {
		coin = cState.Coins().GetCoinBySymbol(types.StrToCoinBaseSymbol(*coinSymbol), types.GetVersionFromSymbol(*coinSymbol))
		if coin == nil {
			return nil, rpctypes.RPCError{Code: 404, Message: "Coin not found"}
		}
	}

	if id != nil {
		coin = cState.Coins().GetCoin(types.CoinID(*id))
		if coin == nil {
			return nil, rpctypes.RPCError{Code: 404, Message: "Coin not found"}
		}
	}

	var ownerAddress *string
	info := cState.Coins().GetSymbolInfo(coin.Symbol())
	if info != nil && info.OwnerAddress() != nil {
		owner := info.OwnerAddress().String()
		ownerAddress = &owner
	}

	return &CoinInfoResponse{
		ID:             coin.ID().Uint32(),
		Name:           coin.Name(),
		Symbol:         coin.GetFullSymbol(),
		Volume:         coin.Volume().String(),
		Crr:            coin.Crr(),
		ReserveBalance: coin.Reserve().String(),
		MaxSupply:      coin.MaxSupply().String(),
		OwnerAddress:   ownerAddress,
	}, nil
}
