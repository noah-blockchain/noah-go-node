package service

import (
	"context"
	pb "github.com/noah-blockchain/node-grpc-gateway/api_pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Halts returns votes
func (s *Service) Halts(_ context.Context, req *pb.HaltsRequest) (*pb.HaltsResponse, error) {
	cState, err := s.blockchain.GetStateForHeight(req.Height)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	blocks := cState.Halts().GetHaltBlocks(req.Height)

	if blocks == nil {
		return &pb.HaltsResponse{}, nil
	}

	var votes []string
	for _, vote := range blocks.List {
		votes = append(votes, vote.Pubkey.String())
	}

	return &pb.HaltsResponse{
		PublicKeys: votes,
	}, nil
}
