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

func TestGRPCCreateTaskReturnsInvalidArgumentForUnknownTaskType(t *testing.T) {
	client, cleanup := newTestGRPCClient(t, &C2RPC{
		BeaconStore: store.NewBeaconStore(),
		TaskStore:   store.NewServerTaskStore(),
	})
	defer cleanup()

	_, err := client.CreateTask(context.Background(), CreateTaskRequest{
		ImplantID: "beacon-1",
		TaskType:  "unknown",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("CreateTask error code = %s, want %s; error = %v", status.Code(err), codes.InvalidArgument, err)
	}
}

func TestGRPCGetTaskResultReturnsNotFoundForMissingTask(t *testing.T) {
	client, cleanup := newTestGRPCClient(t, &C2RPC{
		BeaconStore: store.NewBeaconStore(),
		TaskStore:   store.NewServerTaskStore(),
	})
	defer cleanup()

	_, err := client.GetTaskResult(context.Background(), TaskResultRequest{TaskID: "missing-task"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetTaskResult error code = %s, want %s; error = %v", status.Code(err), codes.NotFound, err)
	}
}

func newTestGRPCClient(t *testing.T, service *C2RPC) (*C2GRPCClient, func()) {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	rpcpb.RegisterC2Server(grpcServer, &C2GRPCServer{Service: service})

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Logf("grpc server stopped: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("dial test grpc server: %v", err)
	}

	cleanup := func() {
		conn.Close()
		grpcServer.Stop()
		listener.Close()
	}
	return &C2GRPCClient{conn: conn, client: rpcpb.NewC2Client(conn)}, cleanup
}
