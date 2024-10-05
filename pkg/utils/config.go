package utils

import (
	"os"
	"strconv"
)

// Config 表示全局配置
type Config struct {
	CoreGroupSize int
	// 其他配置项
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		// 初始化配置
		coreGroupSize := 8 // 默认值
		if val := os.Getenv("CORE_GROUP_SIZE"); val != "" {
			if size, err := strconv.Atoi(val); err == nil && size > 0 {
				coreGroupSize = size
			}
		}

		config = &Config{
			CoreGroupSize: coreGroupSize,
		}
	}
	return config
}
