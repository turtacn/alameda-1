package eviction

// Config is eviction configuration
type Config struct {
	CheckCycle int64 `mapstructure:"check-cycle"`
}

const (
	defaultCheckCycle = 3
)

// NewDefaultConfig returns Config instance
func NewDefaultConfig() Config {
	return Config{
		CheckCycle: defaultCheckCycle,
	}
}

func (c *Config) Validate() error {
	return nil
}
