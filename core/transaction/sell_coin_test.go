package transaction

import (
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/crypto"
	"github.com/noah-blockchain/noah-go-node/helpers"
	"github.com/noah-blockchain/noah-go-node/rlp"
	"math/big"
	"sync"
	"testing"
)

func TestSellCoinTx(t *testing.T) {
	cState := getState()

	createTestCoin(cState)

	privateKey, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	coin := types.GetBaseCoin()

	cState.AddBalance(addr, coin, helpers.NoahToQNoah(big.NewInt(1000000)))

	minValToBuy, _ := big.NewInt(0).SetString("957658277688702625", 10)
	data := SellCoinData{
		CoinToSell:        coin,
		ValueToSell:       helpers.NoahToQNoah(big.NewInt(10)),
		CoinToBuy:         getTestCoinSymbol(),
		MinimumValueToBuy: minValToBuy,
	}

	encodedData, err := rlp.EncodeToBytes(data)

	if err != nil {
		t.Fatal(err)
	}

	tx := Transaction{
		Nonce:         1,
		GasPrice:      1,
		ChainID:       types.CurrentChainID,
		GasCoin:       coin,
		Type:          TypeSellCoin,
		Data:          encodedData,
		SignatureType: SigTypeSingle,
	}

	if err := tx.Sign(privateKey); err != nil {
		t.Fatal(err)
	}

	encodedTx, err := rlp.EncodeToBytes(tx)

	if err != nil {
		t.Fatal(err)
	}

	response := RunTx(cState, false, encodedTx, big.NewInt(0), 0, &sync.Map{}, 0)

	if response.Code != 0 {
		t.Fatalf("Response code is not 0. Error: %s", response.Log)
	}

	targetBalance, _ := big.NewInt(0).SetString("999989900000000000000000", 10)
	balance := cState.Accounts.GetBalance(addr, coin)
	if balance.Cmp(targetBalance) != 0 {
		t.Fatalf("Target %s balance is not correct. Expected %s, got %s", coin, targetBalance, balance)
	}

	targetTestBalance, _ := big.NewInt(0).SetString("999955002849793446", 10)
	testBalance := cState.Accounts.GetBalance(addr, getTestCoinSymbol())
	if testBalance.Cmp(targetTestBalance) != 0 {
		t.Fatalf("Target %s balance is not correct. Expected %s, got %s", getTestCoinSymbol(), targetTestBalance, testBalance)
	}
}

func TestSellCoinTxBaseToCustomBaseCommission(t *testing.T) {
	// sell_coin: MNT
	// buy_coin: TEST
	// gas_coin: MNT

	coinToSell := types.GetBaseCoin()
	coinToBuy := types.StrToCoinSymbol("TEST")
	gasCoin := types.GetBaseCoin()
	initialBalance := helpers.BipToPip(big.NewInt(10000000))
	toSell := helpers.BipToPip(big.NewInt(100))

	cState := getState()
	initialVolume, initialReserve, crr := createTestCoinWithSymbol(cState, coinToBuy)

	privateKey, addr := getAccount()
	cState.Accounts.AddBalance(addr, coinToSell, initialBalance)

	tx := createSellCoinTx(coinToSell, coinToBuy, gasCoin, toSell, 1)
	if err := tx.Sign(privateKey); err != nil {
		t.Fatal(err)
	}

	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	// check response
	response := RunTx(cState, false, encodedTx, big.NewInt(0), 0, &sync.Map{}, 0)
	if response.Code != code.OK {
		t.Fatalf("Response code is not 0. Error %s", response.Log)
	}

	// check received coins
	buyCoinBalance := cState.Accounts.GetBalance(addr, coinToBuy)
	estimatedBuyBalance := formula.CalculatePurchaseReturn(initialVolume, initialReserve, crr, toSell)
	if buyCoinBalance.Cmp(estimatedBuyBalance) != 0 {
		t.Fatalf("Buy coin balance is not correct. Expected %s, got %s", estimatedBuyBalance.String(), buyCoinBalance.String())
	}

	// check sold coins + commission
	sellCoinBalance := cState.Accounts.GetBalance(addr, coinToSell)
	estimatedSellCoinBalance := big.NewInt(0).Set(initialBalance)
	estimatedSellCoinBalance.Sub(estimatedSellCoinBalance, tx.CommissionInBaseCoin())
	estimatedSellCoinBalance.Sub(estimatedSellCoinBalance, toSell)
	if sellCoinBalance.Cmp(estimatedSellCoinBalance) != 0 {
		t.Fatalf("Sell coin balance is not correct")
	}

	// check reserve and supply
	coinData := cState.Coins.GetCoin(coinToBuy)

	estimatedReserve := big.NewInt(0).Set(initialReserve)
	estimatedReserve.Add(estimatedReserve, toSell)
	if coinData.Reserve().Cmp(estimatedReserve) != 0 {
		t.Fatalf("Wrong coin reserve")
	}

	estimatedSupply := big.NewInt(0).Set(initialVolume)
	estimatedSupply.Add(estimatedSupply, formula.CalculatePurchaseReturn(initialVolume, initialReserve, crr, toSell))
	if coinData.Volume().Cmp(estimatedSupply) != 0 {
		t.Fatalf("Wrong coin supply")
	}
}

