package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
)

// Atoi converts a string to an integer, including error handling
func Atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("Failed to convert string to int: %v", err)
		return 0
	}
	return i
}

// CPUInfo represents information for a single CPU
type CPUInfo struct {
	CPUID      int
	CoreID     int
	SocketID   int
	NumaNodeID int
}

// Logger interface defines logging methods
type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// SimpleLogger implements the Logger interface
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string) {
	log.Printf("[INFO] %s", l.formatLog(msg))
}

func (l *SimpleLogger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] %s", l.formatLog(fmt.Sprintf(format, args...)))
}

func (l *SimpleLogger) Warn(msg string) {
	log.Printf("[WARN] %s", l.formatLog(msg))
}

func (l *SimpleLogger) Warnf(format string, args ...interface{}) {
	log.Printf("[WARN] %s", l.formatLog(fmt.Sprintf(format, args...)))
}

func (l *SimpleLogger) Error(msg string) {
	log.Printf("[ERROR] %s", l.formatLog(msg))
}

func (l *SimpleLogger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] %s", l.formatLog(fmt.Sprintf(format, args...)))
}

func (l *SimpleLogger) formatLog(msg string) string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	} else {
		// Get only the file name
		if lastSlash := strings.LastIndex(file, "/"); lastSlash != -1 {
			file = file[lastSlash+1:]
		}
	}
	return fmt.Sprintf("%s:%d %s", file, line, msg)
}

// ParseCPUList parses a CPU list string and returns a slice of CPU core numbers as integers
func ParseCPUList(cpuListStr string) ([]int, error) {
	cpuCores := make([]int, 0)
	if cpuListStr == "" {
		return cpuCores, nil // Return empty slice directly
	}
	segments := strings.Split(cpuListStr, ",")
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		if strings.Contains(segment, "-") {
			// Handle range, e.g., "0-3"
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
			// Handle single number, e.g., "5"
			cpuNum, err := strconv.Atoi(segment)
			if err != nil {
				return nil, fmt.Errorf("invalid CPU number: %s", segment)
			}
			cpuCores = append(cpuCores, cpuNum)
		}
	}
	return cpuCores, nil
}

// ParseLSCPUOutput parses the output of the lscpu command
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

func PrettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}
