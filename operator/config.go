package operator

import (
	"errors"

	aiserver "github.com/containers-ai/alameda/operator/ai-server"
	datahub "github.com/containers-ai/alameda/operator/datahub"
	"github.com/containers-ai/alameda/operator/pkg/utils/log"
	"github.com/containers-ai/alameda/operator/services/grpc"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Config defines configurations
type Config struct {
	GRPC     *grpc.Config     `mapstructure:"gRPC"`
	Log      *log.Config      `mapstructure:"log"`
	Datahub  *datahub.Config  `mapstructure:"datahub"`
	AIServer *aiserver.Config `mapstructure:"ai-server"`
	Manager  manager.Manager
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

	c.GRPC = grpc.NewConfig()
	c.Log = log.NewConfig()
	c.AIServer = aiserver.NewConfig()
	c.Datahub = datahub.NewConfig()
}

func (c Config) Validate() error {

	var err error

	err = c.GRPC.Validate()
	if err != nil {
		return errors.New("server config validate failed: " + err.Error())
	}

	return nil
}
