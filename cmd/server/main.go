package main

import (
	"context"
	"net"
	"os/signal"
	"syscall"

	"github.com/gbh007/p2p-chat/internal/server"
	"github.com/gbh007/p2p-chat/proto/gen"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	err := Serve(ctx)
	if err != nil {
		panic(err)
	}
}

func Serve(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	s := server.New()

	grpcServer := grpc.NewServer()
	gen.RegisterServerServer(grpcServer, s)

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	err = grpcServer.Serve(lis)
	if err != nil {
		return err
	}

	return nil
}
