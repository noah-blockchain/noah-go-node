package bus

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
	"math/big"
)

type WaitList interface {
	AddToWaitList(address types.Address, pubkey types.Pubkey, coin types.CoinID, value *big.Int)
}
