package transaction

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"

	"github.com/noah-blockchain/noah-go-node/core/code"
	"github.com/noah-blockchain/noah-go-node/core/state"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/formula"
	"github.com/noah-blockchain/noah-go-node/helpers"
	"github.com/tendermint/tendermint/libs/common"
)

const maxCoinNameBytes = 64
const allowedCoinSymbols = "^[A-Z0-9]{3,10}$"

var (
	minCoinSupply  = helpers.NoahToQNoah(big.NewInt(1))
	minCoinReserve = helpers.NoahToQNoah(big.NewInt(250000))
)

type CreateCoinData struct {
	Name                 string           `json:"name"`
	Symbol               types.CoinSymbol `json:"symbol"`
	InitialAmount        *big.Int         `json:"initial_amount"`
	InitialReserve       *big.Int         `json:"initial_reserve"`
	ConstantReserveRatio uint             `json:"constant_reserve_ratio"`
}

func (data CreateCoinData) TotalSpend(tx *Transaction, context *state.StateDB) (TotalSpends, []Conversion, *big.Int, *Response) {
	panic("implement me")
}

func (data CreateCoinData) BasicCheck(tx *Transaction, context *state.StateDB) *Response {
	if data.InitialReserve == nil || data.InitialAmount == nil {
		return &Response{
			Code: code.DecodeError,
			Log:  "Incorrect tx data"}
	}

	if len(data.Name) > maxCoinNameBytes {
		return &Response{
			Code: code.InvalidCoinName,
			Log:  fmt.Sprintf("Coin name is invalid. Allowed up to %d bytes.", maxCoinNameBytes)}
	}

	if match, _ := regexp.MatchString(allowedCoinSymbols, data.Symbol.String()); !match {
		return &Response{
			Code: code.InvalidCoinSymbol,
			Log:  fmt.Sprintf("Invalid coin symbol. Should be %s", allowedCoinSymbols)}
	}

	if context.Coins.Exists(data.Symbol) {
		return &Response{
			Code: code.CoinAlreadyExists,
			Log:  fmt.Sprintf("Coin already exists")}
	}

	if data.ConstantReserveRatio < 10 || data.ConstantReserveRatio > 100 {
		return &Response{
			Code: code.WrongCrr,
			Log:  fmt.Sprintf("Constant Reserve Ratio should be between 10 and 100")}
	}

	if data.InitialAmount.Cmp(minCoinSupply) == -1 || data.InitialAmount.Cmp(data.MaxSupply) == 1 {
		return &Response{
			Code: code.WrongCoinSupply,
			Log:  fmt.Sprintf("Coin supply should be between %s and %s", minCoinSupply.String(), data.MaxSupply.String())}
	}

	if data.MaxSupply.Cmp(maxCoinSupply) == 1 {
		return &Response{
			Code: code.WrongCoinSupply,
			Log:  fmt.Sprintf("Max coin supply should be less than %s", maxCoinSupply)}
	}

	if data.InitialReserve.Cmp(minCoinReserve) == -1 {
		return &Response{
			Code: code.WrongCoinSupply,
			Log:  fmt.Sprintf("Coin reserve should be greater than or equal to %s", minCoinReserve.String())}
	}

	return nil
}

func (data CreateCoinData) String() string {
	return fmt.Sprintf("CREATE COIN symbol:%s reserve:%s amount:%s crr:%d",
		data.Symbol.String(), data.InitialReserve, data.InitialAmount, data.ConstantReserveRatio)
}

func (data CreateCoinData) Gas() int64 {
	switch len(data.Symbol.String()) {
	case 3:
		return 1000000000 // 1mln bips
	case 4:
		return 100000000 // 100k bips
	case 5:
		return 10000000 // 10k bips
	case 6:
		return 1000000 // 1k bips
	}

	return 100000 // 100 bips
}

