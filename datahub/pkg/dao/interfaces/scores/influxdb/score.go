package influxdb

import (
	EntityInfluxScore "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/scores"
	DaoScore "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/scores/types"
	RepoInfluxScore "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/scores"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"github.com/pkg/errors"
)

type Score struct {
	InfluxDBConfig InternalInflux.Config
}

// NewWithConfig New influxdb score dao implement
func NewScoreWithConfig(config InternalInflux.Config) DaoScore.ScoreDAO {
	return &Score{InfluxDBConfig: config}
}

// ListSimulatedSchedulingScores Function implementation of score dao
func (s *Score) ListSimulatedSchedulingScores(request DaoScore.ListRequest) ([]*DaoScore.SimulatedSchedulingScore, error) {

	var (
		err error

		scoreRepository       RepoInfluxScore.SimulatedSchedulingScoreRepository
		influxdbScoreEntities []*EntityInfluxScore.SimulatedSchedulingScoreEntity
		scores                = make([]*DaoScore.SimulatedSchedulingScore, 0)
	)

	scoreRepository = RepoInfluxScore.NewRepositoryWithConfig(s.InfluxDBConfig)
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

// CreateSimulatedSchedulingScores Function implementation of score dao
func (s *Score) CreateSimulatedSchedulingScores(scores []*DaoScore.SimulatedSchedulingScore) error {

	var (
		err error

		scoreRepository RepoInfluxScore.SimulatedSchedulingScoreRepository
	)

	scoreRepository = RepoInfluxScore.NewRepositoryWithConfig(s.InfluxDBConfig)
	err = scoreRepository.CreateScores(scores)
	if err != nil {
		return errors.Wrap(err, "create simulated scheduing scores failed")
	}

	return nil
}
