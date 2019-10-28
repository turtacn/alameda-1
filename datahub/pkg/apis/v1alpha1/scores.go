package v1alpha1

import (
	DaoScore "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/scores"
	DaoScoreTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/scores/types"
	RequestExtend "github.com/containers-ai/alameda/datahub/pkg/formatconversion/requests"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiScores "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/scores"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// CreateSimulatedSchedulingScores add simulated scheduling scores to database
func (s *ServiceV1alpha1) CreateSimulatedSchedulingScores(ctx context.Context, in *ApiScores.CreateSimulatedSchedulingScoresRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateSimulatedSchedulingScores grpc function: " + AlamedaUtils.InterfaceToString(in))

	scoreDAO := DaoScore.NewScoreDAO(*s.Config)

	daoSimulatedSchedulingScoreEntities := make([]*DaoScoreTypes.SimulatedSchedulingScore, 0)
	for _, scoreEntity := range in.GetScores() {
		if scoreEntity == nil {
			continue
		}

		timestamp, _ := ptypes.Timestamp(scoreEntity.GetTime())
		daoSimulatedSchedulingScoreEntity := DaoScoreTypes.SimulatedSchedulingScore{
			Timestamp:   timestamp,
			ScoreBefore: float64(scoreEntity.GetScoreBefore()),
			ScoreAfter:  float64(scoreEntity.GetScoreAfter()),
		}
		daoSimulatedSchedulingScoreEntities = append(daoSimulatedSchedulingScoreEntities, &daoSimulatedSchedulingScoreEntity)
	}

	err := scoreDAO.CreateSimulatedSchedulingScores(daoSimulatedSchedulingScoreEntities)
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
func (s *ServiceV1alpha1) ListSimulatedSchedulingScores(ctx context.Context, in *ApiScores.ListSimulatedSchedulingScoresRequest) (*ApiScores.ListSimulatedSchedulingScoresResponse, error) {
	scope.Debug("Request received from ListSimulatedSchedulingScores grpc function: " + AlamedaUtils.InterfaceToString(in))

	scoreDAO := DaoScore.NewScoreDAO(*s.Config)

	requestExt := RequestExtend.ListSimulatedSchedulingScoresRequestExtended{Request: in}
	scoreDAOListRequest := requestExt.ProduceRequest()
	scoreDAOSimulatedSchedulingScores, err := scoreDAO.ListSimulatedSchedulingScores(scoreDAOListRequest)
	if err != nil {
		scope.Errorf("api ListSimulatedSchedulingScores failed: %v", err)
		return &ApiScores.ListSimulatedSchedulingScoresResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			Scores: make([]*ApiScores.SimulatedSchedulingScore, 0),
		}, nil
	}

	datahubScores := make([]*ApiScores.SimulatedSchedulingScore, 0)
	for _, daoSimulatedSchedulingScore := range scoreDAOSimulatedSchedulingScores {

		t, err := ptypes.TimestampProto(daoSimulatedSchedulingScore.Timestamp)
		if err != nil {
			scope.Warnf("api ListSimulatedSchedulingScores warn: time convert failed: %s", err.Error())
		}
		datahubScore := ApiScores.SimulatedSchedulingScore{
			Time:        t,
			ScoreBefore: daoSimulatedSchedulingScore.ScoreBefore,
			ScoreAfter:  daoSimulatedSchedulingScore.ScoreAfter,
		}
		datahubScores = append(datahubScores, &datahubScore)
	}

	return &ApiScores.ListSimulatedSchedulingScoresResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Scores: datahubScores,
	}, nil
}
