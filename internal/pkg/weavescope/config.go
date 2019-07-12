package weavescope

const (
	defaultURL = "https://weavescope:4041"
)

// Configuration of weave scope
type Config struct {
	URL string `mapstructure:"url"`
}

// Provide default configuration for weave scope
func NewDefaultConfig() *Config {
	var config = Config{
		URL: defaultURL,
	}
	return &config
}

// Confirm the configuration is validated
func (c *Config) Validate() error {
	return nil
}
