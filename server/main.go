// The C2 server binary runs the HTTP C2 listener, gRPC service, and local
// console against shared runtime state.
package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"xhc2_for_studying/server/c2"
	serverConfig "xhc2_for_studying/server/config"
	"xhc2_for_studying/server/console"
	rpcserver "xhc2_for_studying/server/rpc"
	"xhc2_for_studying/server/store"
)

func main() {
	defaultAddr := getenv("C2_TO_STUDY_ADDR", ":8024")
	addr := flag.String("addr", defaultAddr, "HTTP C2 listen address")
	rpcAddr := flag.String("rpc-addr", ":8025", "gRPC listen address")

	flag.Parse()

	runServer(*addr, *rpcAddr)
}

// runServer starts the HTTP C2 server, gRPC server, and interactive console.
func runServer(addr string, rpcAddr string) {
	// Suppress Gin log output to avoid interfering with the interactive console.
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	cfg, err := serverConfig.Load()
	if err != nil {
		log.Fatalf("load server config: %v", err)
	}

	log.Printf("[+] Age public key: %s", cfg.AgePublicKey)

	stores, closeStores := openStores(cfg)
	defer closeStores()

	httpServer := c2.NewHTTPServer(stores.Beacons, stores.Tasks, stores.Sessions, stores.Implants, cfg.AgePrivateKey, cfg.C2Profile)

	// HTTP C2 listener in background
	go func() {
		if err := httpServer.Run(addr); err != nil {
			log.Fatalf("http server: %v", err)
		}
	}()

	// gRPC server in background
	go func() {
		svc := &rpcserver.C2RPC{
			BeaconStore:  stores.Beacons,
			TaskStore:    stores.Tasks,
			SessionStore: stores.Sessions,
			ImplantStore: stores.Implants,
			Config:       cfg,
		}
		if err := rpcserver.ListenAndServeGRPC(rpcAddr, svc); err != nil {
			log.Fatalf("grpc server: %v", err)
		}
	}()

	// Interactive console in foreground
	con := &console.Console{
		BeaconStore:  stores.Beacons,
		TaskStore:    stores.Tasks,
		SessionStore: stores.Sessions,
	}
	con.Run()
}

func openStores(cfg *serverConfig.ServerConfig) (store.Stores, func()) {
	switch cfg.Database.Driver {
	case "mysql":
		stores, db, err := store.NewMySQLStores(cfg.Database.DSN)
		if err != nil {
			log.Fatalf("open mysql stores: %v", err)
		}
		log.Printf("[+] MySQL persistence enabled")
		return stores, func() {
			db.Close()
		}
	default:
		log.Printf("[+] in-memory persistence enabled")
		return store.NewMemoryStores(), func() {}
	}
}

// getenv returns the environment variable value for key, or fallback if unset.
func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
