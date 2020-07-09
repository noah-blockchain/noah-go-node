package transaction

import (
	"encoding/hex"
	"fmt"
	"github.com/noah-blockchain/noah-go-node/core/code"
	"github.com/noah-blockchain/noah-go-node/core/commissions"
	"github.com/noah-blockchain/noah-go-node/core/state"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/formula"
	"github.com/tendermint/tendermint/libs/common"
	"math/big"
)

type SellAllCoinData struct {
	CoinToSell        types.CoinSymbol `json:"coin_to_sell"`
	CoinToBuy         types.CoinSymbol `json:"coin_to_buy"`
	MinimumValueToBuy *big.Int         `json:"minimum_value_to_buy"`
}

func (data SellAllCoinData) TotalSpend(tx *Transaction, context *state.StateDB) (TotalSpends, []Conversion, *big.Int, *Response) {
	sender, _ := tx.Sender()

	total := TotalSpends{}
	var conversions []Conversion

	commissionInBaseCoin := tx.CommissionInBaseCoin()
	available := context.GetBalance(sender, data.CoinToSell)
	var value *big.Int

	total.Add(data.CoinToSell, available)

	switch {
	case data.CoinToSell.IsBaseCoin():
		amountToSell := big.NewInt(0).Set(available)
		amountToSell.Sub(amountToSell, commissionInBaseCoin)

		coin := context.Coins.GetCoin(data.CoinToBuy)
		value = formula.CalculatePurchaseReturn(coin.Volume(), coin.Reserve(), coin.Crr(), amountToSell)

		if value.Cmp(data.MinimumValueToBuy) == -1 {
			return nil, nil, nil, &Response{
				Code: code.MinimumValueToBuyReached,
				Log:  fmt.Sprintf("You wanted to get minimum %s, but currently you will get %s", data.MinimumValueToBuy.String(), value.String()),
				Info: EncodeError(map[string]string{
					"minimum_value_to_buy": data.MinimumValueToBuy.String(),
					"coin":                 value.String(),
				}),
			}
		}

		if errResp := CheckForCoinSupplyOverflow(coin.Volume(), value, coin.MaxSupply()); errResp != nil {
			return nil, nil, nil, errResp
		}

		conversions = append(conversions, Conversion{
			FromCoin:  data.CoinToSell,
			ToCoin:    data.CoinToBuy,
			ToAmount:  value,
			ToReserve: amountToSell,
		})
	case data.CoinToBuy.IsBaseCoin():
		amountToSell := big.NewInt(0).Set(available)

		coin := context.Coins.GetCoin(data.CoinToSell)
		ret := formula.CalculateSaleReturn(coin.Volume(), coin.Reserve(), coin.Crr(), amountToSell)

		if ret.Cmp(data.MinimumValueToBuy) == -1 {
			return nil, nil, nil, &Response{
				Code: code.MinimumValueToBuyReached,
				Log:  fmt.Sprintf("You wanted to get minimum %s, but currently you will get %s", data.MinimumValueToBuy.String(), ret.String()),
				Info: EncodeError(map[string]string{
					"minimum_value_to_buy": data.MinimumValueToBuy.String(),
					"will_get_value":       ret.String(),
				}),
			}
		}

		if ret.Cmp(commissionInBaseCoin) == -1 {
			return nil, nil, nil, &Response{
				Code: code.InsufficientFunds,
				Log:  fmt.Sprintf("Insufficient funds for sender account"),
			}
		}

		value = big.NewInt(0).Set(ret)
		value.Sub(ret, commissionInBaseCoin)

		conversions = append(conversions, Conversion{
			FromCoin:    data.CoinToSell,
			FromAmount:  amountToSell,
			FromReserve: ret,
			ToCoin:      data.CoinToBuy,
		})
	default:
		amountToSell := big.NewInt(0).Set(available)

		coinFrom := context.Coins.GetCoin(data.CoinToSell)
		coinTo := context.Coins.GetCoin(data.CoinToBuy)

		basecoinValue := formula.CalculateSaleReturn(coinFrom.Volume(), coinFrom.Reserve(), coinFrom.Crr(), amountToSell)
		if basecoinValue.Cmp(commissionInBaseCoin) == -1 {
			return nil, nil, nil, &Response{
				Code: code.InsufficientFunds,
				Log:  fmt.Sprintf("Insufficient funds for sender account"),
			}
		}

		basecoinValue.Sub(basecoinValue, commissionInBaseCoin)

		value = formula.CalculatePurchaseReturn(coinTo.Volume(), coinTo.Reserve(), coinTo.Crr(), basecoinValue)
		if value.Cmp(data.MinimumValueToBuy) == -1 {
			return nil, nil, nil, &Response{
				Code: code.MinimumValueToBuyReached,
				Log:  fmt.Sprintf("You wanted to get minimum %s, but currently you will get %s", data.MinimumValueToBuy.String(), value.String()),
				Info: EncodeError(map[string]string{
					"minimum_value_to_buy": data.MinimumValueToBuy.String(),
					"will_get_value":       value.String(),
				}),
			}
		}

		if errResp := CheckForCoinSupplyOverflow(coinTo.Volume(), value, coinTo.MaxSupply()); errResp != nil {
			return nil, nil, nil, errResp
		}

		conversions = append(conversions, Conversion{
			FromCoin:    data.CoinToSell,
			FromAmount:  amountToSell,
			FromReserve: big.NewInt(0).Add(basecoinValue, commissionInBaseCoin),
			ToCoin:      data.CoinToBuy,
			ToAmount:    value,
			ToReserve:   basecoinValue,
		})
	}

	return total, conversions, value, nil
}

