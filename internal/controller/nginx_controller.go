package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webv1alpha1 "github.com/Sherpa99/nginx-opr/api/v1alpha1"
)

type NginxReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=web.mydomain.com,resources=nginxes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=web.mydomain.com,resources=nginxes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=web.mydomain.com,resources=nginxes/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

func (r *NginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 1. Get the Nginx CR
	var nginx webv1alpha1.Nginx
	if err := r.Get(ctx, req.NamespacedName, &nginx); err != nil {
		if apierrors.IsNotFound(err) {
			// CR deleted, nothing to do
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Nginx")
		return ctrl.Result{}, err
	}

	// 2. Construct desired Deployment
	deploy := constructDeployment(&nginx)

	// 3. Set OwnerReference for automatic cleanup
	if err := controllerutil.SetControllerReference(&nginx, deploy, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// 4. Check if Deployment already exists
	var existing appsv1.Deployment
	err := r.Get(ctx, types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, &existing)

	if err != nil && apierrors.IsNotFound(err) {
		// 5. Create Deployment if missing
		log.Info("Creating Deployment", "name", deploy.Name)
		if err := r.Create(ctx, deploy); err != nil {
			return ctrl.Result{}, err
		}
	} else if err == nil {
		// 6. Update if replica count has changed
		if *existing.Spec.Replicas != *deploy.Spec.Replicas {
			existing.Spec.Replicas = deploy.Spec.Replicas
			log.Info("Updating Deployment replicas", "name", deploy.Name)
			if err := r.Update(ctx, &existing); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// 7. Update Status
	nginxCopy := nginx.DeepCopy()
	nginxCopy.Status.AvailableReplicas = existing.Status.AvailableReplicas
	if err := r.Status().Update(ctx, nginxCopy); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NginxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1alpha1.Nginx{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
