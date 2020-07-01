package bus

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
	"math/big"
)

type Coins interface {
	GetCoin(types.CoinSymbol) *Coin
	SubCoinVolume(types.CoinSymbol, *big.Int)
	SubCoinReserve(types.CoinSymbol, *big.Int)
}

type Coin struct {
	Name    string
	Crr     uint
	Symbol  types.CoinSymbol
	Volume  *big.Int
	Reserve *big.Int
}
