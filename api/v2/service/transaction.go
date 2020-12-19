package service

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/noah-blockchain/noah-go-node/core/transaction"
	pb "github.com/noah-blockchain/node-grpc-gateway/api_pb"
	"github.com/golang/protobuf/jsonpb"
	_struct "github.com/golang/protobuf/ptypes/struct"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) Transaction(_ context.Context, req *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	if len(req.Hash) < 3 {
		return new(pb.TransactionResponse), status.Error(codes.InvalidArgument, "invalid hash")
	}
	decodeString, err := hex.DecodeString(req.Hash[2:])
	if err != nil {
		return new(pb.TransactionResponse), status.Error(codes.InvalidArgument, err.Error())
	}

	tx, err := s.client.Tx(decodeString, false)
	if err != nil {
		return new(pb.TransactionResponse), status.Error(codes.FailedPrecondition, err.Error())
	}

	decodedTx, _ := transaction.TxDecoder.DecodeFromBytes(tx.Tx)
	sender, _ := decodedTx.Sender()

	tags := make(map[string]string)
	for _, tag := range tx.TxResult.Events[0].Attributes {
		tags[string(tag.Key)] = string(tag.Value)
	}

	cState := s.blockchain.CurrentState()

	cState.RLock()
	defer cState.RUnlock()

	if timeoutStatus := s.checkTimeout(ctx); timeoutStatus != nil {
		return nil, timeoutStatus.Err()
	}

	dataStruct, err := encode(decodedTx.GetDecodedData(), cState.Coins())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.TransactionResponse{
		Hash:     "Mt" + strings.ToLower(hex.EncodeToString(tx.Tx.Hash())),
		RawTx:    fmt.Sprintf("%x", []byte(tx.Tx)),
		Height:   uint64(tx.Height),
		Index:    uint64(tx.Index),
		From:     sender.String(),
		Nonce:    decodedTx.Nonce,
		GasPrice: uint64(decodedTx.GasPrice),
		GasCoin: &pb.Coin{
			Id:     uint64(decodedTx.GasCoin),
			Symbol: cState.Coins().GetCoin(decodedTx.GasCoin).GetFullSymbol(),
		},
		Gas:     uint64(decodedTx.Gas()),
		Type:    uint64(decodedTx.Type),
		Data:    dataStruct,
		Payload: decodedTx.Payload,
		Tags:    tags,
		Code:    uint64(tx.TxResult.Code),
		Log:     tx.TxResult.Log,
	}, nil
}
