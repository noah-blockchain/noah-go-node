package abcicli

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	types2 "github.com/noah-blockchain/noah-go-node/abci/types"
	"net"
	"reflect"
	"sync"
	"time"

	cmn "github.com/tendermint/tendermint/libs/common"
)

const reqQueueSize = 256 // TODO make configurable
// const maxResponseSize = 1048576 // 1MB TODO make configurable
const flushThrottleMS = 20 // Don't wait longer than...

var _ Client = (*socketClient)(nil)

// This is goroutine-safe, but users should beware that
// the application in general is not meant to be interfaced
// with concurrent callers.
type socketClient struct {
	cmn.BaseService

	addr        string
	mustConnect bool
	conn        net.Conn

	reqQueue   chan *ReqRes
	flushTimer *cmn.ThrottleTimer

	mtx     sync.Mutex
	err     error
	reqSent *list.List                              // list of requests sent, waiting for response
	resCb   func(*types2.Request, *types2.Response) // called on all requests, if set.

}

func NewSocketClient(addr string, mustConnect bool) *socketClient {
	cli := &socketClient{
		reqQueue:    make(chan *ReqRes, reqQueueSize),
		flushTimer:  cmn.NewThrottleTimer("socketClient", flushThrottleMS),
		mustConnect: mustConnect,

		addr:    addr,
		reqSent: list.New(),
		resCb:   nil,
	}
	cli.BaseService = *cmn.NewBaseService(nil, "socketClient", cli)
	return cli
}

func (cli *socketClient) OnStart() error {
	var err error
	var conn net.Conn
RETRY_LOOP:
	for {
		conn, err = cmn.Connect(cli.addr)
		if err != nil {
			if cli.mustConnect {
				return err
			}
			cli.Logger.Error(fmt.Sprintf("abci.socketClient failed to connect to %v.  Retrying...", cli.addr), "err", err)
			time.Sleep(time.Second * dialRetryIntervalSeconds)
			continue RETRY_LOOP
		}
		cli.conn = conn

		go cli.sendRequestsRoutine(conn)
		go cli.recvResponseRoutine(conn)

		return nil
	}
}

func (cli *socketClient) OnStop() {
	if cli.conn != nil {
		cli.conn.Close()
	}

	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	cli.flushQueue()
}

// Stop the client and set the error
func (cli *socketClient) StopForError(err error) {
	if !cli.IsRunning() {
		return
	}

	cli.mtx.Lock()
	if cli.err == nil {
		cli.err = err
	}
	cli.mtx.Unlock()

	cli.Logger.Error(fmt.Sprintf("Stopping abci.socketClient for error: %v", err.Error()))
	cli.Stop()
}

func (cli *socketClient) Error() error {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	return cli.err
}

// Set listener for all responses
// NOTE: callback may get internally generated flush responses.
func (cli *socketClient) SetResponseCallback(resCb Callback) {
	cli.mtx.Lock()
	cli.resCb = resCb
	cli.mtx.Unlock()
}

//----------------------------------------

func (cli *socketClient) sendRequestsRoutine(conn net.Conn) {

	w := bufio.NewWriter(conn)
	for {
		select {
		case <-cli.flushTimer.Ch:
			select {
			case cli.reqQueue <- NewReqRes(types2.ToRequestFlush()):
			default:
				// Probably will fill the buffer, or retry later.
			}
		case <-cli.Quit():
			return
		case reqres := <-cli.reqQueue:
			cli.willSendReq(reqres)
			err := types2.WriteMessage(Request, w)
			if err != nil {
				cli.StopForError(fmt.Errorf("Error writing msg: %v", err))
				return
			}
			// cli.Logger.Debug("Sent request", "requestType", reflect.TypeOf(reqres.Request), "request", reqres.Request)
			if _, ok := Request.Value.(*types2.Request_Flush); ok {
				err = w.Flush()
				if err != nil {
					cli.StopForError(fmt.Errorf("Error flushing writer: %v", err))
					return
				}
			}
		}
	}
}

