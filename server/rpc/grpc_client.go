package rpc

import (
	"context"

	rpcpb "xhc2_for_studying/protocol/rpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type C2GRPCClient struct {
	conn   *grpc.ClientConn
	client rpcpb.C2Client
}

func NewC2GRPCClient(ctx context.Context, addr string) (*C2GRPCClient, error) {
	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &C2GRPCClient{conn: conn, client: rpcpb.NewC2Client(conn)}, nil
}

func (c *C2GRPCClient) Close() error {
	return c.conn.Close()
}

func (c *C2GRPCClient) ListBeacons(ctx context.Context) ([]BeaconInfo, error) {
	resp, err := c.client.ListBeacons(ctx, &rpcpb.Empty{})
	if err != nil {
		return nil, err
	}

	out := make([]BeaconInfo, 0, len(resp.GetBeacons()))
	for _, b := range resp.GetBeacons() {
		out = append(out, BeaconInfo{
			ID:          b.GetId(),
			Hostname:    b.GetHostname(),
			Username:    b.GetUsername(),
			OS:          b.GetOs(),
			Arch:        b.GetArch(),
			Interval:    b.GetInterval(),
			Jitter:      b.GetJitter(),
			LastCheckIn: b.GetLastCheckin(),
		})
	}
	return out, nil
}

func (c *C2GRPCClient) ListTasks(ctx context.Context) ([]TaskInfo, error) {
	resp, err := c.client.ListTasks(ctx, &rpcpb.Empty{})
	if err != nil {
		return nil, err
	}

	out := make([]TaskInfo, 0, len(resp.GetTasks()))
	for _, task := range resp.GetTasks() {
		out = append(out, TaskInfo{
			TaskID:    task.GetTaskId(),
			ImplantID: task.GetImplantId(),
			Type:      task.GetType(),
			Status:    task.GetStatus(),
		})
	}
	return out, nil
}

func (c *C2GRPCClient) CreateTask(ctx context.Context, req CreateTaskRequest) (CreateTaskResponse, error) {
	resp, err := c.client.CreateTask(ctx, &rpcpb.CreateTaskRequest{
		TaskType:  req.TaskType,
		ImplantId: req.ImplantID,
		Payload:   req.Payload,
	})
	if err != nil {
		return CreateTaskResponse{}, err
	}
	return CreateTaskResponse{TaskID: resp.GetTaskId()}, nil
}

func (c *C2GRPCClient) GenerateImplant(ctx context.Context, req GenerateImplantRequest) (GenerateImplantResponse, error) {
	resp, err := c.client.GenerateImplant(ctx, &rpcpb.GenerateImplantRequest{
		ServerUrl:  req.ServerURL,
		PathPrefix: req.PathPrefix,
		Interval:   req.Interval,
		Jitter:     req.Jitter,
		Goos:       req.GOOS,
		Goarch:     req.GOARCH,
	})
	if err != nil {
		return GenerateImplantResponse{}, err
	}
	return GenerateImplantResponse{
		Digest:   resp.GetDigest(),
		Filename: resp.GetFilename(),
		Binary:   resp.GetBinary(),
	}, nil
}

func (c *C2GRPCClient) GetTaskResult(ctx context.Context, req TaskResultRequest) (TaskResultResponse, error) {
	resp, err := c.client.GetTaskResult(ctx, &rpcpb.TaskResultRequest{TaskId: req.TaskID})
	if err != nil {
		return TaskResultResponse{}, err
	}
	return TaskResultResponse{
		TaskID:    resp.GetTaskId(),
		Type:      resp.GetType(),
		ImplantID: resp.GetImplantId(),
		Status:    resp.GetStatus(),
		Error:     resp.GetError(),
		Output:    resp.GetOutput(),
		Completed: resp.GetCompleted(),
	}, nil
}
