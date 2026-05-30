package rpc

import (
	"fmt"
	"net"

	rpcpb "xhc2_for_studying/protocol/rpc"

	"google.golang.org/grpc"
)

func ListenAndServeGRPC(addr string, service *C2RPC) error {
	if service == nil {
		return fmt.Errorf("rpc service is nil")
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	rpcpb.RegisterC2Server(grpcServer, &C2GRPCServer{Service: service})
	return grpcServer.Serve(lis)
}
