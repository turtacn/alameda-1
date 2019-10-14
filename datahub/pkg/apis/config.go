package apis

// Configuration of APIs
type Config struct {
	Metrics *MetricsConfig `mapstructure:"metrics"`
}

// Configuration of metrics related APIs
type MetricsConfig struct {
	Source string `mapstructure:"source"`
	Target string `mapstructure:"target"`
}

// Provide default configuration for APIs
func NewDefaultConfig() *Config {
	var config = Config{
		Metrics: &MetricsConfig{
			Source: "prometheus",
			Target: "influxdb",
		},
	}
	return &config
}

// Confirm the APIs configuration is validated
func (c *Config) Validate() error {
	return nil
}
