package api

import (
	"github.com/noah-blockchain/noah-go-node/core/state"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/rpc/lib/types"
)

type Stake struct {
	Owner     types.Address    `json:"owner"`
	Coin      types.CoinSymbol `json:"coin"`
	Value     string           `json:"value"`
	NoahValue string           `json:"noah_value"`
}

type CandidateResponse struct {
	RewardAddress  types.Address `json:"reward_address"`
	OwnerAddress   types.Address `json:"owner_address"`
	TotalStake     string        `json:"total_stake"`
	PubKey         types.Pubkey  `json:"pub_key"`
	Commission     uint          `json:"commission"`
	Stakes         []Stake       `json:"stakes,omitempty"`
	CreatedAtBlock uint          `json:"created_at_block"`
	Status         byte          `json:"status"`
}

func makeResponseCandidate(c state.Candidate, includeStakes bool) CandidateResponse {
	candidate := CandidateResponse{
		RewardAddress:  c.RewardAddress,
		OwnerAddress:   c.OwnerAddress,
		TotalStake:     c.TotalNoahStake.String(),
		PubKey:         c.PubKey,
		Commission:     c.Commission,
		CreatedAtBlock: c.CreatedAtBlock,
		Status:         c.Status,
	}

	if includeStakes {
		candidate.Stakes = make([]Stake, len(c.Stakes))
		for i, stake := range c.Stakes {
			candidate.Stakes[i] = Stake{
				Owner:     stake.Owner,
				Coin:      stake.Coin,
				Value:     stake.Value.String(),
				NoahValue: stake.NoahValue.String(),
			}
		}
	}

	return candidate
}

func Candidate(pubkey []byte, height int) (*CandidateResponse, error) {
	cState, err := GetStateForHeight(height)
	if err != nil {
		return nil, err
	}

	candidate := cState.GetStateCandidate(pubkey)
	if candidate == nil {
		return nil, rpctypes.RPCError{Code: 404, Message: "Candidate not found"}
	}

	response := makeResponseCandidate(*candidate, true)
	return &response, nil
}
