package service

import (
	"context"
	"github.com/noah-blockchain/noah-go-node/config"
	"github.com/noah-blockchain/noah-go-node/core/noah"
	"github.com/tendermint/go-amino"
	tmNode "github.com/tendermint/tendermint/node"
	rpc "github.com/tendermint/tendermint/rpc/client/local"
	"google.golang.org/grpc/status"
	"time"
)

// Service is gRPC implementation ApiServiceServer
type Service struct {
	cdc        *amino.Codec
	blockchain *noah.Blockchain
	client     *rpc.Local
	tmNode     *tmNode.Node
	noahCfg  *config.Config
	version    string
}

// NewService create gRPC server implementation
func NewService(cdc *amino.Codec, blockchain *noah.Blockchain, client *rpc.Local, node *tmNode.Node, noahCfg *config.Config, version string) *Service {
	return &Service{
		cdc:        cdc,
		blockchain: blockchain,
		client:     client,
		noahCfg:  noahCfg,
		version:    version,
		tmNode:     node,
	}
}

// TimeoutDuration gRPC
func (s *Service) TimeoutDuration() time.Duration {
	return s.noahCfg.APIv2TimeoutDuration
}

func (s *Service) createError(statusErr *status.Status, data string) error {
	if len(data) == 0 {
		return statusErr.Err()
	}

	detailsMap, err := encodeToStruct([]byte(data))
	if err != nil {
		s.client.Logger.Error(err.Error())
		return statusErr.Err()
	}

	withDetails, err := statusErr.WithDetails(detailsMap)
	if err != nil {
		s.client.Logger.Error(err.Error())
		return statusErr.Err()
	}

	return withDetails.Err()
}

func (s *Service) checkTimeout(ctx context.Context) *status.Status {
	select {
	case <-ctx.Done():
		return status.FromContextError(ctx.Err())
	default:
		return nil
	}
}
