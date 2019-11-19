package validate

import (
	"fmt"

	autoscalingapi "github.com/containers-ai/alameda/operator/api"
	"github.com/containers-ai/alameda/operator/pkg/utils/resources"
	"github.com/containers-ai/alameda/pkg/utils"
	"github.com/containers-ai/alameda/pkg/utils/kubernetes"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	openshift_apps_v1 "github.com/openshift/api/apps/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var scope = logUtil.RegisterScope("resources_validate", "resources validate", 0)

type ResourceValidate struct {
}

func (r *ResourceValidate) IsTopControllerValid(client *client.Client, topCtl *autoscalingapi.ValidatingObject) (bool, error) {
	listResources := resources.NewListResources(*client)
	// TODO: may use ListAllAlamedaScaler if alamedascaler supports selectNamespace option
	scalers, err := listResources.ListNamespaceAlamedaScaler(topCtl.Namespace)
	scope.Debugf("%v alamedascaler in namespace %s to check deplicated selection of %s %s",
		len(scalers), topCtl.Namespace, topCtl.Kind, topCtl.Name)
	if err != nil {
		return false, err
	}
	matchedScalerList := []*autoscalingapi.ValidatingObject{}
	for _, scaler := range scalers {
		if r.isLabelSelected(scaler.Spec.Selector.MatchLabels, topCtl.Labels) {
			matchedScalerList = append(matchedScalerList, &autoscalingapi.ValidatingObject{
				Name:      scaler.GetName(),
				Namespace: scaler.GetNamespace(),
			})
		}
	}
	if len(matchedScalerList) > 1 {
		matchedNamesapcedNames := fmt.Sprintf("%s/%s", matchedScalerList[0].Namespace, matchedScalerList[0].Name)
		for idx, matched := range matchedScalerList {
			if idx > 0 {
				matchedNamesapcedNames = fmt.Sprintf("%s, %s/%s", matchedNamesapcedNames, matched.Namespace, matched.Name)
			}
		}

		return false, fmt.Errorf("%s (%s/%s) is selected by more than 1 alamedascaler (%s)",
			topCtl.Kind, topCtl.Namespace, topCtl.Name, matchedNamesapcedNames)
	}
	return true, nil
}

func (r *ResourceValidate) getSelectedDeploymentConfigs(listResources *resources.ListResources, namespace string, selectorMatchLabels map[string]string) ([]openshift_apps_v1.DeploymentConfig, error) {
	okdCluster, err := kubernetes.IsOKDCluster()
	if err != nil {
		scope.Errorf(err.Error())
	}
	if !okdCluster {
		return []openshift_apps_v1.DeploymentConfig{}, nil
	}
	// TODO: may use ListDeploymentConfigsByLabels if alamedascaler supports selectNamespace option
	return listResources.ListDeploymentConfigsByNamespaceLabels(namespace, selectorMatchLabels)
}

func (r *ResourceValidate) IsScalerValid(client *client.Client, scalerObj *autoscalingapi.ValidatingObject) (bool, error) {
	listResources := resources.NewListResources(*client)
	// TODO: may use ListAllAlamedaScaler if alamedascaler supports selectNamespace option
	scalers, err := listResources.ListNamespaceAlamedaScaler(scalerObj.Namespace)
	if err != nil {
		return false, err
	}
	// TODO: may use ListDeploymentsByLabels if alamedascaler supports selectNamespace option
	selectedDeployments, err := listResources.ListDeploymentsByNamespaceLabels(
		scalerObj.Namespace, scalerObj.SelectorMatchLabels)
	if err != nil {
		return false, err
	}

	selectedDeploymentConfigs, err := r.getSelectedDeploymentConfigs(
		listResources, scalerObj.Namespace, scalerObj.SelectorMatchLabels)
	if err != nil {
		return false, err
	}

	for _, scaler := range scalers {
		if scaler.GetNamespace() == scalerObj.Namespace && scaler.GetName() == scalerObj.Name {
			continue
		}

		for _, selectedDeployment := range selectedDeployments {
			if _, ok := scaler.Status.AlamedaController.Deployments[fmt.Sprintf("%s/%s", selectedDeployment.GetNamespace(), selectedDeployment.GetName())]; ok {
				return false, fmt.Errorf("Deployment %s/%s selected by scaler %s/%s is already selected by scaler %s/%s",
					selectedDeployment.GetNamespace(), selectedDeployment.GetName(),
					scalerObj.Namespace, scalerObj.Name, scaler.GetNamespace(), scaler.GetName())
			}
		}
		for _, selectedDeploymentConfig := range selectedDeploymentConfigs {
			if _, ok := scaler.Status.AlamedaController.DeploymentConfigs[fmt.Sprintf("%s/%s", selectedDeploymentConfig.GetNamespace(), selectedDeploymentConfig.GetName())]; ok {
				return false, fmt.Errorf("DeploymentConfig %s/%s selected by scaler %s/%s is already selected by scaler %s/%s",
					selectedDeploymentConfig.GetNamespace(), selectedDeploymentConfig.GetName(),
					scalerObj.Namespace, scalerObj.Name, scaler.GetNamespace(), scaler.GetName())
			}
		}
	}
	return true, nil
}

func (r *ResourceValidate) isLabelSelected(selector, label map[string]string) bool {
	isSelected := true
	scope.Debugf("Check label is selected by selector.")
	scope.Debugf("Selector is %s.", utils.InterfaceToString(selector))
	scope.Debugf("Label is %s.", utils.InterfaceToString(label))
	for selKey, selVal := range selector {
		if _, ok := label[selKey]; !ok {
			isSelected = false
			break
		}
		if label[selKey] != selVal {
			isSelected = false
			break
		}
	}
	if isSelected {
		scope.Debugf("Label is matched by selector.")
	} else {
		scope.Debugf("Label is not matched by selector.")
	}
	return isSelected
}
