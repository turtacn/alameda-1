package scores

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/scores/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/scores/types"
)

func NewScoreDAO(config config.Config) types.ScoreDAO {
	return influxdb.NewScoreWithConfig(*config.InfluxDB)
}
