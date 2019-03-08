package operator

import (
	datahub "github.com/containers-ai/alameda/operator/datahub"
	"github.com/containers-ai/alameda/operator/k8s-webhook-server"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Config defines configurations
type Config struct {
	Log              *log.Config                `mapstructure:"log"`
	Datahub          *datahub.Config            `mapstructure:"datahub"`
	K8SWebhookServer *k8swhsrv.Config `mapstructure:"k8s-webhook-server"`
	Manager          manager.Manager
}

// NewConfig returns Config objecdt
func NewConfig(manager manager.Manager) Config {

	c := Config{
		Manager: manager,
	}
	c.init()

	return c
}

func (c *Config) init() {

	defaultLogConfig := log.NewDefaultConfig()

	c.Log = &defaultLogConfig
	c.Datahub = datahub.NewConfig()
	c.K8SWebhookServer = k8swhsrv.NewConfig()
}

func (c Config) Validate() error {

	return nil
}
