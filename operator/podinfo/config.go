package podinfo

type Config struct {
	LabelsFile string `mapstructure:"labelsFile"`
}

func NewConfig() *Config {
	return &Config{
		LabelsFile: "/etc/podinfo/labels",
	}
}
