package v1alpha1

import (
	DaoScore "github.com/containers-ai/alameda/datahub/pkg/dao/score"
	DaoScoreImplInflux "github.com/containers-ai/alameda/datahub/pkg/dao/score/impl/influxdb"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// CreateSimulatedSchedulingScores add simulated scheduling scores to database
func (s *ServiceV1alpha1) CreateSimulatedSchedulingScores(ctx context.Context, in *DatahubV1alpha1.CreateSimulatedSchedulingScoresRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateSimulatedSchedulingScores grpc function: " + AlamedaUtils.InterfaceToString(in))

	var (
		err error

		scoreDAO                            DaoScore.DAO
		daoSimulatedSchedulingScoreEntities = make([]*DaoScore.SimulatedSchedulingScore, 0)
	)

	scoreDAO = DaoScoreImplInflux.NewWithConfig(*s.Config.InfluxDB)

	for _, scoreEntity := range in.GetScores() {

		if scoreEntity == nil {
			continue
		}

		timestamp, _ := ptypes.Timestamp(scoreEntity.GetTime())
		daoSimulatedSchedulingScoreEntity := DaoScore.SimulatedSchedulingScore{
			Timestamp:   timestamp,
			ScoreBefore: float64(scoreEntity.GetScoreBefore()),
			ScoreAfter:  float64(scoreEntity.GetScoreAfter()),
		}
		daoSimulatedSchedulingScoreEntities = append(daoSimulatedSchedulingScoreEntities, &daoSimulatedSchedulingScoreEntity)
	}

	err = scoreDAO.CreateSimulatedScheduingScores(daoSimulatedSchedulingScoreEntities)
	if err != nil {
		scope.Errorf("api CreateSimulatedSchedulingScores failed: %+v", err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// ListSimulatedSchedulingScores list simulated scheduling scores
func (s *ServiceV1alpha1) ListSimulatedSchedulingScores(ctx context.Context, in *DatahubV1alpha1.ListSimulatedSchedulingScoresRequest) (*DatahubV1alpha1.ListSimulatedSchedulingScoresResponse, error) {
	scope.Debug("Request received from ListSimulatedSchedulingScores grpc function: " + AlamedaUtils.InterfaceToString(in))

	var (
		err error

		scoreDAO                          DaoScore.DAO
		scoreDAOListRequest               DaoScore.ListRequest
		scoreDAOSimulatedSchedulingScores = make([]*DaoScore.SimulatedSchedulingScore, 0)

		datahubScores = make([]*DatahubV1alpha1.SimulatedSchedulingScore, 0)
	)

	scoreDAO = DaoScoreImplInflux.NewWithConfig(*s.Config.InfluxDB)

	datahubListSimulatedSchedulingScoresRequestExtended := datahubListSimulatedSchedulingScoresRequestExtended{in}
	scoreDAOListRequest = datahubListSimulatedSchedulingScoresRequestExtended.daoLisRequest()
	scoreDAOSimulatedSchedulingScores, err = scoreDAO.ListSimulatedScheduingScores(scoreDAOListRequest)
	if err != nil {
		scope.Errorf("api ListSimulatedSchedulingScores failed: %v", err)
		return &DatahubV1alpha1.ListSimulatedSchedulingScoresResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			Scores: datahubScores,
		}, nil
	}

	for _, daoSimulatedSchedulingScore := range scoreDAOSimulatedSchedulingScores {

		t, err := ptypes.TimestampProto(daoSimulatedSchedulingScore.Timestamp)
		if err != nil {
			scope.Warnf("api ListSimulatedSchedulingScores warn: time convert failed: %s", err.Error())
		}
		datahubScore := DatahubV1alpha1.SimulatedSchedulingScore{
			Time:        t,
			ScoreBefore: float32(daoSimulatedSchedulingScore.ScoreBefore),
			ScoreAfter:  float32(daoSimulatedSchedulingScore.ScoreAfter),
		}
		datahubScores = append(datahubScores, &datahubScore)
	}

	return &DatahubV1alpha1.ListSimulatedSchedulingScoresResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Scores: datahubScores,
	}, nil
}
