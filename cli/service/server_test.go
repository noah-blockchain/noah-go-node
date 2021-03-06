package service

import (
	"context"
	"github.com/noah-blockchain/noah-go-node/config"
	"github.com/noah-blockchain/noah-go-node/core/noah"
	"github.com/tendermint/tendermint/node"
	rpc "github.com/tendermint/tendermint/rpc/client/local"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"
)

func TestStartCLIServer(t *testing.T) {
	var (
		blockchain *noah.Blockchain
		tmRPC      *rpc.Local
		tmNode     *node.Node
		cfg        *config.Config
	)
	ctx, cancel := context.WithCancel(context.Background())
	socketPath, _ := filepath.Abs(filepath.Join(".", "file.sock"))
	_ = ioutil.WriteFile(socketPath, []byte("address already in use"), 0644)
	go func() {
		err := StartCLIServer(socketPath, NewManager(blockchain, tmRPC, tmNode, cfg), ctx)
		if err != nil {
			t.Log(err)
		}
	}()
	time.Sleep(time.Millisecond)
	console, err := NewCLI(socketPath)
	if err != nil {
		t.Log(err)
	}
	err = console.Execute([]string{"test"})
	if err != nil {
		t.Log(err)
	}
	cancel()
}
