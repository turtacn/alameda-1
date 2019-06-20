package eviction

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewPodReplicaStatus(t *testing.T) {

	now := meta_v1.NewTime(time.Now())

	type have struct {
		pods           []core_v1.Pod
		replicasCount  int32
		maxUnavailable string
	}

	type testCase struct {
		have have
		want podReplicaStatus
	}

	successTestCases := []testCase{

		testCase{
			have: have{
				pods: []core_v1.Pod{
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
				},
				replicasCount:  5,
				maxUnavailable: "25%",
			},
			want: podReplicaStatus{
				preservedPodCount: 3,
				runningPodCount:   5,
				evictedPodCount:   0,
			},
		},
		testCase{
			have: have{
				pods: []core_v1.Pod{
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
				},
				replicasCount:  5,
				maxUnavailable: "200%",
			},
			want: podReplicaStatus{
				preservedPodCount: 0,
				runningPodCount:   5,
				evictedPodCount:   0,
			},
		},
		testCase{
			have: have{
				pods: []core_v1.Pod{
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
				},
				replicasCount:  5,
				maxUnavailable: "10",
			},
			want: podReplicaStatus{
				preservedPodCount: 0,
				runningPodCount:   5,
				evictedPodCount:   0,
			},
		},
		testCase{
			have: have{
				pods: []core_v1.Pod{
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: &now}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: &now}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
					core_v1.Pod{ObjectMeta: meta_v1.ObjectMeta{DeletionTimestamp: nil}, Status: core_v1.PodStatus{Phase: core_v1.PodRunning}},
				},
				replicasCount:  5,
				maxUnavailable: "25%",
			},
			want: podReplicaStatus{
				preservedPodCount: 3,
				runningPodCount:   3,
				evictedPodCount:   0,
			},
		},
	}

	assert := assert.New(t)
	for _, testCase := range successTestCases {

		actual, err := NewPodReplicaStatus(testCase.have.pods, testCase.have.replicasCount, testCase.have.maxUnavailable)
		assert.Nil(err)
		assert.Equal(testCase.want, actual)
	}
}
