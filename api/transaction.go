package api

import (
	"encoding/json"

	"github.com/noah-blockchain/noah-go-node/core/transaction"
	"github.com/noah-blockchain/noah-go-node/core/transaction/encoder"
)

func Transaction(hash []byte) (json.RawMessage, error) {
	tx, err := client.Tx(hash, false)
	if err != nil {
		return nil, err
	}

	decodedTx, _ := transaction.TxDecoder.DecodeFromBytes(tx.Tx)

	cState, err := GetStateForHeight(0)
	if err != nil {
		return nil, err
	}

	cState.RLock()
	defer cState.RUnlock()

	txJsonEncoder := encoder.NewTxEncoderJSON(cState)

	return txJsonEncoder.Encode(decodedTx, tx)
}
