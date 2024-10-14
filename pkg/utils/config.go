package utils

import (
	"os"
	"strconv"
)

// Config is the global configuration
type Config struct {
	CoreGroupSize int
	// other configurations
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		// initialize config
		coreGroupSize := 8 // default value
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