func TestSellCoinTxCustomToBaseBaseCommission(t *testing.T) {
	// sell_coin: TEST
	// buy_coin: MNT
	// gas_coin: MNT

	coinToSell := types.StrToCoinSymbol("TEST")
	coinToBuy := types.GetBaseCoin()
	gasCoin := types.GetBaseCoin()
	initialBalance := helpers.BipToPip(big.NewInt(10000000))
	initialGasBalance := helpers.BipToPip(big.NewInt(1))
	toSell := helpers.BipToPip(big.NewInt(100))

	cState := getState()
	initialVolume, initialReserve, crr := createTestCoinWithSymbol(cState, coinToSell)

	privateKey, addr := getAccount()
	cState.Accounts.AddBalance(addr, coinToSell, initialBalance)
	cState.Accounts.AddBalance(addr, gasCoin, initialGasBalance)

	tx := createSellCoinTx(coinToSell, coinToBuy, gasCoin, toSell, 1)
	if err := tx.Sign(privateKey); err != nil {
		t.Fatal(err)
	}

	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	// check response
	response := RunTx(cState, false, encodedTx, big.NewInt(0), 0, &sync.Map{}, 0)
	if response.Code != code.OK {
		t.Fatalf("Response code is not 0. Error %s", response.Log)
	}

	// check received coins + commission
	buyCoinBalance := cState.Accounts.GetBalance(addr, coinToBuy)
	estimatedBuyBalance := formula.CalculateSaleReturn(initialVolume, initialReserve, crr, toSell)
	estimatedBuyBalance.Add(estimatedBuyBalance, initialGasBalance)
	estimatedBuyBalance.Sub(estimatedBuyBalance, tx.CommissionInBaseCoin())
	if buyCoinBalance.Cmp(estimatedBuyBalance) != 0 {
		t.Fatalf("Buy coin balance is not correct. Expected %s, got %s", estimatedBuyBalance.String(), buyCoinBalance.String())
	}

	// check sold coins
	sellCoinBalance := cState.Accounts.GetBalance(addr, coinToSell)
	estimatedSellCoinBalance := big.NewInt(0).Set(initialBalance)
	estimatedSellCoinBalance.Sub(estimatedSellCoinBalance, toSell)
	if sellCoinBalance.Cmp(estimatedSellCoinBalance) != 0 {
		t.Fatalf("Sell coin balance is not correct")
	}

	// check reserve and supply
	coinData := cState.Coins.GetCoin(coinToSell)

	estimatedReserve := big.NewInt(0).Set(initialReserve)
	estimatedReserve.Sub(estimatedReserve, formula.CalculateSaleReturn(initialVolume, initialReserve, crr, toSell))
	if coinData.Reserve().Cmp(estimatedReserve) != 0 {
		t.Fatalf("Wrong coin reserve")
	}

	estimatedSupply := big.NewInt(0).Set(initialVolume)
	estimatedSupply.Sub(estimatedSupply, toSell)
	if coinData.Volume().Cmp(estimatedSupply) != 0 {
		t.Fatalf("Wrong coin supply")
	}
}

