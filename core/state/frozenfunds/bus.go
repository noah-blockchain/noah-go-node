package frozenfunds

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
	"math/big"
)

type Bus struct {
	frozenfunds *FrozenFunds
}

func (b *Bus) AddFrozenFund(height uint64, address types.Address, pubkey types.Pubkey, candidateID uint32, coin types.CoinID, value *big.Int) {
	b.frozenfunds.AddFund(height, address, pubkey, candidateID, coin, value)
}

func NewBus(frozenfunds *FrozenFunds) *Bus {
	return &Bus{frozenfunds: frozenfunds}
}
