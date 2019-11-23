package v1alpha1

import (
	autoscalingapi "github.com/containers-ai/alameda/operator/api"
	runtime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *AlamedaScaler) SetupWebhookWithManager(mgr ctrl.Manager) error {
	r.Mgr = mgr
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-autoscaling-containers-ai-v1alpha1-alamedascaler,mutating=true,failurePolicy=fail,groups=autoscaling.containers.ai,resources=alamedascalers,verbs=create;update,versions=v1alpha1,name=malamedascaler.kb.io

var _ webhook.Defaulter = &AlamedaScaler{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AlamedaScaler) Default() {
	scope.Infof("Default webhook for AlamedaScaler %s", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-autoscaling-containers-ai-v1alpha1-alamedascaler,mutating=false,failurePolicy=fail,groups=autoscaling.containers.ai,resources=alamedascalers,versions=v1alpha1,name=valamedascaler.kb.io

var _ webhook.Validator = &AlamedaScaler{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaScaler) ValidateCreate() error {
	scope.Infof("Validate Create webhook for AlamedaScaler %s", r.Name)
	_, err := r.validateAlamedaScalersFn()
	// TODO(user): fill in your validation logic upon object creation.
	return err
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaScaler) ValidateUpdate(old runtime.Object) error {
	scope.Infof("Validate Create webhook for AlamedaScaler %s", r.Name)
	_, err := r.validateAlamedaScalersFn()
	// TODO(user): fill in your validation logic upon object update.
	return err
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaScaler) ValidateDelete() error {
	scope.Infof("Validate Delete webhook for AlamedaScaler %s", r.Name)
	_, err := r.validateAlamedaScalersFn()
	// TODO(user): fill in your validation logic upon object deletion.
	return err
}

// validateAlamedaScalersFn validate the given alamedaScalerLabeler
func (r *AlamedaScaler) validateAlamedaScalersFn() (bool, error) {
	clnt := r.Mgr.GetClient()
	return r.Validate.IsScalerValid(&clnt, &autoscalingapi.ValidatingObject{
		Namespace:           r.GetNamespace(),
		Name:                r.GetName(),
		Kind:                r.GetObjectKind().GroupVersionKind().Kind,
		SelectorMatchLabels: r.Spec.Selector.MatchLabels,
	})
}
