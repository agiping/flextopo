package utils

import (
	"log"
	"strconv"
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
	CPUID    int
	CoreID   int
	SocketID int
	NodeID   int
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
