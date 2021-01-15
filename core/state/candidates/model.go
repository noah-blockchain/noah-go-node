package candidates

import (
	"encoding/binary"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"math/big"
	"sort"
)

type pubkeyID struct {
	PubKey types.Pubkey
	ID     uint32
}

// Candidate represents candidate object which is stored on disk
type Candidate struct {
	PubKey         types.Pubkey
	RewardAddress  types.Address
	OwnerAddress   types.Address
	ControlAddress types.Address
	Commission     uint32
	Status         byte
	ID             uint32

	totalNoahStake *big.Int
	stakesCount   int
	stakes        [MaxDelegatorsPerCandidate]*stake
	updates       []*stake
	tmAddress     *types.TmAddress

	isDirty           bool
	isTotalStakeDirty bool
	isUpdatesDirty    bool
	dirtyStakes       [MaxDelegatorsPerCandidate]bool
}

func (candidate *Candidate) idBytes() []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, candidate.ID)
	return bs
}

func (candidate *Candidate) setStatus(status byte) {
	candidate.isDirty = true
	candidate.Status = status
}

func (candidate *Candidate) setOwner(address types.Address) {
	candidate.isDirty = true
	candidate.OwnerAddress = address
}

func (candidate *Candidate) setReward(address types.Address) {
	candidate.isDirty = true
	candidate.RewardAddress = address
}

func (candidate *Candidate) setControl(address types.Address) {
	candidate.isDirty = true
	candidate.ControlAddress = address
}

func (candidate *Candidate) setPublicKey(pubKey types.Pubkey) {
	candidate.isDirty = true
	candidate.PubKey = pubKey
	candidate.setTmAddress()
}

func (candidate *Candidate) addUpdate(stake *stake) {
	candidate.isUpdatesDirty = true
	stake.markDirty = func(i int) {
		candidate.isUpdatesDirty = true
	}
	candidate.updates = append(candidate.updates, stake)
}

func (candidate *Candidate) clearUpdates() {
	if len(candidate.updates) != 0 {
		candidate.isUpdatesDirty = true
	}

	candidate.updates = nil
}

func (candidate *Candidate) setTotalNoahStake(totalNoahValue *big.Int) {
	if totalNoahValue.Cmp(candidate.totalNoahStake) != 0 {
		candidate.isTotalStakeDirty = true
	}

	candidate.totalNoahStake.Set(totalNoahValue)
}

// GetTmAddress returns tendermint-address of a candidate
func (candidate *Candidate) GetTmAddress() types.TmAddress {
	return *candidate.tmAddress
}

func (candidate *Candidate) setTmAddress() {
	// set tm address
	var pubkey ed25519.PubKeyEd25519
	copy(pubkey[:], candidate.PubKey[:])

	var address types.TmAddress
	copy(address[:], pubkey.Address().Bytes())

	candidate.tmAddress = &address
}

// getFilteredUpdates returns updates which is > 0 in their value + merge similar updates
func (candidate *Candidate) getFilteredUpdates() []*stake {
	var updates []*stake
	for _, update := range candidate.updates {
		// skip updates with 0 stakes
		if update.Value.Cmp(big.NewInt(0)) != 1 {
			continue
		}

		// merge updates
		merged := false
		for _, u := range updates {
			if u.Coin == update.Coin && u.Owner == update.Owner {
				u.Value = big.NewInt(0).Add(u.Value, update.Value)
				merged = true
				break
			}
		}

		if !merged {
			updates = append(updates, update)
		}
	}

	return updates
}

// filterUpdates filters candidate updates: remove 0-valued updates and merge similar ones
func (candidate *Candidate) filterUpdates() {
	if len(candidate.updates) == 0 {
		return
	}

	updates := candidate.getFilteredUpdates()

	sort.SliceStable(updates, func(i, j int) bool {
		return updates[i].NoahValue.Cmp(updates[j].NoahValue) == 1
	})

	candidate.updates = updates
	candidate.isUpdatesDirty = true
}

// GetTotalNoahStake returns total stake value of a candidate
func (candidate *Candidate) GetTotalNoahStake() *big.Int {
	return big.NewInt(0).Set(candidate.totalNoahStake)
}

func (candidate *Candidate) setStakeAtIndex(index int, stake *stake, isDirty bool) {
	stake.markDirty = func(i int) {
		candidate.dirtyStakes[i] = true
	}
	stake.index = index

	candidate.stakes[index] = stake

	if isDirty {
		stake.markDirty(index)
	}
}

type stake struct {
	Owner    types.Address
	Coin     types.CoinID
	Value    *big.Int
	NoahValue *big.Int

	index     int
	markDirty func(int)
}

func (stake *stake) addValue(value *big.Int) {
	stake.markDirty(stake.index)
	stake.Value = big.NewInt(0).Add(stake.Value, value)
}

func (stake *stake) subValue(value *big.Int) {
	stake.markDirty(stake.index)
	stake.Value = big.NewInt(0).Sub(stake.Value, value)
}

func (stake *stake) setNoahValue(value *big.Int) {
	if stake.NoahValue.Cmp(value) != 0 {
		stake.markDirty(stake.index)
	}

	stake.NoahValue = big.NewInt(0).Set(value)
}

func (stake *stake) setValue(ret *big.Int) {
	stake.markDirty(stake.index)
	stake.Value = big.NewInt(0).Set(ret)
}
