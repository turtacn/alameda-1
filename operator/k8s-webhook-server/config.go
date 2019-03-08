package k8swhsrv

import "fmt"

type secret struct {
	Namespace string `mapstructure:"namespace"`
	Name      string `mapstructure:"name"`
}

type service struct {
	Namespace string `mapstructure:"namespace"`
	Name      string `mapstructure:"name"`
}

type Config struct {
	Port                        int32   `mapstructure:"port"`
	CertDir                     string  `mapstructure:"cert-dir"`
	Service                     service `mapstructure:"service"`
	Secret                      secret  `mapstructure:"secret"`
	ValidatingWebhookConfigName string  `mapstructure:"validating-webhook-config-name"`
	MutatingWebhookConfigName   string  `mapstructure:"mutating-webhook-config-name"`
}

func NewConfig() *Config {

	c := Config{
		Port:    443,
		CertDir: "/k8s-webhook-server/cert/",
		Service: service{
			Namespace: "alameda",
			Name:      "operator-admission-service",
		},
		Secret: secret{
			Namespace: "alameda",
			Name:      "operator-admission-secret",
		},
		ValidatingWebhookConfigName: "operator-k8s-admission-validation",
		MutatingWebhookConfigName:   "operator-k8s-admission-mutation",
	}
	return &c
}

func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("K8S webhook server port %v is not valid", c.Port)
	}
	return nil
}
