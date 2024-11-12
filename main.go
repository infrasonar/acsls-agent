package main

import (
	"log"

	"github.com/infrasonar/go-libagent"
)

func main() {
	// Start collector
	log.Printf("Starting InfraSonar ACSLS Agent Collector v%s\n", version)

	// Initialize random
	libagent.RandInit()

	// Initialize Helper
	libagent.GetHelper()

	// Set-up signal handler
	quit := make(chan bool)
	go libagent.SigHandler(quit)

	// Create Collector
	collector := libagent.NewCollector("acsls", version)

	// Create Asset
	asset := libagent.NewAsset(collector)

	// asset.Kind = "Linux"
	asset.Announce()

	// Create and plan checks
	checkAcsss := libagent.Check{
		Key:             "acsss",
		Collector:       collector,
		Asset:           asset,
		IntervalEnv:     "CHECK_ACSSS_INTERVAL",
		DefaultInterval: 300,
		NoCount:         false,
		SetTimestamp:    false,
		Fn:              CheckAcsss,
	}
	go checkAcsss.Plan(quit)

	// Wait for quit
	<-quit
}
