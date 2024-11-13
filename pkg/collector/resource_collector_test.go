package collector

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCollectResourceInfo(t *testing.T) {
	// 创建模拟的 pods
	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "high-p-1000-1-card-6cd96cdd49-z246h",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "low-p-500-2-card-6898d46d9b-28vr2",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "low-p-500-1-card-6898d46d9b-4hc8m",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "low-p-500-1-card-6898d46d9b-dfgdfg",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "low-p-500-1-card-6898d46d9b-dfvsdv",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "low-p-500-1-card-6898d46d9b-dsvada",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "low-p-500-1-card-6898d46d9b-derada",
				},
			},
		},
	}

	gpuPods := []corev1.Pod{}
	// totlalGPUs := 0
	for _, pod := range pods.Items {
		if !strings.Contains(pod.Name, "-card") {
			continue
		}
		gpuPods = append(gpuPods, pod)
		// numCardsStr := strings.Split(pod.Name, "-")[3]
		// numCards, err := strconv.Atoi(numCardsStr)
		// if err != nil {
		// 	rc.logger.Error("Failed to convert numCardsStr to int: " + err.Error())
		// 	continue
		// }
		// totlalGPUs += numCards
	}

	totalCoresUsed := 0
	for _, pod := range gpuPods {
		// 从 pod 名称中解析出卡数
		nameParts := strings.Split(pod.Name, "-")
		if len(nameParts) < 4 {
			//rc.logger.Error("Pod name format is incorrect: " + pod.Name)
			continue
		}
		numCardsStr := nameParts[3]
		numCards, err := strconv.Atoi(numCardsStr)
		if err != nil {
			//rc.logger.Error("Failed to convert numCardsStr to int: " + err.Error())
			continue
		}

		numCoresNeeded := numCards * 7
		cpuCores := []int{}
		coreCount := 0

		for coreCount < numCoresNeeded {
			if totalCoresUsed > 63 {
				//rc.logger.Error("Not enough CPU cores to allocate to pod " + pod.Name)
				break
			}
			if totalCoresUsed%8 == 7 {
				// 跳过每个 CPU 组的最后一个核心
				totalCoresUsed++
				continue
			}
			cpuCores = append(cpuCores, totalCoresUsed)
			totalCoresUsed++
			coreCount++
		}

		if coreCount < numCoresNeeded {
			//rc.logger.Error("Insufficient CPU cores allocated to pod " + pod.Name)
			continue
		}

		fmt.Println("CPU cores of pod " + pod.Name + ": " + fmt.Sprintf("%v", cpuCores))
		//graph.UpdateCPUUsage(pod.Name, cpuCores)
	}
}
