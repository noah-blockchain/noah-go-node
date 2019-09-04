package abcicli

import (
	"fmt"
	types2 "github.com/noah-blockchain/noah-go-node/abci/types"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	cmn "github.com/tendermint/tendermint/libs/common"
)

var _ Client = (*grpcClient)(nil)

// A stripped copy of the remoteClient that makes
// synchronous calls using grpc
type grpcClient struct {
	cmn.BaseService
	mustConnect bool

	client types2.ABCIApplicationClient
	conn   *grpc.ClientConn

	mtx   sync.Mutex
	addr  string
	err   error
	resCb func(*types2.Request, *types2.Response) // listens to all callbacks
}

func NewGRPCClient(addr string, mustConnect bool) *grpcClient {
	cli := &grpcClient{
		addr:        addr,
		mustConnect: mustConnect,
	}
	cli.BaseService = *cmn.NewBaseService(nil, "grpcClient", cli)
	return cli
}

func dialerFunc(ctx context.Context, addr string) (net.Conn, error) {
	return cmn.Connect(addr)
}

func (cli *grpcClient) OnStart() error {
	if err := cli.BaseService.OnStart(); err != nil {
		return err
	}
RETRY_LOOP:
	for {
		conn, err := grpc.Dial(cli.addr, grpc.WithInsecure(), grpc.WithContextDialer(dialerFunc))
		if err != nil {
			if cli.mustConnect {
				return err
			}
			cli.Logger.Error(fmt.Sprintf("abci.grpcClient failed to connect to %v.  Retrying...\n", cli.addr), "err", err)
			time.Sleep(time.Second * dialRetryIntervalSeconds)
			continue RETRY_LOOP
		}

		cli.Logger.Info("Dialed server. Waiting for echo.", "addr", cli.addr)
		client := types2.NewABCIApplicationClient(conn)
		cli.conn = conn

	ENSURE_CONNECTED:
		for {
			_, err := client.Echo(context.Background(), &types2.RequestEcho{Message: "hello"}, grpc.WaitForReady(true))
			if err == nil {
				break ENSURE_CONNECTED
			}
			cli.Logger.Error("Echo failed", "err", err)
			time.Sleep(time.Second * echoRetryIntervalSeconds)
		}

		cli.client = client
		return nil
	}
}

func (cli *grpcClient) OnStop() {
	cli.BaseService.OnStop()

	if cli.conn != nil {
		cli.conn.Close()
	}
}

func (cli *grpcClient) StopForError(err error) {
	cli.mtx.Lock()
	if !cli.IsRunning() {
		return
	}

	if cli.err == nil {
		cli.err = err
	}
	cli.mtx.Unlock()

	cli.Logger.Error(fmt.Sprintf("Stopping abci.grpcClient for error: %v", err.Error()))
	cli.Stop()
}

func (cli *grpcClient) Error() error {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	return cli.err
}

// Set listener for all responses
// NOTE: callback may get internally generated flush responses.
func (cli *grpcClient) SetResponseCallback(resCb Callback) {
	cli.mtx.Lock()
	cli.resCb = resCb
	cli.mtx.Unlock()
}

//----------------------------------------
// GRPC calls are synchronous, but some callbacks expect to be called asynchronously
// (eg. the mempool expects to be able to lock to remove bad txs from cache).
// To accommodate, we finish each call in its own go-routine,
// which is expensive, but easy - if you want something better, use the socket protocol!
// maybe one day, if people really want it, we use grpc streams,
// but hopefully not :D