func (cli *socketClient) recvResponseRoutine(conn net.Conn) {

	r := bufio.NewReader(conn) // Buffer reads
	for {
		var res = &types2.Response{}
		err := types2.ReadMessage(r, res)
		if err != nil {
			cli.StopForError(err)
			return
		}
		switch r := res.Value.(type) {
		case *types2.Response_Exception:
			// XXX After setting cli.err, release waiters (e.g. reqres.Done())
			cli.StopForError(errors.New(r.Exception.Error))
			return
		default:
			// cli.Logger.Debug("Received response", "responseType", reflect.TypeOf(res), "response", res)
			err := cli.didRecvResponse(res)
			if err != nil {
				cli.StopForError(err)
				return
			}
		}
	}
}

func (cli *socketClient) willSendReq(reqres *ReqRes) {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	cli.reqSent.PushBack(reqres)
}

func (cli *socketClient) didRecvResponse(res *types2.Response) error {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()

	// Get the first ReqRes
	next := cli.reqSent.Front()
	if next == nil {
		return fmt.Errorf("Unexpected result type %v when nothing expected", reflect.TypeOf(res.Value))
	}
	reqres := next.Value.(*ReqRes)
	if !resMatchesReq(Request, res) {
		return fmt.Errorf("Unexpected result type %v when response to %v expected",
			reflect.TypeOf(res.Value), reflect.TypeOf(Request.Value))
	}

	Response = res           // Set response
	reqres.Done()            // Release waiters
	cli.reqSent.Remove(next) // Pop first item from linked list

	// Notify client listener if set (global callback).
	if cli.resCb != nil {
		cli.resCb(Request, res)
	}

	// Notify reqRes listener if set (request specific callback).
	// NOTE: it is possible this callback isn't set on the reqres object.
	// at this point, in which case it will be called after, when it is set.
	if cb := reqres.GetCallback(); cb != nil {
		cb(res)
	}

	return nil
}

//----------------------------------------

func (cli *socketClient) EchoAsync(msg string) *ReqRes {
	return cli.queueRequest(types2.ToRequestEcho(msg))
}

func (cli *socketClient) FlushAsync() *ReqRes {
	return cli.queueRequest(types2.ToRequestFlush())
}

func (cli *socketClient) InfoAsync(req types2.RequestInfo) *ReqRes {
	return cli.queueRequest(types2.ToRequestInfo(req))
}

func (cli *socketClient) SetOptionAsync(req types2.RequestSetOption) *ReqRes {
	return cli.queueRequest(types2.ToRequestSetOption(req))
}

func (cli *socketClient) DeliverTxAsync(req types2.RequestDeliverTx) *ReqRes {
	return cli.queueRequest(types2.ToRequestDeliverTx(req))
}

func (cli *socketClient) CheckTxAsync(req types2.RequestCheckTx) *ReqRes {
	return cli.queueRequest(types2.ToRequestCheckTx(req))
}

func (cli *socketClient) QueryAsync(req types2.RequestQuery) *ReqRes {
	return cli.queueRequest(types2.ToRequestQuery(req))
}

func (cli *socketClient) CommitAsync() *ReqRes {
	return cli.queueRequest(types2.ToRequestCommit())
}

func (cli *socketClient) InitChainAsync(req types2.RequestInitChain) *ReqRes {
	return cli.queueRequest(types2.ToRequestInitChain(req))
}

func (cli *socketClient) BeginBlockAsync(req types2.RequestBeginBlock) *ReqRes {
	return cli.queueRequest(types2.ToRequestBeginBlock(req))
}

func (cli *socketClient) EndBlockAsync(req types2.RequestEndBlock) *ReqRes {
	return cli.queueRequest(types2.ToRequestEndBlock(req))
}

//----------------------------------------

func (cli *socketClient) FlushSync() error {
	reqRes := cli.queueRequest(types2.ToRequestFlush())
	if err := cli.Error(); err != nil {
		return err
	}
	reqRes.Wait() // NOTE: if we don't flush the queue, its possible to get stuck here
	return cli.Error()
}

func (cli *socketClient) EchoSync(msg string) (*types2.ResponseEcho, error) {
	reqres := cli.queueRequest(types2.ToRequestEcho(msg))
	cli.FlushSync()
	return Response.GetEcho(), cli.Error()
}

func (cli *socketClient) InfoSync(req types2.RequestInfo) (*types2.ResponseInfo, error) {
	reqres := cli.queueRequest(types2.ToRequestInfo(req))
	cli.FlushSync()
	return Response.GetInfo(), cli.Error()
}