func TestSellCoinTxCustomToCustomBaseCommission(t *testing.T) {
	// sell_coin: TEST1
	// buy_coin: TEST2
	// gas_coin: MNT

	coinToSell := types.StrToCoinSymbol("TEST1")
	coinToBuy := types.StrToCoinSymbol("TEST2")
	gasCoin := types.GetBaseCoin()
	initialBalance := helpers.BipToPip(big.NewInt(10000000))
	initialGasBalance := helpers.BipToPip(big.NewInt(1))
	toSell := helpers.BipToPip(big.NewInt(100))

	cState := getState()
	initialVolume1, initialReserve1, crr1 := createTestCoinWithSymbol(cState, coinToSell)
	initialVolume2, initialReserve2, crr2 := createTestCoinWithSymbol(cState, coinToBuy)

	privateKey, addr := getAccount()
	cState.Accounts.AddBalance(addr, coinToSell, initialBalance)
	cState.Accounts.AddBalance(addr, gasCoin, initialGasBalance)

	tx := createSellCoinTx(coinToSell, coinToBuy, gasCoin, toSell, 1)
	if err := tx.Sign(privateKey); err != nil {
		t.Fatal(err)
	}

	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	// check response
	response := RunTx(cState, false, encodedTx, big.NewInt(0), 0, &sync.Map{}, 0)
	if response.Code != code.OK {
		t.Fatalf("Response code is not 0. Error %s", response.Log)
	}

	// check received coins
	buyCoinBalance := cState.Accounts.GetBalance(addr, coinToBuy)
	estimatedBuyBalance := formula.CalculatePurchaseReturn(initialVolume2, initialReserve2, crr2, formula.CalculateSaleReturn(initialVolume1, initialReserve1, crr1, toSell))
	if buyCoinBalance.Cmp(estimatedBuyBalance) != 0 {
		t.Fatalf("Buy coin balance is not correct. Expected %s, got %s", estimatedBuyBalance.String(), buyCoinBalance.String())
	}

	// check sold coins
	sellCoinBalance := cState.Accounts.GetBalance(addr, coinToSell)
	estimatedSellCoinBalance := big.NewInt(0).Set(initialBalance)
	estimatedSellCoinBalance.Sub(estimatedSellCoinBalance, toSell)
	if sellCoinBalance.Cmp(estimatedSellCoinBalance) != 0 {
		t.Fatalf("Sell coin balance is not correct")
	}

	// check reserve and supply
	{
		coinData := cState.Coins.GetCoin(coinToSell)

		estimatedReserve := big.NewInt(0).Set(initialReserve1)
		estimatedReserve.Sub(estimatedReserve, formula.CalculateSaleReturn(initialVolume1, initialReserve1, crr1, toSell))
		if coinData.Reserve().Cmp(estimatedReserve) != 0 {
			t.Fatalf("Wrong coin reserve")
		}

		estimatedSupply := big.NewInt(0).Set(initialVolume1)
		estimatedSupply.Sub(estimatedSupply, toSell)
		if coinData.Volume().Cmp(estimatedSupply) != 0 {
			t.Fatalf("Wrong coin supply")
		}
	}

	{
		coinData := cState.Coins.GetCoin(coinToBuy)

		estimatedReserve := big.NewInt(0).Set(initialReserve2)
		estimatedReserve.Add(estimatedReserve, formula.CalculateSaleReturn(initialVolume1, initialReserve1, crr1, toSell))
		if coinData.Reserve().Cmp(estimatedReserve) != 0 {
			t.Fatalf("Wrong coin reserve")
		}

		estimatedSupply := big.NewInt(0).Set(initialVolume2)
		estimatedSupply.Add(estimatedSupply, formula.CalculatePurchaseReturn(initialVolume2, initialReserve2, crr2, formula.CalculateSaleReturn(initialVolume1, initialReserve1, crr1, toSell)))
		if coinData.Volume().Cmp(estimatedSupply) != 0 {
			t.Fatalf("Wrong coin supply")
		}
	}
}

