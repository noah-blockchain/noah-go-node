package transaction

import (
	"encoding/hex"
	"fmt"
	"github.com/noah-blockchain/noah-go-node/core/code"
	"github.com/noah-blockchain/noah-go-node/core/commissions"
	"github.com/noah-blockchain/noah-go-node/core/state"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/formula"
	"github.com/noah-blockchain/noah-go-node/hexutil"
	"github.com/tendermint/tendermint/libs/common"
	"math/big"
)

const unbondPeriod = 1555200

type UnbondData struct {
	PubKey types.Pubkey     `json:"pub_key"`
	Coin   types.CoinSymbol `json:"coin"`
	Value  *big.Int         `json:"value"`
}

func (data UnbondData) TotalSpend(tx *Transaction, context *state.StateDB) (TotalSpends, []Conversion, *big.Int, *Response) {
	panic("implement me")
}

func (data UnbondData) BasicCheck(tx *Transaction, context *state.StateDB) *Response {
	if data.PubKey == nil || data.Value == nil {
		return &Response{
			Code: code.DecodeError,
			Log:  "Incorrect tx data"}
	}

	if !context.CoinExists(data.Coin) {
		return &Response{
			Code: code.CoinNotExists,
			Log:  fmt.Sprintf("Coin %s not exists", data.Coin)}
	}

	if !context.CandidateExists(data.PubKey) {
		return &Response{
			Code: code.CandidateNotFound,
			Log:  fmt.Sprintf("Candidate with such public key not found")}
	}

	candidate := context.GetStateCandidate(data.PubKey)

	sender, _ := tx.Sender()
	stake := candidate.GetStakeOfAddress(sender, data.Coin)

	if stake == nil {
		return &Response{
			Code: code.StakeNotFound,
			Log:  fmt.Sprintf("Stake of current user not found")}
	}

	if stake.Cmp(data.Value) < 0 {
		return &Response{
			Code: code.InsufficientStake,
			Log:  fmt.Sprintf("Insufficient stake for sender account"),
			Info: EncodeError(map[string]string{
				"pub_key": data.PubKey.String(),
			})}
	}

	return nil
}

func (data UnbondData) String() string {
	return fmt.Sprintf("UNBOND pubkey:%s",
		hexutil.Encode(data.PubKey[:]))
}

func (data UnbondData) Gas() int64 {
	return commissions.UnbondTx
}

func (data UnbondData) Run(tx *Transaction, context *state.State, isCheck bool, rewardPool *big.Int, currentBlock uint64) Response {
	sender, _ := tx.Sender()

	response := data.BasicCheck(tx, context)
	if response != nil {
		return *response
	}

	commissionInBaseCoin := tx.CommissionInBaseCoin()
	commission := big.NewInt(0).Set(commissionInBaseCoin)

	if !tx.GasCoin.IsBaseCoin() {
		coin := context.Coins.GetCoin(tx.GasCoin)

		errResp := CheckReserveUnderflow(coin, commissionInBaseCoin)
		if errResp != nil {
			return *errResp
		}

		if coin.Reserve().Cmp(commissionInBaseCoin) < 0 {
			return Response{
				Code: code.CoinReserveNotSufficient,
				Log:  fmt.Sprintf("Coin reserve balance is not sufficient for transaction. Has: %s, required %s", coin.Reserve().String(), commissionInBaseCoin.String()),
				Info: EncodeError(map[string]string{
					"has_reserve": coin.Reserve().String(),
					"commission":  commissionInBaseCoin.String(),
					"gas_coin":    coin.CName,
				}),
			}
		}

		commission = formula.CalculateSaleAmount(coin.Volume(), coin.Reserve(), coin.Crr(), commissionInBaseCoin)
	}

	if context.Accounts.GetBalance(sender, tx.GasCoin).Cmp(commission) < 0 {
		return Response{
			Code: code.InsufficientFunds,
			Log:  fmt.Sprintf("Insufficient funds for sender account: %s. Wanted %s %s", sender.String(), commission, tx.GasCoin),
			Info: EncodeError(map[string]string{
				"sender":       sender.String(),
				"needed_value": commission.String(),
				"gas_coin":     fmt.Sprintf("%s", tx.GasCoin),
			}),
		}
	}

	if !isCheck {
		// now + 30 days
		unbondAtBlock := currentBlock + unbondPeriod

		rewardPool.Add(rewardPool, commissionInBaseCoin)

		context.Coins.SubReserve(tx.GasCoin, commissionInBaseCoin)
		context.Coins.SubVolume(tx.GasCoin, commission)

		context.Accounts.SubBalance(sender, tx.GasCoin, commission)
		context.Candidates.SubStake(sender, data.PubKey, data.Coin, data.Value)
		context.FrozenFunds.AddFund(unbondAtBlock, sender, data.PubKey, data.Coin, data.Value)
		context.Accounts.SetNonce(sender, tx.Nonce)
	}

	tags := kv.Pairs{
		kv.Pair{Key: []byte("tx.type"), Value: []byte(hex.EncodeToString([]byte{byte(TypeUnbond)}))},
		kv.Pair{Key: []byte("tx.from"), Value: []byte(hex.EncodeToString(sender[:]))},
	}

	return Response{
		Code:      code.OK,
		GasUsed:   tx.Gas(),
		GasWanted: tx.Gas(),
		Tags:      tags,
	}
}
