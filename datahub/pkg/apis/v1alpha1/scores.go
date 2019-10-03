package v1alpha1

import (
	DaoScore "github.com/containers-ai/alameda/datahub/pkg/dao/score"
	DaoScoreImplInflux "github.com/containers-ai/alameda/datahub/pkg/dao/score/impl/influxdb"
	RequestExtend "github.com/containers-ai/alameda/datahub/pkg/formatextension/requests"
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

	scoreDAO := DaoScoreImplInflux.NewWithConfig(*s.Config.InfluxDB)

	daoSimulatedSchedulingScoreEntities := make([]*DaoScore.SimulatedSchedulingScore, 0)
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

	err := scoreDAO.CreateSimulatedScheduingScores(daoSimulatedSchedulingScoreEntities)
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

	scoreDAO := DaoScoreImplInflux.NewWithConfig(*s.Config.InfluxDB)

	requestExt := RequestExtend.ListSimulatedSchedulingScoresRequestExtended{Request: in}
	scoreDAOListRequest := requestExt.ProduceRequest()
	scoreDAOSimulatedSchedulingScores, err := scoreDAO.ListSimulatedScheduingScores(scoreDAOListRequest)
	if err != nil {
		scope.Errorf("api ListSimulatedSchedulingScores failed: %v", err)
		return &DatahubV1alpha1.ListSimulatedSchedulingScoresResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			Scores: make([]*DatahubV1alpha1.SimulatedSchedulingScore, 0),
		}, nil
	}

	datahubScores := make([]*DatahubV1alpha1.SimulatedSchedulingScore, 0)
	for _, daoSimulatedSchedulingScore := range scoreDAOSimulatedSchedulingScores {

		t, err := ptypes.TimestampProto(daoSimulatedSchedulingScore.Timestamp)
		if err != nil {
			scope.Warnf("api ListSimulatedSchedulingScores warn: time convert failed: %s", err.Error())
		}
		datahubScore := DatahubV1alpha1.SimulatedSchedulingScore{
			Time:        t,
			ScoreBefore: daoSimulatedSchedulingScore.ScoreBefore,
			ScoreAfter:  daoSimulatedSchedulingScore.ScoreAfter,
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
