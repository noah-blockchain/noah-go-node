package types

import (
	"fmt"
	"math/big"
	"strings"
)

type BigInt struct {
	big.Int
}

func (b BigInt) MarshalJSON() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b *BigInt) UnmarshalJSON(p []byte) error {
	if string(p) == "null" {
		return nil
	}

	var z big.Int
	trim := strings.Replace(string(p), "\"", "",-1)
	_, ok := z.SetString(trim, 10)
	if !ok {
		return fmt.Errorf("not a valid big integer: %s", p)
	}
	b.Int = z
	return nil
}

type AppState struct {
	Note         string       `json:"note"`
	StartHeight  uint64       `json:"start_height"`
	Validators   []Validator  `json:"validators,omitempty"`
	Candidates   []Candidate  `json:"candidates,omitempty"`
	Accounts     []Account    `json:"accounts,omitempty"`
	Coins        []Coin       `json:"coins,omitempty"`
	FrozenFunds  []FrozenFund `json:"frozen_funds,omitempty"`
	UsedChecks   []UsedCheck  `json:"used_checks,omitempty"`
	MaxGas       uint64       `json:"max_gas"`
	TotalSlashed *BigInt      `json:"total_slashed"`
}

type Validator struct {
	RewardAddress  Address   `json:"reward_address"`
	TotalNoahStake *BigInt   `json:"total_bip_stake"`
	PubKey         Pubkey    `json:"pub_key"`
	Commission     uint      `json:"commission"`
	AccumReward    *BigInt   `json:"accum_reward"`
	AbsentTimes    *BitArray `json:"absent_times"`
}

type Candidate struct {
	RewardAddress  Address `json:"reward_address"`
	OwnerAddress   Address `json:"owner_address"`
	TotalNoahStake *BigInt `json:"total_bip_stake"`
	PubKey         Pubkey  `json:"pub_key"`
	Commission     uint    `json:"commission"`
	Stakes         []Stake `json:"stakes"`
	CreatedAtBlock uint    `json:"created_at_block"`
	Status         byte    `json:"status"`
}

type Stake struct {
	Owner     Address    `json:"owner"`
	Coin      CoinSymbol `json:"coin"`
	Value     *BigInt    `json:"value"`
	NoahValue *BigInt    `json:"bip_value"`
}

type Coin struct {
	Name           string     `json:"name"`
	Symbol         CoinSymbol `json:"symbol"`
	Volume         *BigInt    `json:"volume"`
	Crr            uint       `json:"crr"`
	ReserveBalance *BigInt    `json:"reserve_balance"`
}

type FrozenFund struct {
	Height       uint64     `json:"height"`
	Address      Address    `json:"address"`
	CandidateKey Pubkey     `json:"candidate_key"`
	Coin         CoinSymbol `json:"coin"`
	Value        *BigInt    `json:"value"`
}

type UsedCheck string

type Account struct {
	Address      Address   `json:"address"`
	Balance      []Balance `json:"balance"`
	Nonce        uint64    `json:"nonce"`
	MultisigData *Multisig `json:"multisig_data,omitempty"`
}

type Balance struct {
	Coin  CoinSymbol `json:"coin"`
	Value *BigInt    `json:"value"`
}

type Multisig struct {
	Weights   []uint    `json:"weights"`
	Threshold uint      `json:"threshold"`
	Addresses []Address `json:"addresses"`
}
