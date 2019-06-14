package eviction

import "fmt"

type triggerThreshold struct {
	CPU    float64 `mapstructure:"cpu"`
	Memory float64 `mapstructure:"memory"`
}

// Config is eviction configuration
type Config struct {
	CheckCycle              int64            `mapstructure:"check-cycle"`
	Enable                  bool             `mapstructure:"enable"`
	TriggerThreshold        triggerThreshold `mapstructure:"trigger-threshold"`
	PurgeContainerCPUMemory bool             `mapstructure:"purge-container-cpu-memory"`
	PreservationPercentage  float64          `mapstructure:"preservation-percentage"`
}

// NewDefaultConfig returns Config instance
func NewDefaultConfig() Config {
	return Config{
		CheckCycle: 3,
		Enable:     false,
		TriggerThreshold: triggerThreshold{
			CPU:    1,
			Memory: 1,
		},
		PurgeContainerCPUMemory: false,
		PreservationPercentage:  50,
	}
}

func (c *Config) Validate() error {
	if c.TriggerThreshold.CPU <= 0 {
		return fmt.Errorf("Invalid CPU trigger threshold value %v", c.TriggerThreshold.CPU)
	}
	if c.TriggerThreshold.Memory <= 0 {
		return fmt.Errorf("Invalid Memory trigger threshold value %v", c.TriggerThreshold.Memory)
	}
	return nil
}
