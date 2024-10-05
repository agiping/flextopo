package main

import (
	"flextopo/pkg/collector"
	"flextopo/pkg/reporter"
	"flextopo/pkg/utils"
	"os"
	"time"
)

func main() {
	logger := &utils.SimpleLogger{}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		logger.Error("NODE_NAME environment variable not set")
		os.Exit(1)
	}

	collector, err := collector.NewCollector(nodeName, logger)
	if err != nil {
		logger.Error("Failed to create collector: " + err.Error())
		os.Exit(1)
	}

	reporter, err := reporter.NewReporter(nodeName, logger)
	if err != nil {
		logger.Error("Failed to create reporter: " + err.Error())
		os.Exit(1)
	}

	// 定时收集和上报拓扑信息
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		graph, err := collector.Collect()
		if err != nil {
			logger.Error("Failed to collect topology data: " + err.Error())
			continue
		}

		err = reporter.Report(graph)
		if err != nil {
			logger.Error("Failed to report topology data: " + err.Error())
			continue
		}

		logger.Info("Successfully reported topology data")

		<-ticker.C
	}
}
