package kvstore

import (
	types2 "github.com/noah-blockchain/noah-go-node/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// RandVal creates one random validator, with a key derived
// from the input value
func RandVal(i int) types2.ValidatorUpdate {
	pubkey := cmn.RandBytes(32)
	power := cmn.RandUint16() + 1
	v := types2.Ed25519ValidatorUpdate(pubkey, int64(power))
	return v
}

// RandVals returns a list of cnt validators for initializing
// the application. Note that the keys are deterministically
// derived from the index in the array, while the power is
// random (Change this if not desired)
func RandVals(cnt int) []types2.ValidatorUpdate {
	res := make([]types2.ValidatorUpdate, cnt)
	for i := 0; i < cnt; i++ {
		res[i] = RandVal(i)
	}
	return res
}

// InitKVStore initializes the kvstore app with some data,
// which allows tests to pass and is fine as long as you
// don't make any tx that modify the validator state
func InitKVStore(app *PersistentKVStoreApplication) {
	app.InitChain(types2.RequestInitChain{
		Validators: RandVals(1),
	})
}