func (data CreateCoinData) Run(tx *Transaction, context *state.State, isCheck bool, rewardPool *big.Int, currentBlock uint64) Response {
	sender, _ := tx.Sender()

	response := data.BasicCheck(tx, context)
	if response != nil {
		return *response
	}

	commissionInBaseCoin := tx.CommissionInBaseCoin()
	commission := big.NewInt(0).Set(commissionInBaseCoin)

	if tx.GasCoin != types.GetBaseCoin() {
		coin := context.Coins.GetCoin(tx.GasCoin)

		errResp := CheckReserveUnderflow(coin, commissionInBaseCoin)
		if errResp != nil {
			return *errResp
		}

		if coin.Reserve().Cmp(commissionInBaseCoin) < 0 {
			return Response{
				Code: code.CoinReserveNotSufficient,
				Log:  fmt.Sprintf("Gas coin reserve balance is not sufficient for transaction. Has: %s %s, required %s %s", coin.Reserve().String(), types.GetBaseCoin(), commissionInBaseCoin.String(), types.GetBaseCoin()),
				Info: EncodeError(map[string]string{
					"has_value":      coin.Reserve().String(),
					"required_value": commissionInBaseCoin.String(),
					"gas_coin":       fmt.Sprintf("%s", types.GetBaseCoin()),
				}),
			}
		}

		commission = formula.CalculateSaleAmount(coin.Volume(), coin.Reserve(), coin.Crr(), commissionInBaseCoin)
	}

	if context.Accounts.GetBalance(sender, tx.GasCoin).Cmp(commission) < 0 {
		return Response{
			Code: code.InsufficientFunds,
			Log:  fmt.Sprintf("Insufficient funds for sender account: %s. Wanted %s %s", sender.String(), commission.String(), tx.GasCoin),
			Info: EncodeError(map[string]string{
				"sender":       sender.String(),
				"needed_value": commission.String(),
				"gas_coin":     fmt.Sprintf("%s", tx.GasCoin),
			}),
		}
	}

	if context.Accounts.GetBalance(sender, types.GetBaseCoin()).Cmp(data.InitialReserve) < 0 {
		return Response{
			Code: code.InsufficientFunds,
			Log:  fmt.Sprintf("Insufficient funds for sender account: %s. Wanted %s %s", sender.String(), data.InitialReserve.String(), types.GetBaseCoin()),
			Info: EncodeError(map[string]string{
				"sender":         sender.String(),
				"needed_reserve": data.InitialReserve.String(),
				"base_coin":      fmt.Sprintf("%s", types.GetBaseCoin()),
			}),
		}
	}

	if tx.GasCoin.IsBaseCoin() {
		totalTxCost := big.NewInt(0)
		totalTxCost.Add(totalTxCost, data.InitialReserve)
		totalTxCost.Add(totalTxCost, commission)

		if context.Accounts.GetBalance(sender, types.GetBaseCoin()).Cmp(totalTxCost) < 0 {
			return Response{
				Code: code.InsufficientFunds,
				Log:  fmt.Sprintf("Insufficient funds for sender account: %s. Wanted %s %s", sender.String(), totalTxCost.String(), tx.GasCoin),
				Info: EncodeError(map[string]string{
					"sender":       sender.String(),
					"needed_value": totalTxCost.String(),
					"gas_coin":     fmt.Sprintf("%s", tx.GasCoin),
				}),
			}
		}
	}

	if !isCheck {
		rewardPool.Add(rewardPool, commissionInBaseCoin)

		context.Coins.SubReserve(tx.GasCoin, commissionInBaseCoin)
		context.Coins.SubVolume(tx.GasCoin, commission)

		context.Accounts.SubBalance(sender, types.GetBaseCoin(), data.InitialReserve)
		context.Accounts.SubBalance(sender, tx.GasCoin, commission)
		context.Coins.Create(data.Symbol, data.Name, data.InitialAmount, data.ConstantReserveRatio, data.InitialReserve, data.MaxSupply)
		context.Accounts.AddBalance(sender, data.Symbol, data.InitialAmount)
		context.Accounts.SetNonce(sender, tx.Nonce)
	}

	tags := kv.Pairs{
		kv.Pair{Key: []byte("tx.type"), Value: []byte(hex.EncodeToString([]byte{byte(TypeCreateCoin)}))},
		kv.Pair{Key: []byte("tx.from"), Value: []byte(hex.EncodeToString(sender[:]))},
		kv.Pair{Key: []byte("tx.coin"), Value: []byte(data.Symbol.String())},
	}

	return Response{
		Code:      code.OK,
		Tags:      tags,
		GasUsed:   tx.Gas(),
		GasWanted: tx.Gas(),
	}
}
