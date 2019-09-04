package abcicli_test

import (
	"errors"
	"fmt"
	server2 "github.com/noah-blockchain/noah-go-node/abci/server"
	types2 "github.com/noah-blockchain/noah-go-node/abci/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmn "github.com/tendermint/tendermint/libs/common"
)

func TestSocketClientStopForErrorDeadlock(t *testing.T) {
	c := NewSocketClient(":80", false)
	err := errors.New("foo-tendermint")

	// See Issue https://github.com/tendermint/abci/issues/114
	doneChan := make(chan bool)
	go func() {
		defer close(doneChan)
		c.StopForError(err)
		c.StopForError(err)
	}()

	select {
	case <-doneChan:
	case <-time.After(time.Second * 4):
		t.Fatalf("Test took too long, potential deadlock still exists")
	}
}

func TestProperSyncCalls(t *testing.T) {
	app := slowApp{}

	s, c := setupClientServer(t, app)
	defer s.Stop()
	defer c.Stop()

	resp := make(chan error, 1)
	go func() {
		// This is BeginBlockSync unrolled....
		reqres := c.BeginBlockAsync(types2.RequestBeginBlock{})
		c.FlushSync()
		res := reqres.Response.GetBeginBlock()
		require.NotNil(t, res)
		resp <- c.Error()
	}()

	select {
	case <-time.After(time.Second):
		require.Fail(t, "No response arrived")
	case err, ok := <-resp:
		require.True(t, ok, "Must not close channel")
		assert.NoError(t, err, "This should return success")
	}
}

func TestHangingSyncCalls(t *testing.T) {
	app := slowApp{}

	s, c := setupClientServer(t, app)
	defer s.Stop()
	defer c.Stop()

	resp := make(chan error, 1)
	go func() {
		// Start BeginBlock and flush it
		reqres := c.BeginBlockAsync(types2.RequestBeginBlock{})
		flush := c.FlushAsync()
		// wait 20 ms for all events to travel socket, but
		// no response yet from server
		time.Sleep(20 * time.Millisecond)
		// kill the server, so the connections break
		s.Stop()

		// wait for the response from BeginBlock
		reqres.Wait()
		flush.Wait()
		resp <- c.Error()
	}()

	select {
	case <-time.After(time.Second):
		require.Fail(t, "No response arrived")
	case err, ok := <-resp:
		require.True(t, ok, "Must not close channel")
		assert.Error(t, err, "We should get EOF error")
	}
}

func setupClientServer(t *testing.T, app types2.Application) (
	cmn.Service, Client) {
	// some port between 20k and 30k
	port := 20000 + cmn.RandInt32()%10000
	addr := fmt.Sprintf("localhost:%d", port)

	s, err := server2.NewServer(addr, "socket", app)
	require.NoError(t, err)
	err = s.Start()
	require.NoError(t, err)

	c := NewSocketClient(addr, true)
	err = c.Start()
	require.NoError(t, err)

	return s, c
}

type slowApp struct {
	types2.BaseApplication
}

func (slowApp) BeginBlock(req types2.RequestBeginBlock) types2.ResponseBeginBlock {
	time.Sleep(200 * time.Millisecond)
	return types2.ResponseBeginBlock{}
}
