package main

import (
	"flag"
	"log"
	"os"

	"xhc2_for_studying/server/c2"
	serverConfig "xhc2_for_studying/server/config"
	"xhc2_for_studying/server/store"
)

func main() {
	defaultAddr := getenv("C2_TO_STUDY_ADDR", ":8024")
	addr := flag.String("addr", defaultAddr, "HTTP listen address")
	flag.Parse()

	cfg, err := serverConfig.Load()
	if err != nil {
		log.Fatalf("load server config: %v", err)
	}

	beaconStore := store.NewBeaconStore()
	taskStore := store.NewServerTaskStore()

	httpServer := c2.NewHTTPServer(beaconStore, taskStore, cfg.C2Profile)
	if err := httpServer.Run(*addr); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
