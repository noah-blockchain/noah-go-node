package tests

import (
	"github.com/noah-blockchain/noah-go-node/abci/client"
	kvstore2 "github.com/noah-blockchain/noah-go-node/abci/example/kvstore"
	server2 "github.com/noah-blockchain/noah-go-node/abci/server"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientServerNoAddrPrefix(t *testing.T) {
	addr := "localhost:26658"
	transport := "socket"
	app := kvstore2.NewKVStoreApplication()

	server, err := server2.NewServer(addr, transport, app)
	assert.NoError(t, err, "expected no error on NewServer")
	err = server.Start()
	assert.NoError(t, err, "expected no error on server.Start")

	client, err := abcicli.NewClient(addr, transport, true)
	assert.NoError(t, err, "expected no error on NewClient")
	err = client.Start()
	assert.NoError(t, err, "expected no error on client.Start")
}
