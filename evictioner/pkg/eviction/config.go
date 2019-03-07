package eviction

import "fmt"

type triggerThreashold struct {
	CPU    float64 `mapstructure:"cpu"`
	Memory float64 `mapstructure:"memory"`
}

// Config is eviction configuration
type Config struct {
	CheckCycle        int64             `mapstructure:"check-cycle"`
	Enable            bool              `mapstructure:"enable"`
	TriggerThreashold triggerThreashold `mapstructure:"trigger-threashold"`
}

// NewDefaultConfig returns Config instance
func NewDefaultConfig() Config {
	return Config{
		CheckCycle: 3,
		Enable:     false,
		TriggerThreashold: triggerThreashold{
			CPU:    1,
			Memory: 1,
		},
	}
}

func (c *Config) Validate() error {
	if c.TriggerThreashold.CPU <= 0 {
		return fmt.Errorf("Invalid CPU trigger threashold value %v", c.TriggerThreashold.CPU)
	}
	if c.TriggerThreashold.Memory <= 0 {
		return fmt.Errorf("Invalid Memory trigger threashold value %v", c.TriggerThreashold.Memory)
	}
	return nil
}
