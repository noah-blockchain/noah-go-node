package store

import (
	"github.com/noah-blockchain/noah-go-node/types"
	"github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

func init() {
	types.RegisterBlockAmino(cdc)
}
