package influxdb

import (
	DaoScore "github.com/containers-ai/alameda/datahub/pkg/dao/score"
	EntityInfluxScore "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/score"
	RepoInfluxScore "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/score"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"github.com/pkg/errors"
)

type influxdbDAO struct {
	config InternalInflux.Config
}

// NewWithConfig New influxdb score dao implement
func NewWithConfig(config InternalInflux.Config) DaoScore.DAO {
	return influxdbDAO{
		config: config,
	}
}

// ListSimulatedScheduingScores Function implementation of score dao
func (dao influxdbDAO) ListSimulatedScheduingScores(request DaoScore.ListRequest) ([]*DaoScore.SimulatedSchedulingScore, error) {

	var (
		err error

		scoreRepository       RepoInfluxScore.SimulatedSchedulingScoreRepository
		influxdbScoreEntities []*EntityInfluxScore.SimulatedSchedulingScoreEntity
		scores                = make([]*DaoScore.SimulatedSchedulingScore, 0)
	)

	scoreRepository = RepoInfluxScore.NewRepositoryWithConfig(dao.config)
	influxdbScoreEntities, err = scoreRepository.ListScoresByRequest(request)
	if err != nil {
		return scores, errors.Wrap(err, "list simulated scheduing scores failed")
	}

	for _, influxdbScoreEntity := range influxdbScoreEntities {

		score := DaoScore.SimulatedSchedulingScore{
			Timestamp: influxdbScoreEntity.Time,
		}

		if scoreBefore := influxdbScoreEntity.ScoreBefore; scoreBefore != nil {
			score.ScoreBefore = *scoreBefore
		}

		if scoreAfter := influxdbScoreEntity.ScoreAfter; scoreAfter != nil {
			score.ScoreAfter = *scoreAfter
		}

		scores = append(scores, &score)
	}

	return scores, nil
}

// CreateSimulatedScheduingScores Function implementation of score dao
func (dao influxdbDAO) CreateSimulatedScheduingScores(scores []*DaoScore.SimulatedSchedulingScore) error {

	var (
		err error

		scoreRepository RepoInfluxScore.SimulatedSchedulingScoreRepository
	)

	scoreRepository = RepoInfluxScore.NewRepositoryWithConfig(dao.config)
	err = scoreRepository.CreateScores(scores)
	if err != nil {
		return errors.Wrap(err, "create simulated scheduing scores failed")
	}

	return nil
}
