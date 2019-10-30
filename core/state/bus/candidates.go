package bus

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
	"math/big"
)

type Candidates interface {
	DeleteCoin(types.Pubkey, types.CoinSymbol)
	GetStakes(types.Pubkey) []Stake
	Punish(uint64, types.TmAddress) *big.Int
}

type Stake struct {
	Owner    types.Address
	Value    *big.Int
	Coin     types.CoinSymbol
	NoahValue *big.Int
}
