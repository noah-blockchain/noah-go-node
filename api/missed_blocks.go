package api

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/rpc/lib/types"
)

type MissedBlocksResponse struct {
	MissedBlocks      *types.BitArray `json:"missed_blocks"`
	MissedBlocksCount int             `json:"missed_blocks_count"`
}

func MissedBlocks(pubkey types.Pubkey, height int) (*MissedBlocksResponse, error) {
	cState, err := GetStateForHeight(height)
	if err != nil {
		return nil, err
	}

	if height != 0 {
		cState.Lock()
		cState.Validators.LoadValidators()
		cState.Unlock()
	}

	cState.RLock()
	defer cState.RUnlock()

	vals := cState.Validators.GetValidators()
	if vals == nil {
		return nil, rpctypes.RPCError{Code: 404, Message: "Validators not found"}
	}

	for _, val := range vals {
		if val.PubKey == pubkey {
			return &MissedBlocksResponse{
				MissedBlocks:      val.AbsentTimes,
				MissedBlocksCount: val.CountAbsentTimes(),
			}, nil
		}
	}

	return nil, rpctypes.RPCError{Code: 404, Message: "Validator not found"}
}
