package main

import (
	"context"
	"log"
	
	"xhc2_for_studying/implant/client"
	"xhc2_for_studying/implant/config"
	implantRuntime "xhc2_for_studying/implant/runtime"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	
	httpClient, err := client.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	
	runner, err := implantRuntime.NewRunner(cfg, httpClient)
	if err != nil {
		log.Fatal(err)
	}
	
	if err := runner.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
