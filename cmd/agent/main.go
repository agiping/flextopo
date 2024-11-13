package main

import (
	"context"
	"flextopo/pkg/collector"
	"flextopo/pkg/reporter"
	"flextopo/pkg/utils"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	logger := &utils.SimpleLogger{}

	// 创建 Kubernetes 客户端
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("Failed to create in-cluster config: " + err.Error())
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error("Failed to create clientset: " + err.Error())
		os.Exit(1)
	}

	// 创建定时器
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		// 获取所有节点
		nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logger.Error("Failed to list nodes: " + err.Error())
			<-ticker.C
			continue
		}

		// 对每个节点进行收集和上报
		for _, node := range nodes.Items {
			nodeName := node.Name

			collector, err := collector.NewCollector(nodeName, logger)
			if err != nil {
				logger.Error("Failed to create collector for node " + nodeName + ": " + err.Error())
				continue
			}

			reporter, err := reporter.NewReporter(nodeName, logger)
			if err != nil {
				logger.Error("Failed to create reporter for node " + nodeName + ": " + err.Error())
				continue
			}

			logger.Info("Collecting topology data for node: " + nodeName)
			graph, err := collector.Collect()
			if err != nil {
				logger.Error("Failed to collect topology data for node " + nodeName + ": " + err.Error())
				continue
			}

			logger.Info("Reporting topology data for node: " + nodeName)
			err = reporter.Report(graph)
			if err != nil {
				logger.Error("Failed to report topology data for node " + nodeName + ": " + err.Error())
				continue
			}

			logger.Info("Successfully reported topology data for node " + nodeName)
		}

		<-ticker.C
	}
}