func TestSellCoinTxBaseToCustomCustomCommission(t *testing.T) {
	// sell_coin: MNT
	// buy_coin: TEST
	// gas_coin: TEST

	coinToSell := types.GetBaseCoin()
	coinToBuy := types.StrToCoinSymbol("TEST")
	gasCoin := types.StrToCoinSymbol("TEST")
	initialBalance := helpers.BipToPip(big.NewInt(10000000))
	initialGasBalance := helpers.BipToPip(big.NewInt(1))
	toSell := helpers.BipToPip(big.NewInt(100))

	cState := getState()
	initialVolume, initialReserve, crr := createTestCoinWithSymbol(cState, coinToBuy)

	privateKey, addr := getAccount()
	cState.Accounts.AddBalance(addr, coinToSell, initialBalance)
	cState.Accounts.AddBalance(addr, gasCoin, initialGasBalance)

	tx := createSellCoinTx(coinToSell, coinToBuy, gasCoin, toSell, 1)
	if err := tx.Sign(privateKey); err != nil {
		t.Fatal(err)
	}

	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	// check response
	response := RunTx(cState, false, encodedTx, big.NewInt(0), 0, &sync.Map{}, 0)
	if response.Code != code.OK {
		t.Fatalf("Response code is not 0. Error %s", response.Log)
	}

	// check received coins + commission
	buyCoinBalance := cState.Accounts.GetBalance(addr, coinToBuy)
	estimatedReturn := formula.CalculatePurchaseReturn(initialVolume, initialReserve, crr, toSell)
	estimatedBuyBalance := big.NewInt(0).Set(estimatedReturn)
	estimatedBuyBalance.Add(estimatedBuyBalance, initialGasBalance)
	estimatedBuyBalance.Sub(estimatedBuyBalance, formula.CalculateSaleAmount(big.NewInt(0).Add(initialVolume, estimatedReturn), big.NewInt(0).Add(initialReserve, toSell), crr, tx.CommissionInBaseCoin()))
	if buyCoinBalance.Cmp(estimatedBuyBalance) != 0 {
		t.Fatalf("Buy coin balance is not correct. Expected %s, got %s", estimatedBuyBalance.String(), buyCoinBalance.String())
	}

	// check sold coins
	sellCoinBalance := cState.Accounts.GetBalance(addr, coinToSell)
	estimatedSellCoinBalance := big.NewInt(0).Set(initialBalance)
	estimatedSellCoinBalance.Sub(estimatedSellCoinBalance, toSell)
	if sellCoinBalance.Cmp(estimatedSellCoinBalance) != 0 {
		t.Fatalf("Sell coin balance is not correct")
	}

	// check reserve and supply
	coinData := cState.Coins.GetCoin(coinToBuy)

	estimatedReserve := big.NewInt(0).Set(initialReserve)
	estimatedReserve.Add(estimatedReserve, toSell)
	estimatedReserve.Sub(estimatedReserve, tx.CommissionInBaseCoin())
	if coinData.Reserve().Cmp(estimatedReserve) != 0 {
		t.Fatalf("Wrong coin reserve")
	}

	estimatedSupply := big.NewInt(0).Set(initialVolume)
	estimatedSupply.Add(estimatedSupply, formula.CalculatePurchaseReturn(initialVolume, initialReserve, crr, toSell))
	estimatedSupply.Sub(estimatedSupply, formula.CalculateSaleAmount(big.NewInt(0).Add(initialVolume, estimatedReturn), big.NewInt(0).Add(initialReserve, toSell), crr, tx.CommissionInBaseCoin()))
	if coinData.Volume().Cmp(estimatedSupply) != 0 {
		t.Fatalf("Wrong coin supply")
	}
}

