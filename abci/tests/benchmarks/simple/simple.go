package main

import (
	"bufio"
	"fmt"
	types2 "github.com/noah-blockchain/noah-go-node/abci/types"
	"log"
	"net"
	"reflect"

	cmn "github.com/tendermint/tendermint/libs/common"
)

func main() {

	conn, err := cmn.Connect("unix://test.sock")
	if err != nil {
		log.Fatal(err.Error())
	}

	// Make a bunch of requests
	counter := 0
	for i := 0; ; i++ {
		req := types2.ToRequestEcho("foobar")
		_, err := makeRequest(conn, req)
		if err != nil {
			log.Fatal(err.Error())
		}
		counter++
		if counter%1000 == 0 {
			fmt.Println(counter)
		}
	}
}

func makeRequest(conn net.Conn, req *types2.Request) (*types2.Response, error) {
	var bufWriter = bufio.NewWriter(conn)

	// Write desired request
	err := types2.WriteMessage(req, bufWriter)
	if err != nil {
		return nil, err
	}
	err = types2.WriteMessage(types2.ToRequestFlush(), bufWriter)
	if err != nil {
		return nil, err
	}
	err = bufWriter.Flush()
	if err != nil {
		return nil, err
	}

	// Read desired response
	var res = &types2.Response{}
	err = types2.ReadMessage(conn, res)
	if err != nil {
		return nil, err
	}
	var resFlush = &types2.Response{}
	err = types2.ReadMessage(conn, resFlush)
	if err != nil {
		return nil, err
	}
	if _, ok := resFlush.Value.(*types2.Response_Flush); !ok {
		return nil, fmt.Errorf("Expected flush response but got something else: %v", reflect.TypeOf(resFlush))
	}

	return res, nil
}
