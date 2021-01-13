package bus

import (
	"fmt"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"math/big"
)

type Coins interface {
	GetCoin(types.CoinID) *Coin
	SubCoinVolume(types.CoinID, *big.Int)
	SubCoinReserve(types.CoinID, *big.Int)
}

type Coin struct {
	ID      types.CoinID
	Name    string
	Crr     uint32
	Symbol  types.CoinSymbol
	Version types.CoinVersion
	Volume  *big.Int
	Reserve *big.Int
}

func (m Coin) GetFullSymbol() string {
	if m.Version == 0 {
		return m.Symbol.String()
	}

	return fmt.Sprintf("%s-%d", m.Symbol, m.Version)
}