func TestSellCoinTxCustomToBaseCustomCommission(t *testing.T) {
	// sell_coin: TEST
	// buy_coin: MNT
	// gas_coin: TEST

	coinToSell := types.StrToCoinSymbol("TEST")
	coinToBuy := types.GetBaseCoin()
	gasCoin := types.StrToCoinSymbol("TEST")
	initialBalance := helpers.BipToPip(big.NewInt(10000000))
	toSell := helpers.BipToPip(big.NewInt(100))

	cState := getState()
	initialVolume, initialReserve, crr := createTestCoinWithSymbol(cState, coinToSell)

	privateKey, addr := getAccount()
	cState.Accounts.AddBalance(addr, coinToSell, initialBalance)

	tx := createSellCoinTx(coinToSell, coinToBuy, gasCoin, toSell, 1)
	if err := tx.Sign(privateKey); err != nil {
		t.Fatal(err)
	}

	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	// check response
	response := RunTx(cState, false, encodedTx, big.NewInt(0), 0, &sync.Map{}, 0)
	if response.Code != code.OK {
		t.Fatalf("Response code is not 0. Error %s", response.Log)
	}

	// check received coins
	buyCoinBalance := cState.Accounts.GetBalance(addr, coinToBuy)
	estimatedReturn := formula.CalculateSaleReturn(initialVolume, initialReserve, crr, toSell)
	estimatedBuyBalance := big.NewInt(0).Set(estimatedReturn)
	if buyCoinBalance.Cmp(estimatedBuyBalance) != 0 {
		t.Fatalf("Buy coin balance is not correct. Expected %s, got %s", estimatedBuyBalance.String(), buyCoinBalance.String())
	}

	// check sold coins + commission
	sellCoinBalance := cState.Accounts.GetBalance(addr, coinToSell)
	estimatedSellCoinBalance := big.NewInt(0).Set(initialBalance)
	estimatedSellCoinBalance.Sub(estimatedSellCoinBalance, toSell)
	estimatedSellCoinBalance.Sub(estimatedSellCoinBalance, formula.CalculateSaleAmount(big.NewInt(0).Sub(initialVolume, toSell), big.NewInt(0).Sub(initialReserve, estimatedReturn), crr, tx.CommissionInBaseCoin()))
	if sellCoinBalance.Cmp(estimatedSellCoinBalance) != 0 {
		t.Fatalf("Sell coin balance is not correct. Expected %s, got %s", estimatedSellCoinBalance.String(), sellCoinBalance.String())
	}

	// check reserve and supply
	coinData := cState.Coins.GetCoin(coinToSell)

	estimatedReserve := big.NewInt(0).Set(initialReserve)
	estimatedReserve.Sub(estimatedReserve, estimatedReturn)
	estimatedReserve.Sub(estimatedReserve, tx.CommissionInBaseCoin())
	if coinData.Reserve().Cmp(estimatedReserve) != 0 {
		t.Fatalf("Wrong coin reserve")
	}

	estimatedSupply := big.NewInt(0).Set(initialVolume)
	estimatedSupply.Sub(estimatedSupply, toSell)
	estimatedSupply.Sub(estimatedSupply, formula.CalculateSaleAmount(big.NewInt(0).Sub(initialVolume, toSell), big.NewInt(0).Sub(initialReserve, estimatedReturn), crr, tx.CommissionInBaseCoin()))
	if coinData.Volume().Cmp(estimatedSupply) != 0 {
		t.Fatalf("Wrong coin supply")
	}
}