func (cli *socketClient) SetOptionSync(req types2.RequestSetOption) (*types2.ResponseSetOption, error) {
	reqres := cli.queueRequest(types2.ToRequestSetOption(req))
	cli.FlushSync()
	return Response.GetSetOption(), cli.Error()
}

func (cli *socketClient) DeliverTxSync(req types2.RequestDeliverTx) (*types2.ResponseDeliverTx, error) {
	reqres := cli.queueRequest(types2.ToRequestDeliverTx(req))
	cli.FlushSync()
	return Response.GetDeliverTx(), cli.Error()
}

func (cli *socketClient) CheckTxSync(req types2.RequestCheckTx) (*types2.ResponseCheckTx, error) {
	reqres := cli.queueRequest(types2.ToRequestCheckTx(req))
	cli.FlushSync()
	return Response.GetCheckTx(), cli.Error()
}

func (cli *socketClient) QuerySync(req types2.RequestQuery) (*types2.ResponseQuery, error) {
	reqres := cli.queueRequest(types2.ToRequestQuery(req))
	cli.FlushSync()
	return Response.GetQuery(), cli.Error()
}

func (cli *socketClient) CommitSync() (*types2.ResponseCommit, error) {
	reqres := cli.queueRequest(types2.ToRequestCommit())
	cli.FlushSync()
	return Response.GetCommit(), cli.Error()
}

func (cli *socketClient) InitChainSync(req types2.RequestInitChain) (*types2.ResponseInitChain, error) {
	reqres := cli.queueRequest(types2.ToRequestInitChain(req))
	cli.FlushSync()
	return Response.GetInitChain(), cli.Error()
}

func (cli *socketClient) BeginBlockSync(req types2.RequestBeginBlock) (*types2.ResponseBeginBlock, error) {
	reqres := cli.queueRequest(types2.ToRequestBeginBlock(req))
	cli.FlushSync()
	return Response.GetBeginBlock(), cli.Error()
}

func (cli *socketClient) EndBlockSync(req types2.RequestEndBlock) (*types2.ResponseEndBlock, error) {
	reqres := cli.queueRequest(types2.ToRequestEndBlock(req))
	cli.FlushSync()
	return Response.GetEndBlock(), cli.Error()
}

//----------------------------------------

func (cli *socketClient) queueRequest(req *types2.Request) *ReqRes {
	reqres := NewReqRes(req)

	// TODO: set cli.err if reqQueue times out
	cli.reqQueue <- reqres

	// Maybe auto-flush, or unset auto-flush
	switch req.Value.(type) {
	case *types2.Request_Flush:
		cli.flushTimer.Unset()
	default:
		cli.flushTimer.Set()
	}

	return reqres
}

func (cli *socketClient) flushQueue() {
	// mark all in-flight messages as resolved (they will get cli.Error())
	for req := cli.reqSent.Front(); req != nil; req = req.Next() {
		reqres := req.Value.(*ReqRes)
		reqres.Done()
	}

	// mark all queued messages as resolved
LOOP:
	for {
		select {
		case reqres := <-cli.reqQueue:
			reqres.Done()
		default:
			break LOOP
		}
	}
}

//----------------------------------------

func resMatchesReq(req *types2.Request, res *types2.Response) (ok bool) {
	switch req.Value.(type) {
	case *types2.Request_Echo:
		_, ok = res.Value.(*types2.Response_Echo)
	case *types2.Request_Flush:
		_, ok = res.Value.(*types2.Response_Flush)
	case *types2.Request_Info:
		_, ok = res.Value.(*types2.Response_Info)
	case *types2.Request_SetOption:
		_, ok = res.Value.(*types2.Response_SetOption)
	case *types2.Request_DeliverTx:
		_, ok = res.Value.(*types2.Response_DeliverTx)
	case *types2.Request_CheckTx:
		_, ok = res.Value.(*types2.Response_CheckTx)
	case *types2.Request_Commit:
		_, ok = res.Value.(*types2.Response_Commit)
	case *types2.Request_Query:
		_, ok = res.Value.(*types2.Response_Query)
	case *types2.Request_InitChain:
		_, ok = res.Value.(*types2.Response_InitChain)
	case *types2.Request_BeginBlock:
		_, ok = res.Value.(*types2.Response_BeginBlock)
	case *types2.Request_EndBlock:
		_, ok = res.Value.(*types2.Response_EndBlock)
	}
	return ok
}
