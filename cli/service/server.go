package service

import (
	"context"
	"github.com/noah-blockchain/noah-go-node/cli/pb"
	"google.golang.org/grpc"
	"net"
	"os"
)

func StartCLIServer(socketPath string, manager pb.ManagerServiceServer, ctx context.Context) error {
	if err := os.RemoveAll(socketPath); err != nil {
		return err
	}

	lis, err := net.ListenUnix("unix", &net.UnixAddr{Name: socketPath, Net: "unix"})
	if err != nil {
		return err
	}

	server := grpc.NewServer()

	pb.RegisterManagerServiceServer(server, manager)

	kill := make(chan struct{})
	defer close(kill)
	go func() {
		select {
		case <-ctx.Done():
			server.GracefulStop()
		case <-kill:
		}
		return
	}()

	if err := server.Serve(lis); err != nil {
		return err
	}

	return nil
}
