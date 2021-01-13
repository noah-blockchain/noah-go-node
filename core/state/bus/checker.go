package bus

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
	"math/big"
)

type Checker interface {
	AddCoin(types.CoinID, *big.Int, ...string)
	AddCoinVolume(types.CoinID, *big.Int)
}
