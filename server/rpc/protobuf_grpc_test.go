package rpc

import (
	"context"
	"net"
	"testing"

	rpcpb "xhc2_for_studying/protocol/rpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"xhc2_for_studying/server/store"
)

func TestGeneratedGRPCClientUsesProtobufService(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	rpcpb.RegisterC2Server(grpcServer, &C2GRPCServer{Service: &C2RPC{
		BeaconStore: store.NewBeaconStore(),
		TaskStore:   store.NewServerTaskStore(),
	}})

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Logf("grpc server stopped: %v", err)
		}
	}()
	defer grpcServer.Stop()
	defer listener.Close()

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("dial test grpc server: %v", err)
	}
	defer conn.Close()

	client := rpcpb.NewC2Client(conn)
	_, err = client.GetTaskResult(context.Background(), &rpcpb.TaskResultRequest{TaskId: "missing-task"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetTaskResult error code = %s, want %s; error = %v", status.Code(err), codes.NotFound, err)
	}
}
