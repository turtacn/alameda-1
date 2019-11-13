package prometheus

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
)

func newLocalControllerRepo() (DaoMetricTypes.ControllerMetricsDAO, error) {

	return NewControllerMetricsWithConfig(
		InternalPromth.Config{
			URL: "http://localhost:9090",
		},
		InternalInflux.Config{
			Address:            "https://localhost:8086",
			Username:           "admin",
			Password:           "adminpass",
			InsecureSkipVerify: true,
		},
		"test-cluster",
	), nil
}

func TestControllerListMetrics(t *testing.T) {
	repo, err := newLocalControllerRepo()
	if err != nil {
		t.Error(err)
	}

	now := time.Now()
	oneHourAgo := time.Now().Add(-1 * time.Minute)
	step := 15 * time.Second

	resp, err := repo.ListMetrics(
		context.TODO(),
		DaoMetricTypes.ListControllerMetricsRequest{
			ObjectMetas: []metadata.ObjectMeta{
				metadata.ObjectMeta{},
			},
			Kind: "Deployment",
			QueryCondition: common.QueryCondition{
				StartTime: &oneHourAgo,
				EndTime:   &now,
				StepTime:  &step,
			},
		},
	)
	if err != nil {
		t.Error(err)
	}
	for _, v := range resp.MetricMap {
		s, err := json.Marshal(v)
		if err != nil {
			t.Error(err)
		}
		t.Logf("resp: %s ", s)
	}

}
