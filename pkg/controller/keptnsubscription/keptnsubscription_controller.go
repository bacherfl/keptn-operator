package keptnsubscription

import (
	"context"
	"github.com/go-logr/logr"

	subscriptionv1alpha1 "github.com/keptn/keptn-operator/pkg/apis/subscription/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_keptnsubscription")

const subscriptionFinalizer = "finalizer.subscription.keptn.sh"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new KeptnSubscription Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKeptnSubscription{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("keptnsubscription-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KeptnSubscription
	err = c.Watch(&source.Kind{Type: &subscriptionv1alpha1.KeptnSubscription{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner KeptnSubscription
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &subscriptionv1alpha1.KeptnSubscription{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &subscriptionv1alpha1.KeptnSubscription{},
		IsController: true,
	})
	return nil
}

// blank assignment to verify that ReconcileKeptnSubscription implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKeptnSubscription{}

// ReconcileKeptnSubscription reconciles a KeptnSubscription object
type ReconcileKeptnSubscription struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KeptnSubscription object and makes changes based on the state read
// and what is in the KeptnSubscription.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKeptnSubscription) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KeptnSubscription")

	// Fetch the KeptnSubscription instance
	instance := &subscriptionv1alpha1.KeptnSubscription{}
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

	// Check if the subscription instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	issubscriptionMarkedToBeDeleted := instance.GetDeletionTimestamp() != nil
	if issubscriptionMarkedToBeDeleted {
		if contains(instance.GetFinalizers(), subscriptionFinalizer) {
			// Run finalization logic for subscriptionFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeSubscription(reqLogger, instance); err != nil {
				return reconcile.Result{}, err
			}

			// Remove subscriptionFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			instance.SetFinalizers(removeFinalizer(instance.GetFinalizers(), subscriptionFinalizer))
			err := r.client.Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// Add finalizer for this CR
	if !contains(instance.GetFinalizers(), subscriptionFinalizer) {
		if err := r.addFinalizer(reqLogger, instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Define a new Pod object
	pod := newPodForCR(instance)

	container := newContainerForCR(instance)

	// Set KeptnSubscription instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}


	// Check if this Pod already exists
	/*
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	 */

	foundDepl, err := r.getDeploymentForSubscriber(instance)

	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Did not find deployment ", "Deployment.Name", instance.Spec.Listener)
	}

	reqLogger.Info("Found deployment ", "Deployment.Name", instance.Spec.Listener)

	foundDistributorContainer := false
	distrIndex := 0
	for index, c := range foundDepl.Spec.Template.Spec.Containers {
		if c.Name == "distributor" {
			reqLogger.Info("Updating distributor spec in deployment ", "Deployment.Name", instance.Spec.Listener)
			foundDistributorContainer = true
			distrIndex = index
			break
		}
	}
	if foundDistributorContainer {
		foundDepl.Spec.Template.Spec.Containers = remove(foundDepl.Spec.Template.Spec.Containers, distrIndex)
	}

	reqLogger.Info("Injecting distributor into deployment ", "Deployment.Name", instance.Spec.Listener)
	foundDepl.Spec.Template.Spec.Containers = append(foundDepl.Spec.Template.Spec.Containers, *container)

	err = r.client.Update(context.TODO(), foundDepl)
	if err != nil {
		reqLogger.Info("updating deployment failed", "Deployment.Name", instance.Spec.Listener)
		return reconcile.Result{}, err
	}

	reqLogger.Info("Updated deployment", "Deployment.Name", instance.Spec.Listener)

	// Pod already exists - don't requeue
	// reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

func (r *ReconcileKeptnSubscription) getDeploymentForSubscriber(instance *subscriptionv1alpha1.KeptnSubscription) (*appsv1.Deployment, error) {
	foundDepl := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Listener, Namespace: instance.Namespace}, foundDepl)

	return foundDepl, err
}

func (r *ReconcileKeptnSubscription) finalizeSubscription(reqLogger logr.Logger, s *subscriptionv1alpha1.KeptnSubscription) error {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.
	reqLogger.Info("Successfully finalized subscription")

	subscriber, err := r.getDeploymentForSubscriber(s)

	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Did not find deployment. Nothing to delete. ", "Deployment.Name", s.Spec.Listener)
	}

	foundDistributorContainer := false
	distrIndex := 0
	for index, c := range subscriber.Spec.Template.Spec.Containers {
		if c.Name == "distributor" {
			reqLogger.Info("Updating distributor spec in deployment ", "Deployment.Name", s.Spec.Listener)
			foundDistributorContainer = true
			distrIndex = index
			break
		}
	}
	if !foundDistributorContainer {
		return nil
	}

		subscriber.Spec.Template.Spec.Containers = remove(subscriber.Spec.Template.Spec.Containers, distrIndex)

	err = r.client.Update(context.TODO(), subscriber)
	if err != nil {
		reqLogger.Info("updating deployment failed", "Deployment.Name", s.Spec.Listener)
		return err
	}

	reqLogger.Info("Updated deployment", "Deployment.Name", s.Spec.Listener)
	return nil
}

func (r *ReconcileKeptnSubscription) addFinalizer(reqLogger logr.Logger, m *subscriptionv1alpha1.KeptnSubscription) error {
	reqLogger.Info("Adding Finalizer for the subscription")
	m.SetFinalizers(append(m.GetFinalizers(), subscriptionFinalizer))

	// Update CR
	err := r.client.Update(context.TODO(), m)
	if err != nil {
		reqLogger.Error(err, "Failed to update subscription with finalizer")
		return err
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func removeFinalizer(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}

func remove(slice []corev1.Container, s int) []corev1.Container {
	return append(slice[:s], slice[s+1:]...)
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *subscriptionv1alpha1.KeptnSubscription) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "distributor",
					Image:   "keptn/distributor:latest",
					Env:     []corev1.EnvVar{
						{
							Name:  "PUBSUB_RECIPIENT",
							Value: cr.Spec.Listener,
						},
						{
							Name:  "PUBSUB_TOPIC",
							Value: cr.Spec.Topics,
						},
						{
							Name:  "PUBSUB_URL",
							Value: "nats://keptn-nats-cluster",
						},
					},
				},
			},
		},
	}
}

func newContainerForCR(cr *subscriptionv1alpha1.KeptnSubscription) *corev1.Container {
	return &corev1.Container{
		Name:    "distributor",
		Image:   "keptn/distributor:latest",
		Env:     []corev1.EnvVar{
			{
				Name:  "PUBSUB_RECIPIENT",
				Value: cr.Spec.Listener,
			},
			{
				Name:  "PUBSUB_TOPIC",
				Value: cr.Spec.Topics,
			},
			{
				Name:  "PUBSUB_URL",
				Value: "nats://keptn-nats-cluster",
			},
		},
	}
}
