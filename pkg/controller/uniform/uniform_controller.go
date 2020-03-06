package uniform

import (
	"context"
	"strings"

	uniformv1alpha1 "github.com/keptn/keptn-operator/pkg/apis/uniform/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_uniform")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Uniform Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileUniform{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("uniform-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Uniform
	err = c.Watch(&source.Kind{Type: &uniformv1alpha1.Uniform{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Uniform
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &uniformv1alpha1.Uniform{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &uniformv1alpha1.Uniform{},
	})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileUniform implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileUniform{}

// ReconcileUniform reconciles a Uniform object
type ReconcileUniform struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Uniform object and makes changes based on the state read
// and what is in the Uniform.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileUniform) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Uniform")

	// Fetch the Uniform instance
	instance := &uniformv1alpha1.Uniform{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	retriggerReconcile := false

	for _, service := range instance.Spec.Services {
		deplForCR := newDeploymentForCR(service)
		// Set Uniform instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, deplForCR, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this Pod already exists
		foundDepl := &appsv1.Deployment{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: deplForCR.Name, Namespace: deplForCR.Namespace}, foundDepl)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", deplForCR.Namespace, "Deployment.Name", deplForCR.Name)
			err = r.client.Create(context.TODO(), deplForCR)
			if err != nil {
				return reconcile.Result{}, err
			}

			// Pod created successfully - don't requeue
		} else if err != nil {
			continue
		}

		if strings.Contains(foundDepl.Spec.Template.Labels["app.kubernetes.io/managed-by"], "skaffold") {
			reqLogger.Info("Not upgrading resource managed by skaffold", "Deployment.Namespace", foundDepl.Namespace, "Deployment.Name", foundDepl.Name)
			retriggerReconcile = true
			continue
		}
		foundDepl.Spec = deplForCR.Spec

		err := r.client.Update(context.TODO(), foundDepl)

		if err != nil {
			reqLogger.Error(err, "Could not update service")
		}

		// Deployment already exists - don't requeue
		reqLogger.Info("Updated deployment", "Deployment.Namespace", foundDepl.Namespace, "Deployment.Name", foundDepl.Name)

		serviceForCR := newServiceForCR(service)

		if err := controllerutil.SetControllerReference(instance, serviceForCR, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this Service already exists
		foundSvc := &corev1.Service{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: serviceForCR.Name, Namespace: serviceForCR.Namespace}, foundSvc)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Service", "Service.Namespace", serviceForCR.Namespace, "Service.Name", serviceForCR.Name)
			err = r.client.Create(context.TODO(), serviceForCR)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		// Service already exists - don't requeue
		reqLogger.Info("Skip reconcile: Service already exists", "Service.Namespace", foundSvc.Namespace, "Service.Name", foundSvc.Name)
	}

	if retriggerReconcile {
		// retrigger reconciliation
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

// newDeploymentForCR returns a deployment with the same name/namespace as the cr
func newDeploymentForCR(cr *uniformv1alpha1.UniformService) *appsv1.Deployment {
	topics := ""

	for i, topic := range cr.Events {
		topics = topics + topic
		if i < len(cr.Events) - 1 {
			topics = topics + ","
		}
	}

	f := func(s int32) *int32 {
		return &s
	}

	envVars := cr.Env

	envVars = ensureEnvVarIsSet("EVENTBROKER", "http://event-broker.keptn.svc.cluster.local/keptn", envVars)
	envVars = ensureEnvVarIsSet("CONFIGURATION_SERVICE", "http://configuration-service.keptn.svc.cluster.local:8080", envVars)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: "keptn",
		},
		Spec:       appsv1.DeploymentSpec{
			Replicas: f(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"run": cr.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"run": cr.Name,
					},
				},
				Spec:       corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      cr.Name,
							Image:     cr.Image,
							Ports:     []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env:       envVars,
						},
						{
							Name:    "distributor",
							Image:   "keptn/distributor:latest",
							Env:     []corev1.EnvVar{
								{
									Name:  "PUBSUB_RECIPIENT",
									Value: cr.Name,
								},
								{
									Name:  "PUBSUB_TOPIC",
									Value: topics,
								},
								{
									Name:  "PUBSUB_URL",
									Value: "nats://keptn-nats-cluster",
								},
							},
						},
					},
				},
			},
		},
	}
}

func ensureEnvVarIsSet(name string, value string, envVars []corev1.EnvVar) []corev1.EnvVar {

	for _, env := range envVars {
		if env.Name == name {
			return envVars
		}
	}

	newEnvVar := corev1.EnvVar{
		Name:  name,
		Value: value,
	}

	envVars = append(envVars, newEnvVar)
	return envVars
}

func newServiceForCR(cr *uniformv1alpha1.UniformService) *corev1.Service {
	return &corev1.Service{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: "keptn",
			Labels: map[string]string{
				"run": cr.Name,
			},
		},
		Spec:       corev1.ServiceSpec{
			Ports:        []corev1.ServicePort{
				{
					Protocol: "TCP",
					Port:     8080,
				},
			},
			Selector:     map[string]string{
				"run": cr.Name,
			},
		},
		Status:     corev1.ServiceStatus{},
	}
}
