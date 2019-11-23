package eviction

import (
	"testing"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestBuildTriggerThreshold(t *testing.T) {

	type testCase struct {
		have autoscalingv1alpha1.AlamedaScaler
		want triggerThreshold
	}

	successTestCases := []testCase{
		testCase{
			have: autoscalingv1alpha1.AlamedaScaler{
				Spec: autoscalingv1alpha1.AlamedaScalerSpec{
					ScalingTool: autoscalingv1alpha1.ScalingToolSpec{
						Type: autoscalingv1alpha1.ScalingToolTypeVPA,
						ExecutionStrategy: &autoscalingv1alpha1.ExecutionStrategy{
							TriggerThreshold: &autoscalingv1alpha1.TriggerThreshold{
								CPU:    "20%",
								Memory: "20%",
							},
						},
					},
				},
			},
			want: triggerThreshold{
				CPU:    float64(20),
				Memory: float64(20),
			},
		},
		testCase{
			have: autoscalingv1alpha1.AlamedaScaler{
				Spec: autoscalingv1alpha1.AlamedaScalerSpec{
					ScalingTool: autoscalingv1alpha1.ScalingToolSpec{
						Type: autoscalingv1alpha1.ScalingToolTypeVPA,
						ExecutionStrategy: &autoscalingv1alpha1.ExecutionStrategy{
							TriggerThreshold: &autoscalingv1alpha1.TriggerThreshold{
								CPU:    "20.5%",
								Memory: "20.5%",
							},
						},
					},
				},
			},
			want: triggerThreshold{
				CPU:    float64(20.5),
				Memory: float64(20.5),
			},
		},
	}

	assert := assert.New(t)
	for _, testCase := range successTestCases {
		c := controllerRecommendationInfo{alamedaScaler: &testCase.have}
		tt, err := c.buildTriggerThreshold()

		assert.Nil(err)
		assert.Equal(testCase.want, tt)
	}

}
