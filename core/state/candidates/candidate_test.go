package candidates

import (
	"encoding/json"
	"fmt"
	eventsdb "github.com/noah-blockchain/noah-go-node/core/events"
	"github.com/noah-blockchain/noah-go-node/core/state/accounts"
	"github.com/noah-blockchain/noah-go-node/core/state/app"
	"github.com/noah-blockchain/noah-go-node/core/state/bus"
	"github.com/noah-blockchain/noah-go-node/core/state/checker"
	"github.com/noah-blockchain/noah-go-node/core/state/coins"
	"github.com/noah-blockchain/noah-go-node/core/state/waitlist"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/helpers"
	"github.com/noah-blockchain/noah-go-node/tree"
	"github.com/tendermint/tendermint/crypto/ed25519"
	db "github.com/tendermint/tm-db"
	"math/big"
	"strconv"
	"testing"
)

func TestCandidates_Create_oneCandidate(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	candidates, err := NewCandidates(bus.NewBus(), mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	candidate := candidates.GetCandidate([32]byte{4})
	if candidate == nil {
		t.Fatal("candidate not found")
	}

	if candidates.PubKey(candidate.ID) != [32]byte{4} {
		t.Fatal("candidate error ID or PubKey")
	}
}

func TestCandidates_Commit_createThreeCandidates(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	candidates, err := NewCandidates(bus.NewBus(), mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)
	candidates.Create([20]byte{11}, [20]byte{21}, [20]byte{31}, [32]byte{41}, 10)

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	hash, version, err := mutableTree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != 1 {
		t.Fatalf("version %d", version)
	}

	if fmt.Sprintf("%X", hash) != "D7A17D41EAE39D61D3F85BC3311DA1FE306E885FF03024D0173F23E3739E719B" {
		t.Fatalf("hash %X", hash)
	}

	candidates.Create([20]byte{1, 1}, [20]byte{2, 2}, [20]byte{3, 3}, [32]byte{4, 4}, 10)

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	hash, version, err = mutableTree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != 2 {
		t.Fatalf("version %d", version)
	}

	if fmt.Sprintf("%X", hash) != "01E34A08A0CF18403B8C3708FA773A4D0B152635F321085CE7B68F04FD520A9A" {
		t.Fatalf("hash %X", hash)
	}
}

func TestCandidates_Commit_changePubKeyAndCheckBlockList(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	candidates, err := NewCandidates(bus.NewBus(), mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)
	candidates.Create([20]byte{11}, [20]byte{21}, [20]byte{31}, [32]byte{41}, 10)

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	hash, version, err := mutableTree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != 1 {
		t.Fatalf("version %d", version)
	}

	if fmt.Sprintf("%X", hash) != "D7A17D41EAE39D61D3F85BC3311DA1FE306E885FF03024D0173F23E3739E719B" {
		t.Fatalf("hash %X", hash)
	}

	candidates.ChangePubKey([32]byte{4}, [32]byte{5})
	candidates.ChangePubKey([32]byte{41}, [32]byte{6})

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	hash, version, err = mutableTree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != 2 {
		t.Fatalf("version %d", version)
	}

	if fmt.Sprintf("%X", hash) != "BB335E1AA631D9540C2CB0AC9C959B556C366B79D39B828B07106CF2DACE5A2D" {
		t.Fatalf("hash %X", hash)
	}

	if !candidates.IsBlockedPubKey([32]byte{4}) {
		t.Fatal("pub_key is not blocked")
	}

	candidates, err = NewCandidates(bus.NewBus(), mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.LoadCandidates()
	candidate := candidates.GetCandidate([32]byte{5})
	if candidate == nil {
		t.Fatal("candidate not found")
	}
	var pubkey ed25519.PubKeyEd25519
	copy(pubkey[:], types.Pubkey{5}.Bytes())
	var address types.TmAddress
	copy(address[:], pubkey.Address().Bytes())
	if *(candidate.tmAddress) != address {
		t.Fatal("tmAddress not change")
	}
	if candidates.PubKey(candidate.ID) != [32]byte{5} {
		t.Fatal("candidate map ids and pubKeys invalid")
	}

}
func TestCandidates_AddToBlockPubKey(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	candidates, err := NewCandidates(bus.NewBus(), mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.AddToBlockPubKey([32]byte{4})

	if !candidates.IsBlockedPubKey([32]byte{4}) {
		t.Fatal("pub_key is not blocked")
	}
}

func TestCandidates_Commit_withStakeAndUpdate(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	candidates, err := NewCandidates(bus.NewBus(), mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	hash, version, err := mutableTree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != 1 {
		t.Fatalf("version %d", version)
	}

	if fmt.Sprintf("%X", hash) != "FCF3853839873D3EC344016C04A5E75166F51063745670DF5D561C060E7F45A1" {
		t.Fatalf("hash %X", hash)
	}

	candidates.SetStakes([32]byte{4}, []types.Stake{
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	}, []types.Stake{
		{
			Owner:    [20]byte{2},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	})
	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	hash, version, err = mutableTree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != 2 {
		t.Fatalf("version %d", version)
	}

	if fmt.Sprintf("%X", hash) != "2D206158AA79C3BDAA019C61FEAD47BB9B6170C445EE7B36E935AC954765E99F" {
		t.Fatalf("hash %X", hash)
	}
}

func TestCandidates_Commit_edit(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	candidates, err := NewCandidates(bus.NewBus(), mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	hash, version, err := mutableTree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != 1 {
		t.Fatalf("version %d", version)
	}

	if fmt.Sprintf("%X", hash) != "FCF3853839873D3EC344016C04A5E75166F51063745670DF5D561C060E7F45A1" {
		t.Fatalf("hash %X", hash)
	}

	candidates.Edit([32]byte{4}, [20]byte{1, 1}, [20]byte{2, 2}, [20]byte{3, 3})

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	hash, version, err = mutableTree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != 2 {
		t.Fatalf("version %d", version)
	}

	if fmt.Sprintf("%X", hash) != "482BE887F2E18DC1BB829BD6AFE8887CE4EC74D4DC485DB1355D78093EAB6B35" {
		t.Fatalf("hash %X", hash)
	}

	if candidates.GetCandidateControl([32]byte{4}) != [20]byte{3, 3} {
		t.Fatal("control address is not change")
	}

	if candidates.GetCandidateOwner([32]byte{4}) != [20]byte{2, 2} {
		t.Fatal("owner address is not change")
	}

}

func TestCandidates_Commit_createOneCandidateWithID(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	candidates, err := NewCandidates(bus.NewBus(), mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.CreateWithID([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10, 1)

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	hash, version, err := mutableTree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != 1 {
		t.Fatalf("version %d", version)
	}

	if fmt.Sprintf("%X", hash) != "FCF3853839873D3EC344016C04A5E75166F51063745670DF5D561C060E7F45A1" {
		t.Fatalf("hash %X", hash)
	}

	id := candidates.ID([32]byte{4})
	if id != 1 {
		t.Fatalf("ID %d", id)
	}
}

func TestCandidates_Commit_Delegate(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	b.SetChecker(checker.NewChecker(b))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	hash, version, err := mutableTree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != 1 {
		t.Fatalf("version %d", version)
	}

	if fmt.Sprintf("%X", hash) != "FCF3853839873D3EC344016C04A5E75166F51063745670DF5D561C060E7F45A1" {
		t.Fatalf("hash %X", hash)
	}

	candidates.Delegate([20]byte{1, 1}, [32]byte{4}, 0, big.NewInt(10000000), big.NewInt(10000000))

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	hash, version, err = mutableTree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != 2 {
		t.Fatalf("version %d", version)
	}

	if fmt.Sprintf("%X", hash) != "43FE25EB54D52C6516521FB0F951E87359040A9E8DAA23BDC27C6EC5DFBC10EF" {
		t.Fatalf("hash %X", hash)
	}
}

func TestCandidates_SetOnlineAndBusSetOffline(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	b.SetChecker(checker.NewChecker(b))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	candidates.SetOnline([32]byte{4})

	candidate := candidates.GetCandidate([32]byte{4})
	if candidate == nil {
		t.Fatal("candidate not found")
	}
	if candidate.Status != CandidateStatusOnline {
		t.Fatal("candidate not change status to online")
	}
	candidates.bus.Candidates().SetOffline([32]byte{4})
	if candidate.Status != CandidateStatusOffline {
		t.Fatal("candidate not change status to offline")
	}
}

func TestCandidates_Count(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	b.SetChecker(checker.NewChecker(b))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)
	candidates.Create([20]byte{1, 1}, [20]byte{2, 2}, [20]byte{3, 3}, [32]byte{4, 4}, 20)
	candidates.Create([20]byte{1, 1, 1}, [20]byte{2, 2, 2}, [20]byte{3, 3, 3}, [32]byte{4, 4, 4}, 30)

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	hash, version, err := mutableTree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version != 1 {
		t.Fatalf("version %d", version)
	}

	if fmt.Sprintf("%X", hash) != "25F7EF5A007B3D8A5FB4DCE32F9DBC28C2AE6848B893986E3055BC3045E8F00F" {
		t.Fatalf("hash %X", hash)
	}

	count := candidates.Count()
	if count != 3 {
		t.Fatalf("coun %d", count)
	}
}

func TestCandidates_GetTotalStake_fromModelAndFromDB(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	wl, err := waitlist.NewWaitList(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}
	b.SetWaitList(waitlist.NewBus(wl))
	b.SetEvents(eventsdb.NewEventsStore(db.NewMemDB()))
	accs, err := accounts.NewAccounts(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}
	b.SetAccounts(accounts.NewBus(accs))
	b.SetChecker(checker.NewChecker(b))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)

	var stakes []types.Stake
	for i := 0; i < 1010; i++ {
		value := strconv.Itoa(i + 2000)
		stakes = append(stakes, types.Stake{
			Owner:    types.StringToAddress(strconv.Itoa(i)),
			Coin:     0,
			Value:    value,
			NoahValue: value,
		})
	}
	candidates.SetStakes([32]byte{4}, stakes, []types.Stake{
		{
			Owner:    [20]byte{2},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
		{
			Owner:    types.StringToAddress("1"),
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	})

	candidates.RecalculateStakes(0)

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	totalStake := candidates.GetTotalStake([32]byte{4})
	totalStakeString := totalStake.String()
	if totalStakeString != "2509591" {
		t.Fatalf("total stake %s", totalStakeString)
	}

	candidates, err = NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}
	candidates.LoadCandidates()
	candidates.GetCandidate([32]byte{4}).totalNoahStake = nil
	totalStake = candidates.GetTotalStake([32]byte{4})
	totalStakeString = totalStake.String()
	if totalStakeString != "2509591" {
		t.Fatalf("total stake %s", totalStakeString)
	}
}

func TestCandidates_Export(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	candidates, err := NewCandidates(bus.NewBus(), mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)
	candidates.AddToBlockPubKey([32]byte{10})
	candidates.SetStakes([32]byte{4}, []types.Stake{
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	}, []types.Stake{
		{
			Owner:    [20]byte{2},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	})
	candidates.recalculateStakes(0)
	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	state := new(types.AppState)
	candidates.Export(state)

	bytes, err := json.Marshal(state.Candidates)
	if err != nil {
		t.Fatal(err)
	}

	if string(bytes) != "[{\"id\":1,\"reward_address\":\"NOAHx0200000000000000000000000000000000000000\",\"owner_address\":\"NOAHx0100000000000000000000000000000000000000\",\"control_address\":\"NOAHx0300000000000000000000000000000000000000\",\"total_noah_stake\":\"200\",\"public_key\":\"Mp0400000000000000000000000000000000000000000000000000000000000000\",\"commission\":10,\"stakes\":[{\"owner\":\"NOAHx0100000000000000000000000000000000000000\",\"coin\":0,\"value\":\"100\",\"noah_value\":\"100\"},{\"owner\":\"NOAHx0200000000000000000000000000000000000000\",\"coin\":0,\"value\":\"100\",\"noah_value\":\"100\"}],\"updates\":[],\"status\":1}]" {
		t.Fatal("not equal JSON")
	}

	bytes, err = json.Marshal(state.BlockListCandidates)
	if err != nil {
		t.Fatal(err)
	}

	if string(bytes) != "[\"Mp0a00000000000000000000000000000000000000000000000000000000000000\"]" {
		t.Fatal("not equal JSON")
	}
}

func TestCandidates_busGetStakes(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	candidates, err := NewCandidates(bus.NewBus(), mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)
	candidates.SetStakes([32]byte{4}, []types.Stake{
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	}, []types.Stake{
		{
			Owner:    [20]byte{2},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	})

	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	stakes := candidates.bus.Candidates().GetStakes([32]byte{4})
	if len(stakes) != 1 {
		t.Fatalf("stakes count %d", len(stakes))
	}

	if stakes[0].Owner != [20]byte{1} {
		t.Fatal("owner is invalid")
	}
}

func TestCandidates_GetCandidateByTendermintAddress(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	candidates, err := NewCandidates(bus.NewBus(), mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)

	candidate := candidates.GetCandidate([32]byte{4})
	if candidate == nil {
		t.Fatal("candidate not found")
	}

	candidateByTmAddr := candidates.GetCandidateByTendermintAddress(candidate.GetTmAddress())
	if candidate.ID != candidateByTmAddr.ID {
		t.Fatal("candidate ID != candidateByTmAddr.ID")
	}
}
func TestCandidates_busGetCandidateByTendermintAddress(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	candidates, err := NewCandidates(bus.NewBus(), mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)

	candidate := candidates.GetCandidate([32]byte{4})
	if candidate == nil {
		t.Fatal("candidate not found")
	}

	candidateByTmAddr := candidates.bus.Candidates().GetCandidateByTendermintAddress(candidate.GetTmAddress())
	if candidate.ID != candidateByTmAddr.ID {
		t.Fatal("candidate ID != candidateByTmAddr.ID")
	}
}

func TestCandidates_Punish(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	wl, err := waitlist.NewWaitList(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}
	b.SetEvents(eventsdb.NewEventsStore(db.NewMemDB()))
	b.SetWaitList(waitlist.NewBus(wl))
	accs, err := accounts.NewAccounts(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}
	b.SetAccounts(accounts.NewBus(accs))
	appBus, err := app.NewApp(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}
	b.SetApp(appBus)
	b.SetChecker(checker.NewChecker(b))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	coinsState, err := coins.NewCoins(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)
	coinsState.Create(1,
		types.StrToCoinSymbol("AAA"),
		"AAACOIN",
		helpers.NoahToQNoah(big.NewInt(10)),
		10,
		helpers.NoahToQNoah(big.NewInt(10000)),
		big.NewInt(0).Exp(big.NewInt(10), big.NewInt(10+18), nil),
		nil)

	err = coinsState.Commit()
	if err != nil {
		t.Fatal(err)
	}

	symbol := coinsState.GetCoinBySymbol(types.StrToCoinSymbol("AAA"), 0)
	if symbol == nil {
		t.Fatal("coin not found")
	}

	candidates.SetStakes([32]byte{4}, []types.Stake{
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
		{
			Owner:    [20]byte{1},
			Coin:     uint64(symbol.ID()),
			Value:    "100",
			NoahValue: "0",
		},
	}, nil)

	candidates.RecalculateStakes(1)
	candidate := candidates.GetCandidate([32]byte{4})
	if candidate == nil {
		t.Fatal("candidate not found")
	}
	candidates.bus.Candidates().Punish(0, candidate.GetTmAddress())

	if candidate.stakesCount != 2 {
		t.Fatalf("stakes count %d", candidate.stakesCount)
	}

	if candidate.stakes[0].Value.String() != "99" {
		t.Fatalf("stakes[0] == %s", candidate.stakes[0].Value.String())
	}
}

type fr struct {
	unbounds []*big.Int
}

func (fr *fr) AddFrozenFund(_ uint64, _ types.Address, _ types.Pubkey, _ uint32, _ types.CoinID, value *big.Int) {
	fr.unbounds = append(fr.unbounds, value)
}
func TestCandidates_PunishByzantineCandidate(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	frozenfunds := &fr{}
	b.SetFrozenFunds(frozenfunds)
	wl, err := waitlist.NewWaitList(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}
	b.SetEvents(eventsdb.NewEventsStore(db.NewMemDB()))
	b.SetWaitList(waitlist.NewBus(wl))
	accs, err := accounts.NewAccounts(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}
	b.SetAccounts(accounts.NewBus(accs))
	appBus, err := app.NewApp(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}
	b.SetApp(appBus)
	b.SetChecker(checker.NewChecker(b))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	coinsState, err := coins.NewCoins(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)
	coinsState.Create(1,
		types.StrToCoinSymbol("AAA"),
		"AAACOIN",
		helpers.NoahToQNoah(big.NewInt(10)),
		10,
		helpers.NoahToQNoah(big.NewInt(10000)),
		big.NewInt(0).Exp(big.NewInt(10), big.NewInt(10+18), nil),
		nil)

	err = coinsState.Commit()
	if err != nil {
		t.Fatal(err)
	}

	symbol := coinsState.GetCoinBySymbol(types.StrToCoinSymbol("AAA"), 0)
	if symbol == nil {
		t.Fatal("coin not found")
	}

	candidates.SetStakes([32]byte{4}, []types.Stake{
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
		{
			Owner:    [20]byte{1},
			Coin:     uint64(symbol.ID()),
			Value:    "100",
			NoahValue: "0",
		},
	}, nil)

	candidates.RecalculateStakes(1)

	candidate := candidates.GetCandidate([32]byte{4})
	if candidate == nil {
		t.Fatal("candidate not found")
	}
	candidates.PunishByzantineCandidate(0, candidate.GetTmAddress())

	if candidates.GetStakeValueOfAddress([32]byte{4}, [20]byte{1}, symbol.ID()).String() != "0" {
		t.Error("stake[0] not unbound")
	}
	if candidates.GetStakeValueOfAddress([32]byte{4}, [20]byte{1}, 0).String() != "0" {
		t.Error("stake[1] not unbound")
	}

	if len(frozenfunds.unbounds) != 2 {
		t.Fatalf("count unbounds == %d", len(frozenfunds.unbounds))
	}

	if frozenfunds.unbounds[0].String() != "95" {
		t.Fatalf("frozenfunds.unbounds[0] == %s", frozenfunds.unbounds[0].String())
	}
	if frozenfunds.unbounds[1].String() != "95" {
		t.Fatalf("frozenfunds.unbounds[1] == %s", frozenfunds.unbounds[1].String())
	}
}

func TestCandidates_SubStake(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	b.SetChecker(checker.NewChecker(b))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)
	candidates.SetStakes([32]byte{4}, []types.Stake{
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	}, nil)
	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	candidates.SubStake([20]byte{1}, [32]byte{4}, 0, big.NewInt(10))
	stake := candidates.GetStakeOfAddress([32]byte{4}, [20]byte{1}, 0)
	if stake == nil {
		t.Fatal("stake not found")
	}

	if stake.Value.String() != "90" {
		t.Fatal("sub stake error")
	}
}

func TestCandidates_IsNewCandidateStakeSufficient(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	b.SetChecker(checker.NewChecker(b))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)
	candidates.SetStakes([32]byte{4}, []types.Stake{
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	}, nil)
	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	if !candidates.IsNewCandidateStakeSufficient(0, big.NewInt(1000), 1) {
		t.Log("is not new candidate stake sufficient")
	}
}

func TestCandidates_IsDelegatorStakeSufficient(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	wl, err := waitlist.NewWaitList(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}
	b.SetWaitList(waitlist.NewBus(wl))
	b.SetChecker(checker.NewChecker(b))
	accs, err := accounts.NewAccounts(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}
	b.SetAccounts(accounts.NewBus(accs))
	b.SetEvents(eventsdb.NewEventsStore(db.NewMemDB()))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)

	var stakes []types.Stake
	for i := 0; i < 1010; i++ {
		value := strconv.Itoa(i + 2000)
		stakes = append(stakes, types.Stake{
			Owner:    types.StringToAddress(strconv.Itoa(i)),
			Coin:     0,
			Value:    value,
			NoahValue: value,
		})
	}
	candidates.SetStakes([32]byte{4}, stakes, []types.Stake{
		{
			Owner:    [20]byte{2},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	})
	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	candidates.SetStakes([32]byte{4}, []types.Stake{
		{
			Owner:    types.StringToAddress("10000"),
			Coin:     0,
			Value:    "10000",
			NoahValue: "10000",
		},
	}, nil)

	candidates.recalculateStakes(0)
	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	if candidates.IsDelegatorStakeSufficient([20]byte{1}, [32]byte{4}, 0, big.NewInt(10)) {
		t.Fatal("is not delegator stake sufficient")
	}
}
func TestCandidates_IsDelegatorStakeSufficient_false(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	b.SetChecker(checker.NewChecker(b))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)
	candidates.SetStakes([32]byte{4}, []types.Stake{
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	}, nil)

	candidates.recalculateStakes(0)
	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	if !candidates.IsDelegatorStakeSufficient([20]byte{1}, [32]byte{4}, 0, big.NewInt(10)) {
		t.Fatal("is delegator stake sufficient")
	}
}

func TestCandidates_GetNewCandidates(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	b.SetChecker(checker.NewChecker(b))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)
	candidates.SetStakes([32]byte{4}, []types.Stake{
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "1000000000000000000000",
			NoahValue: "1000000000000000000000",
		},
	}, nil)
	candidates.SetOnline([32]byte{4})

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{5}, 10)
	candidates.SetStakes([32]byte{5}, []types.Stake{
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "1000000000000000000000",
			NoahValue: "1000000000000000000000",
		},
	}, nil)
	candidates.SetOnline([32]byte{5})

	candidates.RecalculateStakes(1)
	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	newCandidates := candidates.GetNewCandidates(2)
	if len(newCandidates) != 2 {
		t.Fatal("error count of new candidates")
	}
}

func TestCandidate_GetFilteredUpdates(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	b.SetChecker(checker.NewChecker(b))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{2}, [20]byte{3}, [32]byte{4}, 10)
	candidates.SetStakes([32]byte{4}, []types.Stake{
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	}, []types.Stake{
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
		{
			Owner:    [20]byte{1},
			Coin:     0,
			Value:    "100",
			NoahValue: "100",
		},
	})
	err = candidates.Commit()
	if err != nil {
		t.Fatal(err)
	}

	candidate := candidates.GetCandidate([32]byte{4})
	if candidate == nil {
		t.Fatal("candidate not found")
	}

	candidate.filterUpdates()

	if len(candidate.updates) != 1 {
		t.Fatal("updates not merged")
	}

	if candidate.updates[0].Value.String() != "200" {
		t.Fatal("error merge updates")
	}
}

func TestCandidates_CalculateNoahValue_RecalculateStakes_GetTotalStake(t *testing.T) {
	mutableTree, _ := tree.NewMutableTree(0, db.NewMemDB(), 1024)
	b := bus.NewBus()
	b.SetChecker(checker.NewChecker(b))
	busCoins, err := coins.NewCoins(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}
	b.SetCoins(coins.NewBus(busCoins))
	candidates, err := NewCandidates(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	coinsState, err := coins.NewCoins(b, mutableTree)
	if err != nil {
		t.Fatal(err)
	}

	candidates.Create([20]byte{1}, [20]byte{1}, [20]byte{1}, [32]byte{1}, 1)
	candidates.SetStakes([32]byte{1}, []types.Stake{
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "27331500301898443574821601",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "26788352158593847436109305",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "23056159980819190092008573",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "11588709101209768903338862",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "10699458018244407488345007",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "10178615801247206484340203",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "9695040709408605598614475",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "9311613733840163086812673",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "8035237015568850680085714",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "7751636678470495902806639",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "7729118857616059555215844",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "7246351659896715230790480",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "5634000000000000000000000",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "5111293424492290525817483",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "4636302767358508700208179",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "4375153667350433703873779",
			NoahValue: "0",
		},
		{
			Owner:    types.Address{1},
			Coin:     52,
			Value:    "6468592759016388938414535",
			NoahValue: "0",
		},
	}, nil)
	volume, _ := big.NewInt(0).SetString("235304453408778922901904166", 10)
	reserve, _ := big.NewInt(0).SetString("3417127836274022127064945", 10)
	maxSupply, _ := big.NewInt(0).SetString("1000000000000000000000000000000000", 10)
	coinsState.Create(52,
		types.StrToCoinSymbol("ONLY1"),
		"ONLY1",
		volume,
		70,
		reserve,
		maxSupply,
		nil)

	amount, _ := big.NewInt(0).SetString("407000000000000000000000", 10)
	cache := newCoinsCache()

	noahValue := candidates.calculateNoahValue(52, amount, false, true, cache)
	if noahValue.Sign() < 0 {
		t.Fatalf("%s", noahValue.String())
	}
	noahValue = candidates.calculateNoahValue(52, amount, false, true, cache)
	if noahValue.Sign() < 0 {
		t.Fatalf("%s", noahValue.String())
	}

	candidates.RecalculateStakes(0)
	totalStake := candidates.GetTotalStake([32]byte{1})
	if totalStake.String() != "2435386873327199834002556" {
		t.Fatalf("total stake %s", totalStake.String())
	}
}
