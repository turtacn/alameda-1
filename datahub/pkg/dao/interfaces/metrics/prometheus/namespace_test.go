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

func newLocalRepo() DaoMetricTypes.NamespaceMetricsDAO {
	return NewNamespaceMetricsWithConfig(
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
	)
}

func TestListNamespaceMetrics(t *testing.T) {
	repo := newLocalRepo()

	now := time.Now()
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	step := 15 * time.Minute

	resp, err := repo.ListMetrics(
		context.Background(),
		DaoMetricTypes.ListNamespaceMetricsRequest{
			ObjectMetas: []metadata.ObjectMeta{
				metadata.ObjectMeta{Name: "default"},
				metadata.ObjectMeta{Name: "federatorai"},
			},
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
