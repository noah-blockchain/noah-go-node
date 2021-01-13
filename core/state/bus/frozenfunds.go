package bus

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
	"math/big"
)

type FrozenFunds interface {
	AddFrozenFund(uint64, types.Address, types.Pubkey, uint32, types.CoinID, *big.Int)
}
