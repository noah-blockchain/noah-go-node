package bus

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
	"math/big"
)

type Accounts interface {
	AddBalance(types.Address, types.CoinID, *big.Int)
}
