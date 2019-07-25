# Alameda Operator
Alameda Operator manages the list of pods that need resource predicting, resource recommendating and pod autoscaling.

# Requirements for development

*   golang 1.12
*   [kubebuilder 1.0.8](#https://github.com/kubernetes-sigs/kubebuilder/releases/tag/v1.0.8)

# Notes

Run ```make generate``` when add/modify types under pkg/apis which types are needed to serialize/deserialize a Kubernetes to/from YAML or JSON.