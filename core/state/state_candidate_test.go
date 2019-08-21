package state

import (
	"crypto/rand"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/helpers"
	"github.com/tendermint/tendermint/libs/db"
	"math/big"
	"testing"
)

func TestStake_CalcSimulatedNoahValue(t *testing.T) {
	s, err := New(0, db.NewMemDB(), false)
	if err != nil {
		panic(err)
	}

	createTestCandidate(s)

	coin := types.StrToCoinSymbol("ABC")
	value := helpers.NoahToPip(big.NewInt(100))
	reserve := helpers.NoahToPip(big.NewInt(201))

	s.CreateCoin(coin, "COIN", value, 30, reserve)

	noahValue := (&Stake{
		Coin:      coin,
		Value:     helpers.NoahToQnoah(big.NewInt(52)),
		NoahValue: big.NewInt(0),
	}).CalcSimulatedNoahValue(s)

	target := "183595287704679693988"
	if noahValue.String() != target {
		t.Fatalf("Noah value is not equals to target. Got %s, expected %s", noahValue, target)
	}
}

func createTestCandidate(stateDB *StateDB) []byte {
	address := types.Address{}
	pubkey := make([]byte, 32)
	rand.Read(pubkey)

	stateDB.CreateCandidate(address, address, pubkey, 10, 0, types.GetBaseCoin(), helpers.NoahToPip(big.NewInt(1)))

	return pubkey
}