func TestSellCoinTxCustomToCustomCustom1Commission(t *testing.T) {
	// sell_coin: TEST1
	// buy_coin: TEST2
	// gas_coin: TEST1

	coinToSell := types.StrToCoinSymbol("TEST1")
	coinToBuy := types.StrToCoinSymbol("TEST2")
	gasCoin := types.StrToCoinSymbol("TEST1")
	initialBalance := helpers.BipToPip(big.NewInt(10000000))
	toSell := helpers.BipToPip(big.NewInt(100))

	cState := getState()
	initialVolume1, initialReserve1, crr1 := createTestCoinWithSymbol(cState, coinToSell)
	initialVolume2, initialReserve2, crr2 := createTestCoinWithSymbol(cState, coinToBuy)

	privateKey, addr := getAccount()
	cState.Accounts.AddBalance(addr, coinToSell, initialBalance)

	tx := createSellCoinTx(coinToSell, coinToBuy, gasCoin, toSell, 1)
	if err := tx.Sign(privateKey); err != nil {
		t.Fatal(err)
	}

	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	// check response
	response := RunTx(cState, false, encodedTx, big.NewInt(0), 0, &sync.Map{}, 0)
	if response.Code != code.OK {
		t.Fatalf("Response code is not 0. Error %s", response.Log)
	}

	// check received coins
	buyCoinBalance := cState.Accounts.GetBalance(addr, coinToBuy)
	bipReturn := formula.CalculateSaleReturn(initialVolume1, initialReserve1, crr1, toSell)
	estimatedBuyBalance := formula.CalculatePurchaseReturn(initialVolume2, initialReserve2, crr2, bipReturn)
	if buyCoinBalance.Cmp(estimatedBuyBalance) != 0 {
		t.Fatalf("Buy coin balance is not correct. Expected %s, got %s", estimatedBuyBalance.String(), buyCoinBalance.String())
	}

	// check sold coins
	sellCoinBalance := cState.Accounts.GetBalance(addr, coinToSell)
	estimatedSellCoinBalance := big.NewInt(0).Set(initialBalance)
	estimatedSellCoinBalance.Sub(estimatedSellCoinBalance, toSell)
	commission := formula.CalculateSaleAmount(big.NewInt(0).Sub(initialVolume1, toSell), big.NewInt(0).Sub(initialReserve1, bipReturn), crr1, tx.CommissionInBaseCoin())
	estimatedSellCoinBalance.Sub(estimatedSellCoinBalance, commission)
	if sellCoinBalance.Cmp(estimatedSellCoinBalance) != 0 {
		t.Fatalf("Sell coin balance is not correct. Expected %s, got %s", estimatedSellCoinBalance.String(), sellCoinBalance.String())
	}

	// check reserve and supply
	{
		coinData := cState.Coins.GetCoin(coinToSell)

		estimatedReserve := big.NewInt(0).Set(initialReserve1)
		estimatedReserve.Sub(estimatedReserve, formula.CalculateSaleReturn(initialVolume1, initialReserve1, crr1, toSell))
		estimatedReserve.Sub(estimatedReserve, tx.CommissionInBaseCoin())
		if coinData.Reserve().Cmp(estimatedReserve) != 0 {
			t.Fatalf("Wrong coin reserve")
		}

		estimatedSupply := big.NewInt(0).Set(initialVolume1)
		estimatedSupply.Sub(estimatedSupply, toSell)
		estimatedSupply.Sub(estimatedSupply, commission)
		if coinData.Volume().Cmp(estimatedSupply) != 0 {
			t.Fatalf("Wrong coin supply")
		}
	}

	{
		coinData := cState.Coins.GetCoin(coinToBuy)

		estimatedReserve := big.NewInt(0).Set(initialReserve2)
		estimatedReserve.Add(estimatedReserve, formula.CalculateSaleReturn(initialVolume1, initialReserve1, crr1, toSell))
		if coinData.Reserve().Cmp(estimatedReserve) != 0 {
			t.Fatalf("Wrong coin reserve")
		}

		estimatedSupply := big.NewInt(0).Set(initialVolume2)
		estimatedSupply.Add(estimatedSupply, formula.CalculatePurchaseReturn(initialVolume2, initialReserve2, crr2, formula.CalculateSaleReturn(initialVolume1, initialReserve1, crr1, toSell)))
		if coinData.Volume().Cmp(estimatedSupply) != 0 {
			t.Fatalf("Wrong coin supply")
		}
	}
}

