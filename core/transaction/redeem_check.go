package transaction

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/noah-blockchain/noah-go-node/core/check"
	"github.com/noah-blockchain/noah-go-node/core/code"
	"github.com/noah-blockchain/noah-go-node/core/commissions"
	"github.com/noah-blockchain/noah-go-node/core/state"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/crypto"
	"github.com/noah-blockchain/noah-go-node/crypto/sha3"
	"github.com/noah-blockchain/noah-go-node/formula"
	"github.com/noah-blockchain/noah-go-node/rlp"
	"github.com/noah-blockchain/noah-go-node/upgrades"
	"github.com/tendermint/tendermint/libs/common"
	"math/big"
)

type RedeemCheckData struct {
	RawCheck []byte   `json:"raw_check"`
	Proof    [65]byte `json:"proof"`
}

func (data RedeemCheckData) TotalSpend(tx *Transaction, context *state.StateDB) (TotalSpends, []Conversion, *big.Int, *Response) {
	panic("implement me")
}

func (data RedeemCheckData) CommissionInBaseCoin(tx *Transaction) *big.Int {
	panic("implement me")
}

func (data RedeemCheckData) BasicCheck(tx *Transaction, context *state.State) *Response {
	if data.RawCheck == nil {
		return &Response{
			Code: code.DecodeError,
			Log:  "Incorrect tx data"}
	}

	// fixed potential problem with making too high commission for sender
	if tx.GasPrice != 1 {
		return &Response{
			Code: code.TooHighGasPrice,
			Log:  fmt.Sprintf("Gas price for check is limited to 1")}
	}

	return nil
}

func (data RedeemCheckData) String() string {
	return fmt.Sprintf("REDEEM CHECK proof: %x", data.Proof)
}

func (data RedeemCheckData) Gas() int64 {
	return commissions.RedeemCheckTx
}

