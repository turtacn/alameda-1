package types

import (
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	"time"
)

// DAO DAO interface of score
type ScoreDAO interface {
	ListSimulatedSchedulingScores(ListRequest) ([]*SimulatedSchedulingScore, error)
	CreateSimulatedSchedulingScores([]*SimulatedSchedulingScore) error
}

// SimulatedSchedulingScore Score entity in dao level
type SimulatedSchedulingScore struct {
	Timestamp   time.Time
	ScoreBefore float64
	ScoreAfter  float64
}

// ListRequest Request argument for list api.
type ListRequest struct {
	DBCommon.QueryCondition
}
