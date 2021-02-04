package state

import (
	"github.com/noah-blockchain/noah-go-node/core/check"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/crypto"
	"github.com/noah-blockchain/noah-go-node/helpers"
	db "github.com/tendermint/tm-db"
	"log"
	"math/big"
	"math/rand"
	"testing"
)

func TestStateExport(t *testing.T) {
	height := uint64(0)

	state, err := NewState(height, db.NewMemDB(), emptyEvents{}, 1, 1)
	if err != nil {
		log.Panic("Cannot create state")
	}

	coinTest := types.StrToCoinSymbol("TEST")
	coinTest2 := types.StrToCoinSymbol("TEST2")

	coinTestID := state.App.GetNextCoinID()
	coinTest2ID := coinTestID + 1

	state.Coins.Create(
		coinTestID,
		coinTest,
		"TEST",
		helpers.NoahToQNoah(big.NewInt(602)),
		10,
		helpers.NoahToQNoah(big.NewInt(100)),
		helpers.NoahToQNoah(big.NewInt(100)),
		nil,
	)

	state.Coins.Create(
		coinTest2ID,
		coinTest2,
		"TEST2",
		helpers.NoahToQNoah(big.NewInt(1004)),
		50,
		helpers.NoahToQNoah(big.NewInt(200)),
		helpers.NoahToQNoah(big.NewInt(200)),
		nil,
	)

	state.App.SetCoinsCount(coinTest2ID.Uint32())

	privateKey1, _ := crypto.GenerateKey()
	address1 := crypto.PubkeyToAddress(privateKey1.PublicKey)

	privateKey2, _ := crypto.GenerateKey()
	address2 := crypto.PubkeyToAddress(privateKey2.PublicKey)

	state.Accounts.AddBalance(address1, types.GetBaseCoinID(), helpers.NoahToQNoah(big.NewInt(1)))
	state.Accounts.AddBalance(address1, coinTestID, helpers.NoahToQNoah(big.NewInt(1)))
	state.Accounts.AddBalance(address2, coinTest2ID, helpers.NoahToQNoah(big.NewInt(2)))

	candidatePubKey1 := [32]byte{}
	rand.Read(candidatePubKey1[:])

	candidatePubKey2 := [32]byte{}
	rand.Read(candidatePubKey2[:])

	state.Candidates.Create(address1, address1, address1, candidatePubKey1, 10)
	state.Candidates.Create(address2, address2, address2, candidatePubKey2, 30)
	state.Validators.Create(candidatePubKey1, helpers.NoahToQNoah(big.NewInt(1)))
	state.FrozenFunds.AddFund(height, address1, candidatePubKey1, state.Candidates.ID(candidatePubKey1), coinTestID, helpers.NoahToQNoah(big.NewInt(100)))
	state.FrozenFunds.AddFund(height+10, address1, candidatePubKey1, state.Candidates.ID(candidatePubKey1), types.GetBaseCoinID(), helpers.NoahToQNoah(big.NewInt(3)))
	state.FrozenFunds.AddFund(height+100, address2, candidatePubKey1, state.Candidates.ID(candidatePubKey1), coinTestID, helpers.NoahToQNoah(big.NewInt(500)))
	state.FrozenFunds.AddFund(height+150, address2, candidatePubKey1, state.Candidates.ID(candidatePubKey1), coinTest2ID, helpers.NoahToQNoah(big.NewInt(1000)))

	newCheck := &check.Check{
		Nonce:    []byte("test nonce"),
		ChainID:  types.CurrentChainID,
		DueBlock: height + 1,
		Coin:     coinTestID,
		Value:    helpers.NoahToQNoah(big.NewInt(100)),
		GasCoin:  coinTest2ID,
	}

	err = newCheck.Sign(privateKey1)
	if err != nil {
		log.Panicf("Cannot sign check: %s", err)
	}

	state.Checks.UseCheck(newCheck)

	state.Halts.AddHaltBlock(height, types.Pubkey{0})
	state.Halts.AddHaltBlock(height+1, types.Pubkey{1})
	state.Halts.AddHaltBlock(height+2, types.Pubkey{2})

	wlAddr1 := types.StringToAddress("1")
	wlAddr2 := types.StringToAddress("2")

	state.Waitlist.AddWaitList(wlAddr1, candidatePubKey1, coinTestID, big.NewInt(1e18))
	state.Waitlist.AddWaitList(wlAddr2, candidatePubKey2, coinTest2ID, big.NewInt(2e18))

	_, err = state.Commit()
	if err != nil {
		log.Panicf("Cannot commit state: %s", err)
	}

	newState := state.Export(height)
	if err := newState.Verify(); err != nil {
		t.Error(err)
	}

	if newState.StartHeight != height {
		t.Fatalf("Wrong new state start height. Expected %d, got %d", height, newState.StartHeight)
	}

	if newState.MaxGas != state.App.GetMaxGas() {
		t.Fatalf("Wrong new state max gas. Expected %d, got %d", state.App.GetMaxGas(), newState.MaxGas)
	}

	if newState.TotalSlashed != state.App.GetTotalSlashed().String() {
		t.Fatalf("Wrong new state total slashes. Expected %d, got %s", state.App.GetMaxGas(), newState.TotalSlashed)
	}

	if len(newState.Coins) != 2 {
		t.Fatalf("Wrong new state coins size. Expected %d, got %d", 2, len(newState.Coins))
	}

	newStateCoin := newState.Coins[1]
	newStateCoin1 := newState.Coins[0]

	if newStateCoin.Name != "TEST" ||
		newStateCoin.Symbol != coinTest ||
		newStateCoin.Volume != helpers.NoahToQNoah(big.NewInt(602)).String() ||
		newStateCoin.Reserve != helpers.NoahToQNoah(big.NewInt(100)).String() ||
		newStateCoin.MaxSupply != helpers.NoahToQNoah(big.NewInt(100)).String() ||
		newStateCoin.Crr != 10 {
		t.Fatalf("Wrong new state coin data")
	}

	if newStateCoin1.Name != "TEST2" ||
		newStateCoin1.Symbol != coinTest2 ||
		newStateCoin1.Volume != helpers.NoahToQNoah(big.NewInt(1004)).String() ||
		newStateCoin1.Reserve != helpers.NoahToQNoah(big.NewInt(200)).String() ||
		newStateCoin1.MaxSupply != helpers.NoahToQNoah(big.NewInt(200)).String() ||
		newStateCoin1.Crr != 50 {
		t.Fatalf("Wrong new state coin data")
	}

	if len(newState.FrozenFunds) != 4 {
		t.Fatalf("Wrong new state frozen funds size. Expected %d, got %d", 4, len(newState.FrozenFunds))
	}

	funds := newState.FrozenFunds[0]
	funds1 := newState.FrozenFunds[1]
	funds2 := newState.FrozenFunds[2]
	funds3 := newState.FrozenFunds[3]

	if funds.Height != height ||
		funds.Address != address1 ||
		funds.Coin != uint64(coinTestID) ||
		*funds.CandidateKey != types.Pubkey(candidatePubKey1) ||
		funds.Value != helpers.NoahToQNoah(big.NewInt(100)).String() {
		t.Fatalf("Wrong new state frozen fund data")
	}

	if funds1.Height != height+10 ||
		funds1.Address != address1 ||
		funds1.Coin != uint64(types.GetBaseCoinID()) ||
		*funds1.CandidateKey != types.Pubkey(candidatePubKey1) ||
		funds1.Value != helpers.NoahToQNoah(big.NewInt(3)).String() {
		t.Fatalf("Wrong new state frozen fund data")
	}

	if funds2.Height != height+100 ||
		funds2.Address != address2 ||
		funds2.Coin != uint64(coinTestID) ||
		*funds2.CandidateKey != types.Pubkey(candidatePubKey1) ||
		funds2.Value != helpers.NoahToQNoah(big.NewInt(500)).String() {
		t.Fatalf("Wrong new state frozen fund data")
	}

	if funds3.Height != height+150 ||
		funds3.Address != address2 ||
		funds3.Coin != uint64(coinTest2ID) ||
		*funds3.CandidateKey != types.Pubkey(candidatePubKey1) ||
		funds3.Value != helpers.NoahToQNoah(big.NewInt(1000)).String() {
		t.Fatalf("Wrong new state frozen fund data")
	}

	if len(newState.UsedChecks) != 1 {
		t.Fatalf("Wrong new state used checks size. Expected %d, got %d", 1, len(newState.UsedChecks))
	}

	if string("NOAHx"+newState.UsedChecks[0]) != newCheck.Hash().String() {
		t.Fatal("Wrong new state used check data")
	}

	if len(newState.Accounts) != 2 {
		t.Fatalf("Wrong new state accounts size. Expected %d, got %d", 2, len(newState.Accounts))
	}

	var account1, account2 types.Account

	if newState.Accounts[0].Address == address1 {
		account1 = newState.Accounts[0]
		account2 = newState.Accounts[1]
	}

	if newState.Accounts[0].Address == address2 {
		account1 = newState.Accounts[1]
		account2 = newState.Accounts[0]
	}

	if account1.Address != address1 || account2.Address != address2 {
		t.Fatal("Wrong new state account addresses")
	}

	if len(account1.Balance) != 2 || len(account2.Balance) != 1 {
		t.Fatal("Wrong new state account balances size")
	}

	if account1.Balance[0].Coin != uint64(coinTestID) || account1.Balance[0].Value != helpers.NoahToQNoah(big.NewInt(1)).String() {
		t.Fatal("Wrong new state account balance data")
	}

	if account1.Balance[1].Coin != uint64(types.GetBaseCoinID()) || account1.Balance[1].Value != helpers.NoahToQNoah(big.NewInt(1)).String() {
		t.Fatal("Wrong new state account balance data")
	}

	if account2.Balance[0].Coin != uint64(coinTest2ID) || account2.Balance[0].Value != helpers.NoahToQNoah(big.NewInt(2)).String() {
		t.Fatal("Wrong new state account balance data")
	}

	if len(newState.Validators) != 1 {
		t.Fatal("Wrong new state validators size")
	}

	if newState.Validators[0].PubKey != candidatePubKey1 || newState.Validators[0].TotalNoahStake != helpers.NoahToQNoah(big.NewInt(1)).String() {
		t.Fatal("Wrong new state validator data")
	}

	if len(newState.Candidates) != 2 {
		t.Fatal("Wrong new state candidates size")
	}

	newStateCandidate1 := newState.Candidates[1]
	newStateCandidate2 := newState.Candidates[0]

	if newStateCandidate1.PubKey != candidatePubKey1 ||
		newStateCandidate1.OwnerAddress != address1 ||
		newStateCandidate1.RewardAddress != address1 ||
		newStateCandidate1.Commission != 10 {
		t.Fatal("Wrong new state candidate data")
	}

	if newStateCandidate2.PubKey != candidatePubKey2 ||
		newStateCandidate2.OwnerAddress != address2 ||
		newStateCandidate2.RewardAddress != address2 ||
		newStateCandidate2.Commission != 30 {
		t.Fatal("Wrong new state candidate data")
	}

	if len(newState.HaltBlocks) != 3 {
		t.Fatalf("Invalid amount of halts: %d. Expected 3", len(newState.HaltBlocks))
	}

	pubkey := types.Pubkey{0}
	if newState.HaltBlocks[0].Height != height || !newState.HaltBlocks[0].CandidateKey.Equals(pubkey) {
		t.Fatal("Wrong new state halt blocks")
	}

	pubkey = types.Pubkey{1}
	if newState.HaltBlocks[1].Height != height+1 || !newState.HaltBlocks[1].CandidateKey.Equals(pubkey) {
		t.Fatal("Wrong new state halt blocks")
	}

	pubkey = types.Pubkey{2}
	if newState.HaltBlocks[2].Height != height+2 || !newState.HaltBlocks[2].CandidateKey.Equals(pubkey) {
		t.Fatal("Wrong new state halt blocks")
	}

	if len(newState.Waitlist) != 2 {
		t.Fatalf("Invalid amount of waitlist: %d. Expected 2", len(newState.Waitlist))
	}

	if newState.Waitlist[0].Coin != uint64(coinTest2ID) || newState.Waitlist[0].Value != big.NewInt(2e18).String() || newState.Waitlist[0].Owner.Compare(wlAddr2) != 0 {
		t.Fatal("Invalid waitlist data")
	}

	if newState.Waitlist[1].Coin != uint64(coinTestID) || newState.Waitlist[1].Value != big.NewInt(1e18).String() || newState.Waitlist[1].Owner.Compare(wlAddr1) != 0 {
		t.Fatal("Invalid waitlist data")
	}
}
