package abcicli

import (
	"fmt"
	types2 "github.com/noah-blockchain/noah-go-node/abci/types"
	"sync"

	cmn "github.com/tendermint/tendermint/libs/common"
)

const (
	dialRetryIntervalSeconds = 3
	echoRetryIntervalSeconds = 1
)

// Client defines an interface for an ABCI client.
// All `Async` methods return a `ReqRes` object.
// All `Sync` methods return the appropriate protobuf ResponseXxx struct and an error.
// Note these are client errors, eg. ABCI socket connectivity issues.
// Application-related errors are reflected in response via ABCI error codes and logs.
type Client interface {
	cmn.Service

	SetResponseCallback(Callback)
	Error() error

	FlushAsync() *ReqRes
	EchoAsync(msg string) *ReqRes
	InfoAsync(types2.RequestInfo) *ReqRes
	SetOptionAsync(types2.RequestSetOption) *ReqRes
	DeliverTxAsync(types2.RequestDeliverTx) *ReqRes
	CheckTxAsync(types2.RequestCheckTx) *ReqRes
	QueryAsync(types2.RequestQuery) *ReqRes
	CommitAsync() *ReqRes
	InitChainAsync(types2.RequestInitChain) *ReqRes
	BeginBlockAsync(types2.RequestBeginBlock) *ReqRes
	EndBlockAsync(types2.RequestEndBlock) *ReqRes

	FlushSync() error
	EchoSync(msg string) (*types2.ResponseEcho, error)
	InfoSync(types2.RequestInfo) (*types2.ResponseInfo, error)
	SetOptionSync(types2.RequestSetOption) (*types2.ResponseSetOption, error)
	DeliverTxSync(types2.RequestDeliverTx) (*types2.ResponseDeliverTx, error)
	CheckTxSync(types2.RequestCheckTx) (*types2.ResponseCheckTx, error)
	QuerySync(types2.RequestQuery) (*types2.ResponseQuery, error)
	CommitSync() (*types2.ResponseCommit, error)
	InitChainSync(types2.RequestInitChain) (*types2.ResponseInitChain, error)
	BeginBlockSync(types2.RequestBeginBlock) (*types2.ResponseBeginBlock, error)
	EndBlockSync(types2.RequestEndBlock) (*types2.ResponseEndBlock, error)
}

//----------------------------------------

// NewClient returns a new ABCI client of the specified transport type.
// It returns an error if the transport is not "socket" or "grpc"
func NewClient(addr, transport string, mustConnect bool) (client Client, err error) {
	switch transport {
	case "socket":
		client = NewSocketClient(addr, mustConnect)
	case "grpc":
		client = NewGRPCClient(addr, mustConnect)
	default:
		err = fmt.Errorf("Unknown abci transport %s", transport)
	}
	return
}

//----------------------------------------

type Callback func(*types2.Request, *types2.Response)

//----------------------------------------

type ReqRes struct {
	*types2.Request
	*sync.WaitGroup
	*types2.Response // Not set atomically, so be sure to use WaitGroup.

	mtx  sync.Mutex
	done bool                   // Gets set to true once *after* WaitGroup.Done().
	cb   func(*types2.Response) // A single callback that may be set.
}

func NewReqRes(req *types2.Request) *ReqRes {
	return &ReqRes{
		Request:   req,
		WaitGroup: waitGroup1(),
		Response:  nil,

		done: false,
		cb:   nil,
	}
}

// Sets the callback for this ReqRes atomically.
// If reqRes is already done, calls cb immediately.
// NOTE: reqRes.cb should not change if reqRes.done.
// NOTE: only one callback is supported.
func (reqRes *ReqRes) SetCallback(cb func(res *types2.Response)) {
	reqRes.mtx.Lock()

	if reqRes.done {
		reqRes.mtx.Unlock()
		cb(reqRes.Response)
		return
	}

	reqRes.cb = cb
	reqRes.mtx.Unlock()
}

func (reqRes *ReqRes) GetCallback() func(*types2.Response) {
	reqRes.mtx.Lock()
	defer reqRes.mtx.Unlock()
	return reqRes.cb
}

// NOTE: it should be safe to read reqRes.cb without locks after this.
func (reqRes *ReqRes) SetDone() {
	reqRes.mtx.Lock()
	reqRes.done = true
	reqRes.mtx.Unlock()
}

func waitGroup1() (wg *sync.WaitGroup) {
	wg = &sync.WaitGroup{}
	wg.Add(1)
	return
}
