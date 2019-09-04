package counter

import (
	"encoding/binary"
	"fmt"
	code2 "github.com/noah-blockchain/noah-go-node/abci/example/code"
	types2 "github.com/noah-blockchain/noah-go-node/abci/types"
)

type CounterApplication struct {
	types2.BaseApplication

	hashCount int
	txCount   int
	serial    bool
}

func NewCounterApplication(serial bool) *CounterApplication {
	return &CounterApplication{serial: serial}
}

func (app *CounterApplication) Info(req types2.RequestInfo) types2.ResponseInfo {
	return types2.ResponseInfo{Data: fmt.Sprintf("{\"hashes\":%v,\"txs\":%v}", app.hashCount, app.txCount)}
}

func (app *CounterApplication) SetOption(req types2.RequestSetOption) types2.ResponseSetOption {
	key, value := req.Key, req.Value
	if key == "serial" && value == "on" {
		app.serial = true
	} else {
		/*
			TODO Panic and have the ABCI server pass an exception.
			The client can call SetOptionSync() and get an `error`.
			return types.ResponseSetOption{
				Error: fmt.Sprintf("Unknown key (%s) or value (%s)", key, value),
			}
		*/
		return types2.ResponseSetOption{}
	}

	return types2.ResponseSetOption{}
}

func (app *CounterApplication) DeliverTx(req types2.RequestDeliverTx) types2.ResponseDeliverTx {
	if app.serial {
		if len(req.Tx) > 8 {
			return types2.ResponseDeliverTx{
				Code: code2.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Max tx size is 8 bytes, got %d", len(req.Tx))}
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(req.Tx):], req.Tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue != uint64(app.txCount) {
			return types2.ResponseDeliverTx{
				Code: code2.CodeTypeBadNonce,
				Log:  fmt.Sprintf("Invalid nonce. Expected %v, got %v", app.txCount, txValue)}
		}
	}
	app.txCount++
	return types2.ResponseDeliverTx{Code: code2.CodeTypeOK}
}

func (app *CounterApplication) CheckTx(req types2.RequestCheckTx) types2.ResponseCheckTx {
	if app.serial {
		if len(req.Tx) > 8 {
			return types2.ResponseCheckTx{
				Code: code2.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Max tx size is 8 bytes, got %d", len(req.Tx))}
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(req.Tx):], req.Tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue < uint64(app.txCount) {
			return types2.ResponseCheckTx{
				Code: code2.CodeTypeBadNonce,
				Log:  fmt.Sprintf("Invalid nonce. Expected >= %v, got %v", app.txCount, txValue)}
		}
	}
	return types2.ResponseCheckTx{Code: code2.CodeTypeOK}
}

func (app *CounterApplication) Commit() (resp types2.ResponseCommit) {
	app.hashCount++
	if app.txCount == 0 {
		return types2.ResponseCommit{}
	}
	hash := make([]byte, 8)
	binary.BigEndian.PutUint64(hash, uint64(app.txCount))
	return types2.ResponseCommit{Data: hash}
}

func (app *CounterApplication) Query(reqQuery types2.RequestQuery) types2.ResponseQuery {
	switch reqQuery.Path {
	case "hash":
		return types2.ResponseQuery{Value: []byte(fmt.Sprintf("%v", app.hashCount))}
	case "tx":
		return types2.ResponseQuery{Value: []byte(fmt.Sprintf("%v", app.txCount))}
	default:
		return types2.ResponseQuery{Log: fmt.Sprintf("Invalid query path. Expected hash or tx, got %v", reqQuery.Path)}
	}
}
