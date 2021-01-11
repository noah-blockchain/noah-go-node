package dao

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
)

// Commission which is subtracted from rewards and being send to DAO Address
var (
	Address    = types.HexToAddress("NOAHxf98017d1a37cc4bec05026ef94cb46102e16638e")
	Commission = 10 // in %
)