func (data SellAllCoinData) BasicCheck(tx *Transaction, context *state.State) *Response {
	if data.CoinToSell == data.CoinToBuy {
		return &Response{
			Code: code.CrossConvert,
			Log:  fmt.Sprintf("\"From\" coin equals to \"to\" coin"),
			Info: EncodeError(map[string]string{
				"coin_to_sell": fmt.Sprintf("%s", data.CoinToSell),
				"coin_to_buy":  fmt.Sprintf("%s", data.CoinToBuy),
			}),
		}
	}

	if !context.Coins.Exists(data.CoinToSell) {
		return &Response{
			Code: code.CoinNotExists,
			Log:  fmt.Sprintf("Coin to sell not exists"),
			Info: EncodeError(map[string]string{
				"coin_to_sell": fmt.Sprintf("%s", data.CoinToSell),
			}),
		}
	}

	if !context.Coins.Exists(data.CoinToBuy) {
		return &Response{
			Code: code.CoinNotExists,
			Log:  fmt.Sprintf("Coin to buy not exists"),
			Info: EncodeError(map[string]string{
				"coin_to_buy": fmt.Sprintf("%s", data.CoinToBuy),
			}),
		}
	}

	return nil
}

func (data SellAllCoinData) String() string {
	return fmt.Sprintf("SELL ALL COIN sell:%s buy:%s",
		data.CoinToSell.String(), data.CoinToBuy.String())
}

func (data SellAllCoinData) Gas() int64 {
	return commissions.ConvertTx
}

func (data SellAllCoinData) Run(tx *Transaction, context *state.State, isCheck bool, rewardPool *big.Int, currentBlock uint64) Response {
	sender, _ := tx.Sender()

	response := data.BasicCheck(tx, context)
	if response != nil {
		return *response
	}

	available := context.Accounts.GetBalance(sender, data.CoinToSell)

	totalSpends, conversions, value, response := data.TotalSpend(tx, context)
	if response != nil {
		return *response
	}

	for _, ts := range totalSpends {
		if context.Accounts.GetBalance(sender, ts.Coin).Cmp(ts.Value) < 0 {
			return Response{
				Code: code.InsufficientFunds,
				Log: fmt.Sprintf("Insufficient funds for sender account: %s. Wanted %s %s.",
					sender.String(),
					ts.Value.String(),
					ts.Coin),
				Info: EncodeError(map[string]string{
					"sender":       sender.String(),
					"needed_value": ts.Value.String(),
					"coin":         fmt.Sprintf("%s", ts.Coin),
				}),
			}
		}
	}

	errResp := checkConversionsReserveUnderflow(conversions, context)
	if errResp != nil {
		return *errResp
	}

	if !isCheck {
		for _, ts := range totalSpends {
			context.Accounts.SubBalance(sender, ts.Coin, ts.Value)
		}

		for _, conversion := range conversions {
			context.Coins.SubVolume(conversion.FromCoin, conversion.FromAmount)
			context.Coins.SubReserve(conversion.FromCoin, conversion.FromReserve)

			context.Coins.AddVolume(conversion.ToCoin, conversion.ToAmount)
			context.Coins.AddReserve(conversion.ToCoin, conversion.ToReserve)
		}

		rewardPool.Add(rewardPool, tx.CommissionInBaseCoin())
		context.Accounts.AddBalance(sender, data.CoinToBuy, value)
		context.Accounts.SetNonce(sender, tx.Nonce)
	}

	tags := kv.Pairs{
		kv.Pair{Key: []byte("tx.type"), Value: []byte(hex.EncodeToString([]byte{byte(TypeSellAllCoin)}))},
		kv.Pair{Key: []byte("tx.from"), Value: []byte(hex.EncodeToString(sender[:]))},
		kv.Pair{Key: []byte("tx.coin_to_buy"), Value: []byte(data.CoinToBuy.String())},
		kv.Pair{Key: []byte("tx.coin_to_sell"), Value: []byte(data.CoinToSell.String())},
		kv.Pair{Key: []byte("tx.return"), Value: []byte(value.String())},
		kv.Pair{Key: []byte("tx.sell_amount"), Value: []byte(available.String())},
	}

	return Response{
		Code:      code.OK,
		Tags:      tags,
		GasUsed:   tx.Gas(),
		GasWanted: tx.Gas(),
	}
}
