package bus

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
	"math/big"
)

type FrozenFunds interface {
	DeleteCoin(uint64, types.CoinSymbol)
	AddFrozenFund(uint64, types.Address, types.Pubkey, types.CoinSymbol, *big.Int)
}