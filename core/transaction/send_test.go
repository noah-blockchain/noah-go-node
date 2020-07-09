package transaction

import (
	"github.com/noah-blockchain/noah-go-node/core/code"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/crypto"
	"github.com/noah-blockchain/noah-go-node/helpers"
	"github.com/noah-blockchain/noah-go-node/rlp"
	"math/big"
	"sync"
	"testing"
)

func TestSendTx(t *testing.T) {
	cState := getState()

	privateKey, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	coin := types.GetBaseCoin()

	cState.Accounts.AddBalance(addr, coin, helpers.NoahToQNoah(big.NewInt(1000000)))

	value := helpers.NoahToQNoah(big.NewInt(10))
	to := types.Address([20]byte{1})

	data := SendData{
		Coin:  coin,
		To:    to,
		Value: value,
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
		Type:          TypeSend,
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

	targetBalance, _ := big.NewInt(0).SetString("999989990000000000000000", 10)
	balance := cState.Accounts.GetBalance(addr, coin)
	if balance.Cmp(targetBalance) != 0 {
		t.Fatalf("Target %s balance is not correct. Expected %s, got %s", addr.String(), targetBalance, balance)
	}

	targetTestBalance, _ := big.NewInt(0).SetString("10000000000000000000", 10)
	testBalance := cState.Accounts.GetBalance(to, coin)
	if testBalance.Cmp(targetTestBalance) != 0 {
		t.Fatalf("Target %s balance is not correct. Expected %s, got %s", to.String(), targetTestBalance, testBalance)
	}
}

func TestSendMultisigTx(t *testing.T) {
	cState := getState()

	privateKey1, _ := crypto.GenerateKey()
	addr1 := crypto.PubkeyToAddress(privateKey1.PublicKey)

	privateKey2, _ := crypto.GenerateKey()
	addr2 := crypto.PubkeyToAddress(privateKey2.PublicKey)

	coin := types.GetBaseCoin()

	msig := cState.Accounts.CreateMultisig([]uint{1, 1}, []types.Address{addr1, addr2}, 1, 1)

	cState.Accounts.AddBalance(msig, coin, helpers.NoahToQNoah(big.NewInt(1000000)))

	value := helpers.NoahToQNoah(big.NewInt(10))
	to := types.Address([20]byte{1})

	data := SendData{
		Coin:  coin,
		To:    to,
		Value: value,
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
		Type:          TypeSend,
		Data:          encodedData,
		SignatureType: SigTypeMulti,
	}

	if err := tx.Sign(privateKey1); err != nil {
		t.Fatal(err)
	}

	tx.SetMultisigAddress(msig)

	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	response := RunTx(cState, false, encodedTx, big.NewInt(0), 0, &sync.Map{}, 0)
	if response.Code != 0 {
		t.Fatalf("Response code is not 0. Error: %s", response.Log)
	}

	targetBalance, _ := big.NewInt(0).SetString("999989990000000000000000", 10)
	balance := cState.Accounts.GetBalance(msig, coin)
	if balance.Cmp(targetBalance) != 0 {
		t.Fatalf("Target %s balance is not correct. Expected %s, got %s", msig.String(), targetBalance, balance)
	}

	targetTestBalance, _ := big.NewInt(0).SetString("10000000000000000000", 10)
	testBalance := cState.Accounts.GetBalance(to, coin)
	if testBalance.Cmp(targetTestBalance) != 0 {
		t.Fatalf("Target %s balance is not correct. Expected %s, got %s", to.String(), targetTestBalance, testBalance)
	}
}

func TestSendFailedMultisigTx(t *testing.T) {
	cState := getState()

	privateKey1, _ := crypto.GenerateKey()
	addr1 := crypto.PubkeyToAddress(privateKey1.PublicKey)

	privateKey2, _ := crypto.GenerateKey()
	addr2 := crypto.PubkeyToAddress(privateKey2.PublicKey)

	coin := types.GetBaseCoin()

	msig := cState.Accounts.CreateMultisig([]uint{1, 3}, []types.Address{addr1, addr2}, 3, 1)

	cState.Accounts.AddBalance(msig, coin, helpers.NoahToQNoah(big.NewInt(1000000)))

	value := helpers.NoahToQNoah(big.NewInt(10))
	to := types.Address([20]byte{1})

	data := SendData{
		Coin:  coin,
		To:    to,
		Value: value,
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
		Type:          TypeSend,
		Data:          encodedData,
		SignatureType: SigTypeMulti,
	}

	if err := tx.Sign(privateKey1); err != nil {
		t.Fatal(err)
	}

	tx.SetMultisigAddress(msig)

	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	response := RunTx(cState, false, encodedTx, big.NewInt(0), 0, &sync.Map{}, 0)
	if response.Code != code.IncorrectMultiSignature {
		t.Fatalf("Response code is not %d. Gor: %d", code.IncorrectMultiSignature, response.Code)
	}

	targetBalance, _ := big.NewInt(0).SetString("1000000000000000000000000", 10)
	balance := cState.Accounts.GetBalance(msig, coin)
	if balance.Cmp(targetBalance) != 0 {
		t.Fatalf("Target %s balance is not correct. Expected %s, got %s", msig.String(), targetBalance, balance)
	}

	targetTestBalance, _ := big.NewInt(0).SetString("0", 10)
	testBalance := cState.Accounts.GetBalance(to, coin)
	if testBalance.Cmp(targetTestBalance) != 0 {
		t.Fatalf("Target %s balance is not correct. Expected %s, got %s", to.String(), targetTestBalance, testBalance)
	}
}