func (data RedeemCheckData) Run(tx *Transaction, context *state.State, isCheck bool, rewardPool *big.Int, currentBlock uint64) Response {
	sender, _ := tx.Sender()

	response := data.BasicCheck(tx, context)
	if response != nil {
		return *response
	}

	decodedCheck, err := check.DecodeFromBytes(data.RawCheck)
	if err != nil {
		return Response{
			Code: code.DecodeError,
			Log:  err.Error(),
		}
	}

	if decodedCheck.ChainID != types.CurrentChainID {
		return Response{
			Code: code.WrongChainID,
			Log:  "Wrong chain id",
			Info: EncodeError(map[string]string{
				"current_chain_id": fmt.Sprintf("%d", types.CurrentChainID),
				"got_chain_id":     fmt.Sprintf("%d", decodedCheck.ChainID),
			}),
		}
	}

	if len(decodedCheck.Nonce) > 16 {
		return Response{
			Code: code.TooLongNonce,
			Log:  "Nonce is too big. Should be up to 16 bytes.",
		}
	}

	checkSender, err := decodedCheck.Sender()

	if err != nil {
		return Response{
			Code: code.DecodeError,
			Log:  err.Error()}
	}

	if !context.Coins.Exists(decodedCheck.Coin) {
		return Response{
			Code: code.CoinNotExists,
			Log:  fmt.Sprintf("Coin not exists"),
			Info: EncodeError(map[string]string{
				"coin": fmt.Sprintf("%s", decodedCheck.Coin),
			}),
		}
	}

	if !context.Coins.Exists(decodedCheck.GasCoin) {
		return Response{
			Code: code.CoinNotExists,
			Log:  fmt.Sprintf("Gas coin not exists"),
			Info: EncodeError(map[string]string{
				"gas_coin": fmt.Sprintf("%s", decodedCheck.GasCoin),
			}),
		}
	}

	if tx.GasCoin != decodedCheck.GasCoin {
		return Response{
			Code: code.WrongGasCoin,
			Log:  fmt.Sprintf("Gas coin for redeem check transaction can only be %s", decodedCheck.GasCoin),
			Info: EncodeError(map[string]string{
				"gas_coin": fmt.Sprintf("%s", decodedCheck.GasCoin),
			}),
		}
	}

	if decodedCheck.DueBlock < currentBlock {
		return Response{
			Code: code.CheckExpired,
			Log:  fmt.Sprintf("Check expired"),
			Info: EncodeError(map[string]string{
				"due_block":     fmt.Sprintf("%d", decodedCheck.DueBlock),
				"current_block": fmt.Sprintf("%d", currentBlock),
			}),
		}
	}

	if context.Checks.IsCheckUsed(decodedCheck) {
		return Response{
			Code: code.CheckUsed,
			Log:  fmt.Sprintf("Check already redeemed")}
	}

	lockPublicKey, err := decodedCheck.LockPubKey()

	if err != nil {
		return Response{
			Code: code.DecodeError,
			Log:  err.Error(),
		}
	}

	var senderAddressHash types.Hash
	hw := sha3.NewKeccak256()
	_ = rlp.Encode(hw, []interface{}{
		sender,
	})
	hw.Sum(senderAddressHash[:0])

	pub, err := crypto.Ecrecover(senderAddressHash[:], data.Proof[:])

	if err != nil {
		return Response{
			Code: code.DecodeError,
			Log:  err.Error(),
		}
	}

	if !bytes.Equal(lockPublicKey, pub) {
		return Response{
			Code: code.CheckInvalidLock,
			Log:  "Invalid proof",
		}
	}

	commissionInBaseCoin := big.NewInt(0).Mul(big.NewInt(int64(tx.GasPrice)), big.NewInt(tx.Gas()))
	commissionInBaseCoin.Mul(commissionInBaseCoin, CommissionMultiplier)
	commission := big.NewInt(0).Set(commissionInBaseCoin)

	if !decodedCheck.GasCoin.IsBaseCoin() {
		coin := context.Coins.GetCoin(decodedCheck.GasCoin)
		errResp := CheckReserveUnderflow(coin, commissionInBaseCoin)
		if errResp != nil {
			return *errResp
		}
		commission = formula.CalculateSaleAmount(coin.Volume(), coin.Reserve(), coin.Crr(), commissionInBaseCoin)
	}

	if decodedCheck.Coin == decodedCheck.GasCoin {
		totalTxCost := big.NewInt(0).Add(decodedCheck.Value, commission)
		if context.Accounts.GetBalance(checkSender, decodedCheck.Coin).Cmp(totalTxCost) < 0 {
			return Response{
				Code: code.InsufficientFunds,
				Log:  fmt.Sprintf("Insufficient funds for check issuer account: %s %s. Wanted %s %s", decodedCheck.Coin, checkSender.String(), totalTxCost.String(), decodedCheck.Coin),
				Info: EncodeError(map[string]string{
					"sender":        checkSender.String(),
					"coin":          fmt.Sprintf("%s", decodedCheck.Coin),
					"total_tx_cost": totalTxCost.String(),
				}),
			}
		}
	} else {
		if context.Accounts.GetBalance(checkSender, decodedCheck.Coin).Cmp(decodedCheck.Value) < 0 {
			return Response{
				Code: code.InsufficientFunds,
				Log:  fmt.Sprintf("Insufficient funds for check issuer account: %s %s. Wanted %s %s", checkSender.String(), decodedCheck.Coin, decodedCheck.Value.String(), decodedCheck.Coin),
				Info: EncodeError(map[string]string{
					"sender": checkSender.String(),
					"coin":   fmt.Sprintf("%s", decodedCheck.Coin),
					"value":  decodedCheck.Value.String(),
				}),
			}
		}

		if context.Accounts.GetBalance(checkSender, decodedCheck.GasCoin).Cmp(commission) < 0 {
			return Response{
				Code: code.InsufficientFunds,
				Log:  fmt.Sprintf("Insufficient funds for check issuer account: %s %s. Wanted %s %s", checkSender.String(), decodedCheck.GasCoin, commission.String(), decodedCheck.GasCoin),
				Info: EncodeError(map[string]string{
					"sender":     sender.String(),
					"gas_coin":   fmt.Sprintf("%s", decodedCheck.GasCoin),
					"commission": commission.String(),
				}),
			}
		}
	}

	if !isCheck {
		context.Checks.UseCheck(decodedCheck)
		rewardPool.Add(rewardPool, commissionInBaseCoin)

		context.Coins.SubVolume(decodedCheck.GasCoin, commission)
		context.Coins.SubReserve(decodedCheck.GasCoin, commissionInBaseCoin)

		context.Accounts.SubBalance(checkSender, decodedCheck.GasCoin, commission)
		context.Accounts.SubBalance(checkSender, decodedCheck.Coin, decodedCheck.Value)
		context.Accounts.AddBalance(sender, decodedCheck.Coin, decodedCheck.Value)
		context.Accounts.SetNonce(sender, tx.Nonce)
	}

	tags := kv.Pairs{
		kv.Pair{Key: []byte("tx.type"), Value: []byte(hex.EncodeToString([]byte{byte(TypeRedeemCheck)}))},
		kv.Pair{Key: []byte("tx.from"), Value: []byte(hex.EncodeToString(checkSender[:]))},
		kv.Pair{Key: []byte("tx.to"), Value: []byte(hex.EncodeToString(sender[:]))},
		kv.Pair{Key: []byte("tx.coin"), Value: []byte(decodedCheck.Coin.String())},
	}

	return Response{
		Code:      code.OK,
		Tags:      tags,
		GasUsed:   tx.Gas(),
		GasWanted: tx.Gas(),
	}
}
