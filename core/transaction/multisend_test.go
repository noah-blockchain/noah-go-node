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

func TestMultisendTx(t *testing.T) {
	cState := getState()

	privateKey, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	coin := types.GetBaseCoin()

	cState.AddBalance(addr, coin, helpers.NoahToQNoah(big.NewInt(1000000)))

	value := helpers.NoahToQNoah(big.NewInt(10))
	to := types.Address([20]byte{1})

	data := MultisendData{
		List: []MultisendDataItem{
			{
				Coin:  coin,
				To:    to,
				Value: value,
			},
		},
	}

	encodedData, err := rlp.EncodeToBytes(data)

	if err != nil {
		t.Fatal(err)
	}

	tx := Transaction{
		Nonce:         1,
		GasPrice:      1,
		GasCoin:       coin,
		ChainID:       types.GetCurrentChainID(),
		Type:          TypeMultisend,
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

	response := RunTx(cState, false, encodedTx, big.NewInt(0), 0, sync.Map{}, 0)

	if response.Code != 0 {
		t.Fatalf("Response code is not 0. Error: %s", response.Log)
	}

	targetBalance, _ := big.NewInt(0).SetString("999989990000000000000000", 10)
	balance := cState.GetBalance(addr, coin)
	if balance.Cmp(targetBalance) != 0 {
		t.Fatalf("Target %s balance is not correct. Expected %s, got %s", addr.String(), targetBalance, balance)
	}

	targetTestBalance, _ := big.NewInt(0).SetString("10000000000000000000", 10)
	testBalance := cState.GetBalance(to, coin)
	if testBalance.Cmp(targetTestBalance) != 0 {
		t.Fatalf("Target %s balance is not correct. Expected %s, got %s", to.String(), targetTestBalance, testBalance)
	}
}
