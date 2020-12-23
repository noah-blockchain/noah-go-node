package service

import (
	"context"
	"github.com/noah-blockchain/noah-go-node/core/types"
	pb "github.com/noah-blockchain/node-grpc-gateway/api_pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Validators returns list of active validators.
func (s *Service) Validators(ctx context.Context, req *pb.ValidatorsRequest) (*pb.ValidatorsResponse, error) {
	height := int64(req.Height)
	if height == 0 {
		height = int64(s.blockchain.Height())
	}

	tmVals, err := s.client.Validators(&height, 1, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if timeoutStatus := s.checkTimeout(ctx); timeoutStatus != nil {
		return nil, timeoutStatus.Err()
	}

	responseValidators := make([]*pb.ValidatorsResponse_Result, 0, len(tmVals.Validators))
	for _, val := range tmVals.Validators {
		var pk types.Pubkey
		copy(pk[:], val.PubKey.Bytes()[5:])
		responseValidators = append(responseValidators, &pb.ValidatorsResponse_Result{
			PublicKey:   pk.String(),
			VotingPower: uint64(val.VotingPower),
		})
	}
	return &pb.ValidatorsResponse{Validators: responseValidators}, nil
}
