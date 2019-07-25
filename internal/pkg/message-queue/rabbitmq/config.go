package rabbitmq

const (
	DEFAULT_URL                       = "amqp://admin:adminpass@rabbitmq.alameda.svc.cluster.local:5672"
	DEFAULT_PUBLISH_RETRY_TIME        = 3
	DEFAULT_PUBLISH_RETRY_INTERVAL_MS = 500
)

// Configuration of weave scope
type Config struct {
	URL   string `mapstructure:"url"`
	Retry *Retry `mapstructure:"retry"`
}

type Retry struct {
	PublishTime       int   `mapstructure:"publishTime"`
	PublishIntervalMS int64 `mapstructure:"publishIntervalMs"`
}

// Provide default configuration for weave scope
func NewDefaultConfig() *Config {
	var config = Config{
		URL: DEFAULT_URL,
		Retry: &Retry{
			PublishTime:       DEFAULT_PUBLISH_RETRY_TIME,
			PublishIntervalMS: DEFAULT_PUBLISH_RETRY_INTERVAL_MS,
		},
	}
	return &config
}

// Confirm the configuration is validated
func (c *Config) Validate() error {
	return nil
}
