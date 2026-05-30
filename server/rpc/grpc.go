package rpc

import (
	"context"
	"strings"

	rpcpb "xhc2_for_studying/protocol/rpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// C2GRPCServer adapts C2RPC business logic to the generated protobuf gRPC API.
type C2GRPCServer struct {
	rpcpb.UnimplementedC2Server

	Service *C2RPC
}

func (s *C2GRPCServer) ListBeacons(_ context.Context, _ *rpcpb.Empty) (*rpcpb.BeaconList, error) {
	var reply []BeaconInfo
	if err := s.Service.ListBeacons(Empty{}, &reply); err != nil {
		return nil, toGRPCError(err)
	}

	out := &rpcpb.BeaconList{Beacons: make([]*rpcpb.BeaconInfo, 0, len(reply))}
	for _, b := range reply {
		out.Beacons = append(out.Beacons, &rpcpb.BeaconInfo{
			Id:          b.ID,
			Hostname:    b.Hostname,
			Username:    b.Username,
			Os:          b.OS,
			Arch:        b.Arch,
			Interval:    b.Interval,
			Jitter:      b.Jitter,
			LastCheckin: b.LastCheckIn,
		})
	}
	return out, nil
}

func (s *C2GRPCServer) ListTasks(_ context.Context, _ *rpcpb.Empty) (*rpcpb.TaskList, error) {
	var reply []TaskInfo
	if err := s.Service.ListTasks(Empty{}, &reply); err != nil {
		return nil, toGRPCError(err)
	}

	out := &rpcpb.TaskList{Tasks: make([]*rpcpb.TaskInfo, 0, len(reply))}
	for _, task := range reply {
		out.Tasks = append(out.Tasks, &rpcpb.TaskInfo{
			TaskId:    task.TaskID,
			ImplantId: task.ImplantID,
			Type:      task.Type,
			Status:    task.Status,
		})
	}
	return out, nil
}

func (s *C2GRPCServer) CreateTask(_ context.Context, req *rpcpb.CreateTaskRequest) (*rpcpb.CreateTaskResponse, error) {
	var reply CreateTaskResponse
	if err := s.Service.CreateTask(CreateTaskRequest{
		TaskType:  req.GetTaskType(),
		ImplantID: req.GetImplantId(),
		Payload:   req.GetPayload(),
	}, &reply); err != nil {
		return nil, toGRPCError(err)
	}
	return &rpcpb.CreateTaskResponse{TaskId: reply.TaskID}, nil
}

func (s *C2GRPCServer) GenerateImplant(_ context.Context, req *rpcpb.GenerateImplantRequest) (*rpcpb.GenerateImplantResponse, error) {
	var reply GenerateImplantResponse
	if err := s.Service.GenerateImplant(GenerateImplantRequest{
		ServerURL:  req.GetServerUrl(),
		PathPrefix: req.GetPathPrefix(),
		Interval:   req.GetInterval(),
		Jitter:     req.Jitter,
		GOOS:       req.GetGoos(),
		GOARCH:     req.GetGoarch(),
	}, &reply); err != nil {
		return nil, toGRPCError(err)
	}
	return &rpcpb.GenerateImplantResponse{
		Digest:   reply.Digest,
		Binary:   reply.Binary,
		Filename: reply.Filename,
	}, nil
}

func (s *C2GRPCServer) GetTaskResult(_ context.Context, req *rpcpb.TaskResultRequest) (*rpcpb.TaskResultResponse, error) {
	var reply TaskResultResponse
	if err := s.Service.GetTaskResult(TaskResultRequest{TaskID: req.GetTaskId()}, &reply); err != nil {
		return nil, toGRPCError(err)
	}
	return &rpcpb.TaskResultResponse{
		TaskId:    reply.TaskID,
		Type:      reply.Type,
		ImplantId: reply.ImplantID,
		Status:    reply.Status,
		Error:     reply.Error,
		Output:    reply.Output,
		Completed: reply.Completed,
	}, nil
}

func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	switch {
	case strings.HasPrefix(msg, "unknown task type:"),
		strings.HasPrefix(msg, "interval must be"),
		strings.HasPrefix(msg, "jitter must be"),
		strings.HasSuffix(msg, "is required"):
		return status.Error(codes.InvalidArgument, msg)
	case strings.HasPrefix(msg, "beacon not found:"),
		strings.HasPrefix(msg, "task not found:"):
		return status.Error(codes.NotFound, msg)
	default:
		return status.Error(codes.Internal, msg)
	}
}
