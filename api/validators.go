package api

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
)

type ValidatorResponse struct {
	Pubkey      string `json:"pub_key"`
	VotingPower int64  `json:"voting_power"`
}

type ResponseValidators []ValidatorResponse

func Validators(height uint64, page, perPage int) (*ResponseValidators, error) {
	if height == 0 {
		height = blockchain.Height()
	}

	h := int64(height)
	tmVals, err := client.Validators(&h, page, perPage)
	if err != nil {
		return nil, err
	}

	responseValidators := make(ResponseValidators, len(tmVals.Validators))
	for i, val := range tmVals.Validators {
		var pk types.Pubkey
		copy(pk[:], val.PubKey.Bytes()[5:])
		responseValidators[i] = ValidatorResponse{
			Pubkey:      pk.String(),
			VotingPower: val.VotingPower,
		}
	}

	return &responseValidators, nil
}
