package abcicli

import (
	types2 "github.com/noah-blockchain/noah-go-node/abci/types"
	"sync"

	cmn "github.com/tendermint/tendermint/libs/common"
)

var _ Client = (*localClient)(nil)

// NOTE: use defer to unlock mutex because Application might panic (e.g., in
// case of malicious tx or query). It only makes sense for publicly exposed
// methods like CheckTx (/broadcast_tx_* RPC endpoint) or Query (/abci_query
// RPC endpoint), but defers are used everywhere for the sake of consistency.
type localClient struct {
	cmn.BaseService

	mtx *sync.Mutex
	types2.Application
	Callback
}

func NewLocalClient(mtx *sync.Mutex, app types2.Application) *localClient {
	if mtx == nil {
		mtx = new(sync.Mutex)
	}
	cli := &localClient{
		mtx:         mtx,
		Application: app,
	}
	cli.BaseService = *cmn.NewBaseService(nil, "localClient", cli)
	return cli
}

func (app *localClient) SetResponseCallback(cb Callback) {
	app.mtx.Lock()
	app.Callback = cb
	app.mtx.Unlock()
}

// TODO: change types.Application to include Error()?
func (app *localClient) Error() error {
	return nil
}

func (app *localClient) FlushAsync() *ReqRes {
	// Do nothing
	return newLocalReqRes(types2.ToRequestFlush(), nil)
}

func (app *localClient) EchoAsync(msg string) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.callback(
		types2.ToRequestEcho(msg),
		types2.ToResponseEcho(msg),
	)
}

func (app *localClient) InfoAsync(req types2.RequestInfo) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Info(req)
	return app.callback(
		types2.ToRequestInfo(req),
		types2.ToResponseInfo(res),
	)
}

func (app *localClient) SetOptionAsync(req types2.RequestSetOption) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.SetOption(req)
	return app.callback(
		types2.ToRequestSetOption(req),
		types2.ToResponseSetOption(res),
	)
}

func (app *localClient) DeliverTxAsync(params types2.RequestDeliverTx) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.DeliverTx(params)
	return app.callback(
		types2.ToRequestDeliverTx(params),
		types2.ToResponseDeliverTx(res),
	)
}

func (app *localClient) CheckTxAsync(req types2.RequestCheckTx) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.CheckTx(req)
	return app.callback(
		types2.ToRequestCheckTx(req),
		types2.ToResponseCheckTx(res),
	)
}

func (app *localClient) QueryAsync(req types2.RequestQuery) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Query(req)
	return app.callback(
		types2.ToRequestQuery(req),
		types2.ToResponseQuery(res),
	)
}

func (app *localClient) CommitAsync() *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Commit()
	return app.callback(
		types2.ToRequestCommit(),
		types2.ToResponseCommit(res),
	)
}

func (app *localClient) InitChainAsync(req types2.RequestInitChain) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.InitChain(req)
	return app.callback(
		types2.ToRequestInitChain(req),
		types2.ToResponseInitChain(res),
	)
}

func (app *localClient) BeginBlockAsync(req types2.RequestBeginBlock) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.BeginBlock(req)
	return app.callback(
		types2.ToRequestBeginBlock(req),
		types2.ToResponseBeginBlock(res),
	)
}

func (app *localClient) EndBlockAsync(req types2.RequestEndBlock) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.EndBlock(req)
	return app.callback(
		types2.ToRequestEndBlock(req),
		types2.ToResponseEndBlock(res),
	)
}

//-------------------------------------------------------

func (app *localClient) FlushSync() error {
	return nil
}

func (app *localClient) EchoSync(msg string) (*types2.ResponseEcho, error) {
	return &types2.ResponseEcho{Message: msg}, nil
}

func (app *localClient) InfoSync(req types2.RequestInfo) (*types2.ResponseInfo, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Info(req)
	return &res, nil
}

func (app *localClient) SetOptionSync(req types2.RequestSetOption) (*types2.ResponseSetOption, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.SetOption(req)
	return &res, nil
}

func (app *localClient) DeliverTxSync(req types2.RequestDeliverTx) (*types2.ResponseDeliverTx, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.DeliverTx(req)
	return &res, nil
}

func (app *localClient) CheckTxSync(req types2.RequestCheckTx) (*types2.ResponseCheckTx, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.CheckTx(req)
	return &res, nil
}

func (app *localClient) QuerySync(req types2.RequestQuery) (*types2.ResponseQuery, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Query(req)
	return &res, nil
}

func (app *localClient) CommitSync() (*types2.ResponseCommit, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Commit()
	return &res, nil
}

func (app *localClient) InitChainSync(req types2.RequestInitChain) (*types2.ResponseInitChain, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.InitChain(req)
	return &res, nil
}

func (app *localClient) BeginBlockSync(req types2.RequestBeginBlock) (*types2.ResponseBeginBlock, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.BeginBlock(req)
	return &res, nil
}

func (app *localClient) EndBlockSync(req types2.RequestEndBlock) (*types2.ResponseEndBlock, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.EndBlock(req)
	return &res, nil
}

//-------------------------------------------------------

func (app *localClient) callback(req *types2.Request, res *types2.Response) *ReqRes {
	app.Callback(req, res)
	return newLocalReqRes(req, res)
}

func newLocalReqRes(req *types2.Request, res *types2.Response) *ReqRes {
	reqRes := NewReqRes(req)
	Response = res
	reqRes.SetDone()
	return reqRes
}
