package types

import (
	"testing"
	"time"

	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/stretchr/testify/assert"
)

func TestAPPMerge(t *testing.T) {

	type testCase struct {
		have [2]AppMetric
		want AppMetric
	}

	var (
		time1 = time.Now()
	)
	testCases := []testCase{
		testCase{
			have: [2]AppMetric{
				AppMetric{
					ObjectMeta: metadata.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
				},
				AppMetric{
					ObjectMeta: metadata.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
					Metrics: map[enumconv.MetricKind][]types.Sample{
						enumconv.MetricTypeCPUUsageSecondsPercentage: []types.Sample{
							types.Sample{
								Timestamp: time1,
								Value:     "1",
							},
						},
					},
				},
			},
			want: AppMetric{
				ObjectMeta: metadata.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Metrics: map[enumconv.MetricKind][]types.Sample{
					enumconv.MetricTypeCPUUsageSecondsPercentage: []types.Sample{
						types.Sample{
							Timestamp: time1,
							Value:     "1",
						},
					},
				},
			},
		},
	}

	assert := assert.New(t)
	for _, tc := range testCases {
		a1 := tc.have[0]
		a2 := tc.have[1]
		a1.Merge(&a2)
		actual := a1
		assert.Equal(tc.want, actual)
	}
}
