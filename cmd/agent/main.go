package main

import (
	"context"
	"flextopo/pkg/collector"
	"flextopo/pkg/reporter"
	"flextopo/pkg/utils"
	"os"
	"sync"
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

		// 使用 WaitGroup 等待所有协程完成
		var wg sync.WaitGroup
		// 限制并发数量为10个协程
		semaphore := make(chan struct{}, 10)

		// 对每个节点启动一个协程进行处理
		for _, node := range nodes.Items {
			wg.Add(1)
			semaphore <- struct{}{} // 获取信号量

			go func(nodeName string) {
				defer wg.Done()
				defer func() { <-semaphore }() // 释放信号量

				collector, err := collector.NewCollector(nodeName, logger)
				if err != nil {
					logger.Error("Failed to create collector for node " + nodeName + ": " + err.Error())
					return
				}

				reporter, err := reporter.NewReporter(nodeName, logger)
				if err != nil {
					logger.Error("Failed to create reporter for node " + nodeName + ": " + err.Error())
					return
				}

				logger.Info("Collecting topology data for node: " + nodeName)
				graph, err := collector.Collect()
				if err != nil {
					logger.Error("Failed to collect topology data for node " + nodeName + ": " + err.Error())
					return
				}

				logger.Info("Reporting topology data for node: " + nodeName)
				err = reporter.Report(graph)
				if err != nil {
					logger.Error("Failed to report topology data for node " + nodeName + ": " + err.Error())
					return
				}

				logger.Info("Successfully reported topology data for node " + nodeName)
			}(node.Name)
		}

		// 等待所有协程完成
		wg.Wait()
		<-ticker.C
	}
}
