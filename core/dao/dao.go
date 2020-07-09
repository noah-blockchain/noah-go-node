package dao

import (
	"github.com/noah-blockchain/noah-go-node/core/state/accounts"
	"github.com/noah-blockchain/noah-go-node/core/types"
)

var (
	Address = (&accounts.Multisig{
		Threshold: 2,
		Weights:   []uint{1, 1, 1},
		Addresses: []types.Address{types.HexToAddress("NOAHxf98017d1a37cc4bec05026ef94cb46102e16638e"), types.HexToAddress("NOAHxf98017d1a37cc4bec05026ef94cb46102e16638e"), types.HexToAddress("NOAHxf98017d1a37cc4bec05026ef94cb46102e16638e")},
	}).Address()
	Commission = 10
)
