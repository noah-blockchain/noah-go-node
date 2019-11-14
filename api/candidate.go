package api

import (
	"github.com/noah-blockchain/noah-go-node/core/state"
	"github.com/noah-blockchain/noah-go-node/core/state/candidates"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/rpc/lib/types"
)

type Stake struct {
	Owner    string `json:"owner"`
	Coin     string `json:"coin"`
	Value    string `json:"value"`
	NoahValue string `json:"noah_value"`
}

type CandidateResponse struct {
	RewardAddress string  `json:"reward_address"`
	OwnerAddress  string  `json:"owner_address"`
	TotalStake    string  `json:"total_stake"`
	PubKey        string  `json:"pub_key"`
	Commission    uint    `json:"commission"`
	Stakes        []Stake `json:"stakes,omitempty"`
	Status        byte    `json:"status"`
}

func makeResponseCandidate(state *state.State, c candidates.Candidate, includeStakes bool) CandidateResponse {
	candidate := CandidateResponse{
		RewardAddress: c.RewardAddress.String(),
		OwnerAddress:  c.OwnerAddress.String(),
		TotalStake:    state.Candidates.GetTotalStake(c.PubKey).String(),
		PubKey:        c.PubKey.String(),
		Commission:    c.Commission,
		Status:        c.Status,
	}

	if includeStakes {
		stakes := state.Candidates.GetStakes(c.PubKey)
		candidate.Stakes = make([]Stake, len(stakes))
		for i, stake := range stakes {
			candidate.Stakes[i] = Stake{
				Owner:    stake.Owner.String(),
				Coin:     stake.Coin.String(),
				Value:    stake.Value.String(),
				NoahValue: stake.NoahValue.String(),
			}
		}
	}

	return candidate
}

func Candidate(pubkey types.Pubkey, height int) (*CandidateResponse, error) {
	cState, err := GetStateForHeight(height)
	if err != nil {
		return nil, err
	}

	candidate := cState.Candidates.GetCandidate(pubkey)
	if candidate == nil {
		return nil, rpctypes.RPCError{Code: 404, Message: "Candidate not found"}
	}

	response := makeResponseCandidate(cState, *candidate, true)
	return &response, nil
}
