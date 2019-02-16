package resources

import (
	"context"

	autuscaling "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	updateResourceScope = logUtil.RegisterScope("updateresource", "Update resource", 0)
)

// UpdateResource define resource update functions
type UpdateResource struct {
	client.Client
}

// NewUpdateResource return UpdateResource instance
func NewUpdateResource(client client.Client) *UpdateResource {
	return &UpdateResource{
		client,
	}
}

// UpdateAlamedaScaler updates AlamedaScaler
func (updateResource *UpdateResource) UpdateAlamedaScaler(alamedaScaler *autuscaling.AlamedaScaler) error {
	err := updateResource.updateResource(alamedaScaler)
	return err
}

func (updateResource *UpdateResource) updateResource(resource runtime.Object) error {
	if err := updateResource.Update(context.TODO(),
		resource); err != nil {
		updateResourceScope.Debug(err.Error())
		return err
	}
	return nil
}
