/*
Copyright 2019 The Alameda Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"testing"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestStorageAlamedaRecommendation(t *testing.T) {
	key := types.NamespacedName{
		Name:      "foo",
		Namespace: "default",
	}
	created := &AlamedaRecommendation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: AlamedaRecommendationSpec{
			Containers: []AlamedaContainer{
				AlamedaContainer{
					Name: "nginx",
				},
			},
		},
	}
	g := gomega.NewGomegaWithT(t)

	// Test Create
	fetched := &AlamedaRecommendation{}
	g.Expect(c.Create(context.TODO(), created)).NotTo(gomega.HaveOccurred())

	g.Expect(c.Get(context.TODO(), key, fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fetched).To(gomega.Equal(created))

	// Test Updating the Labels
	updated := fetched.DeepCopy()
	updated.Labels = map[string]string{"hello": "world"}
	g.Expect(c.Update(context.TODO(), updated)).NotTo(gomega.HaveOccurred())

	g.Expect(c.Get(context.TODO(), key, fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fetched).To(gomega.Equal(updated))

	// Test Delete
	g.Expect(c.Delete(context.TODO(), fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(c.Get(context.TODO(), key, fetched)).To(gomega.HaveOccurred())
}

func TestGenAlamedaUpdation(t *testing.T) {

	type input struct {
		AlamedaRecommendation *AlamedaRecommendation
		AlamedaScalerName     string
		AlamedaControllerName string
		AlamedaControllerType AlamedaControllerType
	}

	tests := []struct {
		have input
		want *AlamedaUpdation
	}{
		{
			have: input{
				AlamedaRecommendation: &AlamedaRecommendation{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: AlamedaRecommendationSpec{
						Containers: []AlamedaContainer{
							AlamedaContainer{
								Name: "nginx",
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU: resource.MustParse("1"),
									},
								},
							},
						},
					},
				},
				AlamedaScalerName:     "alamedascaler-test",
				AlamedaControllerName: "nginx-deployment",
				AlamedaControllerType: DeploymentController,
			},
			want: &AlamedaUpdation{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
					Labels:    GenerateMonitoringAlamedaScalerAlamedaControllerIdentityLabels("alamedascaler-test", "nginx-deployment", DeploymentController),
				},
				Spec: AlamedaUpdationSpec{
					AlamedaRecommendationSpec{
						Containers: []AlamedaContainer{
							AlamedaContainer{
								Name: "nginx",
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU: resource.MustParse("1"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	g := gomega.NewGomegaWithT(t)
	for _, test := range tests {
		t.Logf("ala updation %v", test.want)
		g.Expect(*test.have.AlamedaRecommendation.GenAlamedaUpdation(test.have.AlamedaScalerName, test.have.AlamedaControllerName, test.have.AlamedaControllerType)).To(gomega.Equal(*test.want))
	}

}