func TestSellCoinTxCustomToCustomCustom2Commission(t *testing.T) {
	// sell_coin: TEST1
	// buy_coin: TEST2
	// gas_coin: TEST2

	coinToSell := types.StrToCoinSymbol("TEST1")
	coinToBuy := types.StrToCoinSymbol("TEST2")
	gasCoin := types.StrToCoinSymbol("TEST2")
	initialBalance := helpers.BipToPip(big.NewInt(10000000))
	initialGasBalance := helpers.BipToPip(big.NewInt(1))
	toSell := helpers.BipToPip(big.NewInt(100))

	cState := getState()
	initialVolume1, initialReserve1, crr1 := createTestCoinWithSymbol(cState, coinToSell)
	initialVolume2, initialReserve2, crr2 := createTestCoinWithSymbol(cState, coinToBuy)

	privateKey, addr := getAccount()
	cState.Accounts.AddBalance(addr, coinToSell, initialBalance)
	cState.Accounts.AddBalance(addr, gasCoin, initialGasBalance)

	tx := createSellCoinTx(coinToSell, coinToBuy, gasCoin, toSell, 1)
	if err := tx.Sign(privateKey); err != nil {
		t.Fatal(err)
	}

	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	// check response
	response := RunTx(cState, false, encodedTx, big.NewInt(0), 0, &sync.Map{}, 0)
	if response.Code != code.OK {
		t.Fatalf("Response code is not 0. Error %s", response.Log)
	}

	// check received coins
	buyCoinBalance := cState.Accounts.GetBalance(addr, coinToBuy)
	bipReturn := formula.CalculateSaleReturn(initialVolume1, initialReserve1, crr1, toSell)
	estimatedReturn := formula.CalculatePurchaseReturn(initialVolume2, initialReserve2, crr2, bipReturn)
	commission := formula.CalculateSaleAmount(big.NewInt(0).Add(initialVolume2, estimatedReturn), big.NewInt(0).Add(initialReserve2, bipReturn), crr2, tx.CommissionInBaseCoin())

	estimatedBuyBalance := big.NewInt(0).Set(estimatedReturn)
	estimatedBuyBalance.Sub(estimatedBuyBalance, commission)
	estimatedBuyBalance.Add(estimatedBuyBalance, initialGasBalance)
	if buyCoinBalance.Cmp(estimatedBuyBalance) != 0 {
		t.Fatalf("Buy coin balance is not correct. Expected %s, got %s", estimatedBuyBalance.String(), buyCoinBalance.String())
	}

	// check sold coins
	sellCoinBalance := cState.Accounts.GetBalance(addr, coinToSell)
	estimatedSellCoinBalance := big.NewInt(0).Set(initialBalance)
	estimatedSellCoinBalance.Sub(estimatedSellCoinBalance, toSell)
	if sellCoinBalance.Cmp(estimatedSellCoinBalance) != 0 {
		t.Fatalf("Sell coin balance is not correct. Expected %s, got %s", estimatedSellCoinBalance.String(), sellCoinBalance.String())
	}

	// check reserve and supply
	{
		coinData := cState.Coins.GetCoin(coinToSell)

		estimatedReserve := big.NewInt(0).Set(initialReserve1)
		estimatedReserve.Sub(estimatedReserve, formula.CalculateSaleReturn(initialVolume1, initialReserve1, crr1, toSell))
		if coinData.Reserve().Cmp(estimatedReserve) != 0 {
			t.Fatalf("Wrong coin reserve")
		}

		estimatedSupply := big.NewInt(0).Set(initialVolume1)
		estimatedSupply.Sub(estimatedSupply, toSell)
		if coinData.Volume().Cmp(estimatedSupply) != 0 {
			t.Fatalf("Wrong coin supply")
		}
	}

	{
		coinData := cState.Coins.GetCoin(coinToBuy)

		estimatedReserve := big.NewInt(0).Set(initialReserve2)
		estimatedReserve.Add(estimatedReserve, formula.CalculateSaleReturn(initialVolume1, initialReserve1, crr1, toSell))
		estimatedReserve.Sub(estimatedReserve, tx.CommissionInBaseCoin())
		if coinData.Reserve().Cmp(estimatedReserve) != 0 {
			t.Fatalf("Wrong coin reserve")
		}

		estimatedSupply := big.NewInt(0).Set(initialVolume2)
		estimatedSupply.Add(estimatedSupply, formula.CalculatePurchaseReturn(initialVolume2, initialReserve2, crr2, formula.CalculateSaleReturn(initialVolume1, initialReserve1, crr1, toSell)))
		estimatedSupply.Sub(estimatedSupply, commission)
		if coinData.Volume().Cmp(estimatedSupply) != 0 {
			t.Fatalf("Wrong coin supply")
		}
	}
}

func createSellCoinTx(sellCoin, buyCoin, gasCoin types.CoinSymbol, valueToSell *big.Int, nonce uint64) *Transaction {
	data := SellCoinData{
		CoinToSell:        sellCoin,
		ValueToSell:       valueToSell,
		CoinToBuy:         buyCoin,
		MinimumValueToBuy: big.NewInt(0),
	}

	encodedData, err := rlp.EncodeToBytes(data)

	if err != nil {
		panic(err)
	}

	return &Transaction{
		Nonce:         nonce,
		GasPrice:      1,
		ChainID:       types.CurrentChainID,
		GasCoin:       gasCoin,
		Type:          TypeSellCoin,
		Data:          encodedData,
		SignatureType: SigTypeSingle,

		decodedData: data,
	}
}
