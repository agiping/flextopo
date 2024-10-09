package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Atoi 将字符串转换为整数，包含错误处理
func Atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("Failed to convert string to int: %v", err)
		return 0
	}
	return i
}

// CPUInfo 表示单个 CPU 的信息
type CPUInfo struct {
	CPUID      int
	CoreID     int
	SocketID   int
	NumaNodeID int
}

// Logger 接口定义了日志方法
type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

// SimpleLogger 是一个简单的 Logger 实现
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string) {
	log.Printf("[INFO] %s", msg)
}

func (l *SimpleLogger) Warn(msg string) {
	log.Printf("[WARN] %s", msg)
}

func (l *SimpleLogger) Error(msg string) {
	log.Printf("[ERROR] %s", msg)
}

// ParseCPUList 解析 CPU 列表字符串，返回 CPU 核心编号的整数切片
func ParseCPUList(cpuListStr string) ([]int, error) {
	var cpuCores []int
	segments := strings.Split(cpuListStr, ",")
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		if strings.Contains(segment, "-") {
			// 处理范围，例如 "0-3"
			bounds := strings.Split(segment, "-")
			if len(bounds) != 2 {
				return nil, fmt.Errorf("invalid CPU range: %s", segment)
			}
			start, err := strconv.Atoi(bounds[0])
			if err != nil {
				return nil, fmt.Errorf("invalid CPU number: %s", bounds[0])
			}
			end, err := strconv.Atoi(bounds[1])
			if err != nil {
				return nil, fmt.Errorf("invalid CPU number: %s", bounds[1])
			}
			if start > end {
				return nil, fmt.Errorf("invalid CPU range: %s", segment)
			}
			for i := start; i <= end; i++ {
				cpuCores = append(cpuCores, i)
			}
		} else {
			// 处理单个数字，例如 "5"
			cpuNum, err := strconv.Atoi(segment)
			if err != nil {
				return nil, fmt.Errorf("invalid CPU number: %s", segment)
			}
			cpuCores = append(cpuCores, cpuNum)
		}
	}
	return cpuCores, nil
}

// ParseLSCPUOutput 解析 lscpu 命令的输出
func ParseLSCPUOutput(output string) []CPUInfo {
	lines := strings.Split(output, "\n")
	var cpuInfos []CPUInfo
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 4 {
			continue
		}
		cpuID := Atoi(fields[0])
		coreID := Atoi(fields[1])
		socketID := Atoi(fields[2])
		numaNodeID := Atoi(fields[3])

		cpuInfo := CPUInfo{
			CPUID:      cpuID,
			CoreID:     coreID,
			SocketID:   socketID,
			NumaNodeID: numaNodeID,
		}
		cpuInfos = append(cpuInfos, cpuInfo)
	}
	return cpuInfos
}

func WaitForInput() {
	fmt.Print("waiting input...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