func (cli *grpcClient) EchoAsync(msg string) *ReqRes {
	req := types2.ToRequestEcho(msg)
	res, err := cli.client.Echo(context.Background(), req.GetEcho(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &types2.Response{Value: &types2.Response_Echo{Echo: res}})
}

func (cli *grpcClient) FlushAsync() *ReqRes {
	req := types2.ToRequestFlush()
	res, err := cli.client.Flush(context.Background(), req.GetFlush(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &types2.Response{Value: &types2.Response_Flush{Flush: res}})
}

func (cli *grpcClient) InfoAsync(params types2.RequestInfo) *ReqRes {
	req := types2.ToRequestInfo(params)
	res, err := cli.client.Info(context.Background(), req.GetInfo(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &types2.Response{Value: &types2.Response_Info{Info: res}})
}

func (cli *grpcClient) SetOptionAsync(params types2.RequestSetOption) *ReqRes {
	req := types2.ToRequestSetOption(params)
	res, err := cli.client.SetOption(context.Background(), req.GetSetOption(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &types2.Response{Value: &types2.Response_SetOption{SetOption: res}})
}

func (cli *grpcClient) DeliverTxAsync(params types2.RequestDeliverTx) *ReqRes {
	req := types2.ToRequestDeliverTx(params)
	res, err := cli.client.DeliverTx(context.Background(), req.GetDeliverTx(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &types2.Response{Value: &types2.Response_DeliverTx{DeliverTx: res}})
}

func (cli *grpcClient) CheckTxAsync(params types2.RequestCheckTx) *ReqRes {
	req := types2.ToRequestCheckTx(params)
	res, err := cli.client.CheckTx(context.Background(), req.GetCheckTx(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &types2.Response{Value: &types2.Response_CheckTx{CheckTx: res}})
}

func (cli *grpcClient) QueryAsync(params types2.RequestQuery) *ReqRes {
	req := types2.ToRequestQuery(params)
	res, err := cli.client.Query(context.Background(), req.GetQuery(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &types2.Response{Value: &types2.Response_Query{Query: res}})
}

func (cli *grpcClient) CommitAsync() *ReqRes {
	req := types2.ToRequestCommit()
	res, err := cli.client.Commit(context.Background(), req.GetCommit(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &types2.Response{Value: &types2.Response_Commit{Commit: res}})
}

func (cli *grpcClient) InitChainAsync(params types2.RequestInitChain) *ReqRes {
	req := types2.ToRequestInitChain(params)
	res, err := cli.client.InitChain(context.Background(), req.GetInitChain(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &types2.Response{Value: &types2.Response_InitChain{InitChain: res}})
}

func (cli *grpcClient) BeginBlockAsync(params types2.RequestBeginBlock) *ReqRes {
	req := types2.ToRequestBeginBlock(params)
	res, err := cli.client.BeginBlock(context.Background(), req.GetBeginBlock(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &types2.Response{Value: &types2.Response_BeginBlock{BeginBlock: res}})
}

func (cli *grpcClient) EndBlockAsync(params types2.RequestEndBlock) *ReqRes {
	req := types2.ToRequestEndBlock(params)
	res, err := cli.client.EndBlock(context.Background(), req.GetEndBlock(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &types2.Response{Value: &types2.Response_EndBlock{EndBlock: res}})
}

func (cli *grpcClient) finishAsyncCall(req *types2.Request, res *types2.Response) *ReqRes {
	reqres := NewReqRes(req)
	Response = res   // Set response
	reqres.Done()    // Release waiters
	reqres.SetDone() // so reqRes.SetCallback will run the callback

	// goroutine for callbacks
	go func() {
		cli.mtx.Lock()
		defer cli.mtx.Unlock()

		// Notify client listener if set
		if cli.resCb != nil {
			cli.resCb(Request, res)
		}

		// Notify reqRes listener if set
		if cb := reqres.GetCallback(); cb != nil {
			cb(res)
		}
	}()

	return reqres
}

//----------------------------------------

func (cli *grpcClient) FlushSync() error {
	return nil
}

func (cli *grpcClient) EchoSync(msg string) (*types2.ResponseEcho, error) {
	reqres := cli.EchoAsync(msg)
	// StopForError should already have been called if error is set
	return Response.GetEcho(), cli.Error()
}

func (cli *grpcClient) InfoSync(req types2.RequestInfo) (*types2.ResponseInfo, error) {
	reqres := cli.InfoAsync(req)
	return Response.GetInfo(), cli.Error()
}

func (cli *grpcClient) SetOptionSync(req types2.RequestSetOption) (*types2.ResponseSetOption, error) {
	reqres := cli.SetOptionAsync(req)
	return Response.GetSetOption(), cli.Error()
}

func (cli *grpcClient) DeliverTxSync(params types2.RequestDeliverTx) (*types2.ResponseDeliverTx, error) {
	reqres := cli.DeliverTxAsync(params)
	return Response.GetDeliverTx(), cli.Error()
}

func (cli *grpcClient) CheckTxSync(params types2.RequestCheckTx) (*types2.ResponseCheckTx, error) {
	reqres := cli.CheckTxAsync(params)
	return Response.GetCheckTx(), cli.Error()
}

func (cli *grpcClient) QuerySync(req types2.RequestQuery) (*types2.ResponseQuery, error) {
	reqres := cli.QueryAsync(req)
	return Response.GetQuery(), cli.Error()
}

func (cli *grpcClient) CommitSync() (*types2.ResponseCommit, error) {
	reqres := cli.CommitAsync()
	return Response.GetCommit(), cli.Error()
}

func (cli *grpcClient) InitChainSync(params types2.RequestInitChain) (*types2.ResponseInitChain, error) {
	reqres := cli.InitChainAsync(params)
	return Response.GetInitChain(), cli.Error()
}

func (cli *grpcClient) BeginBlockSync(params types2.RequestBeginBlock) (*types2.ResponseBeginBlock, error) {
	reqres := cli.BeginBlockAsync(params)
	return Response.GetBeginBlock(), cli.Error()
}

func (cli *grpcClient) EndBlockSync(params types2.RequestEndBlock) (*types2.ResponseEndBlock, error) {
	reqres := cli.EndBlockAsync(params)
	return Response.GetEndBlock(), cli.Error()
}
