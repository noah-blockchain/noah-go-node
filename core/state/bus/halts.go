package bus

import "github.com/noah-blockchain/noah-go-node/core/types"

type HaltBlocks interface {
	AddHaltBlock(uint64, types.Pubkey)
}
