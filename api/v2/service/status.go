package service

import (
	"context"
	"fmt"
	pb "github.com/noah-blockchain/node-grpc-gateway/api_pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// Status returns current min gas price.
func (s *Service) Status(context.Context, *empty.Empty) (*pb.StatusResponse, error) {
	result, err := s.client.Status()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	cState := s.blockchain.CurrentState()
	cState.RLock()
	defer cState.RUnlock()

	return &pb.StatusResponse{
		Version:           s.version,
		Network:           result.NodeInfo.Network,
		LatestBlockHash:   fmt.Sprintf("%X", result.SyncInfo.LatestBlockHash),
		LatestAppHash:     fmt.Sprintf("%X", result.SyncInfo.LatestAppHash),
		LatestBlockHeight: uint64(result.SyncInfo.LatestBlockHeight),
		LatestBlockTime:   result.SyncInfo.LatestBlockTime.Format(time.RFC3339Nano),
		KeepLastStates:    uint64(s.noahCfg.BaseConfig.KeepLastStates),
		TotalSlashed:      cState.App().GetTotalSlashed().String(),
		CatchingUp:        result.SyncInfo.CatchingUp,
		PublicKey:         fmt.Sprintf("NOAHp%x", result.ValidatorInfo.PubKey.Bytes()[5:]),
		NodeId:            string(result.NodeInfo.ID()),
	}, nil
}
