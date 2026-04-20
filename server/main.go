package main

import (
	"flag"
	"log"
	"os"
	
	"xhc2_for_studying/server/c2"
	"xhc2_for_studying/server/store"
)

func main() {
	defaultAddr := getenv("C2_TO_STUDY_ADDR", ":8024")
	addr := flag.String("addr", defaultAddr, "HTTP listen address")
	flag.Parse()
	
	beaconStore := store.NewBeaconStore()
	taskStore := store.NewServerTaskStore()
	
	httpServer := c2.NewHTTPServer(beaconStore, taskStore)
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
