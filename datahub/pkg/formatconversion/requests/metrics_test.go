package requests

import (
	"testing"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"

	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
)

func TestNormalizeListMetricsRequestTimeRange(t *testing.T) {

	type testCaseHave struct {
		q      ApiCommon.QueryCondition
		dbType MetricsDBType
	}
	type testCase struct {
		have testCaseHave
		want ApiCommon.QueryCondition
	}

	st := timestamp.Timestamp{Seconds: 1575250000}
	et := timestamp.Timestamp{Seconds: 1575260000}
	s30 := duration.Duration{Seconds: 30}
	s3600 := duration.Duration{Seconds: 3600}

	g30WantSt := timestamp.Timestamp{Seconds: 1575249990}
	g30WantPrometheusEt := timestamp.Timestamp{Seconds: 1575259980}
	g30WantInfluxdbEt := timestamp.Timestamp{Seconds: 1575259979}

	g3600WantSt := timestamp.Timestamp{Seconds: 1575248400}
	g3600WantPrometheusEt := timestamp.Timestamp{Seconds: 1575259200}
	g3600WantInfluxdbEt := timestamp.Timestamp{Seconds: 1575259199}

	testCases := []testCase{
		testCase{
			have: testCaseHave{
				dbType: MetricsDBTypePromethues,
				q: ApiCommon.QueryCondition{
					TimeRange: &ApiCommon.TimeRange{
						StartTime: &st,
						EndTime:   &et,
						Step:      &s30,
					},
				},
			},
			want: ApiCommon.QueryCondition{
				TimeRange: &ApiCommon.TimeRange{
					StartTime: &g30WantSt,
					EndTime:   &g30WantPrometheusEt,
					Step:      &s30,
				},
			},
		},
		testCase{
			have: testCaseHave{
				dbType: MetricsDBTypeInfluxdb,
				q: ApiCommon.QueryCondition{
					TimeRange: &ApiCommon.TimeRange{
						StartTime: &st,
						EndTime:   &et,
						Step:      &s30,
					},
				},
			},
			want: ApiCommon.QueryCondition{
				TimeRange: &ApiCommon.TimeRange{
					StartTime: &g30WantSt,
					EndTime:   &g30WantInfluxdbEt,
					Step:      &s30,
				},
			},
		},
		testCase{
			have: testCaseHave{
				dbType: MetricsDBTypePromethues,
				q: ApiCommon.QueryCondition{
					TimeRange: &ApiCommon.TimeRange{
						StartTime: &st,
						EndTime:   &et,
						Step:      &s3600,
					},
				},
			},
			want: ApiCommon.QueryCondition{
				TimeRange: &ApiCommon.TimeRange{
					StartTime: &g3600WantSt,
					EndTime:   &g3600WantPrometheusEt,
					Step:      &s3600,
				},
			},
		},
		testCase{
			have: testCaseHave{
				dbType: MetricsDBTypeInfluxdb,
				q: ApiCommon.QueryCondition{
					TimeRange: &ApiCommon.TimeRange{
						StartTime: &st,
						EndTime:   &et,
						Step:      &s3600,
					},
				},
			},
			want: ApiCommon.QueryCondition{
				TimeRange: &ApiCommon.TimeRange{
					StartTime: &g3600WantSt,
					EndTime:   &g3600WantInfluxdbEt,
					Step:      &s3600,
				},
			},
		},
	}

	assert := assert.New(t)
	for i, testCase := range testCases {
		actual := normalizeListMetricsRequestQueryConditionWthMetricsDBType(testCase.have.q, testCase.have.dbType)
		expect := testCase.want
		assert.EqualValues(expect, actual, "Test case #%d", i)
		t.Logf("Test case #%d: %+v", i, actual)
	}
}
