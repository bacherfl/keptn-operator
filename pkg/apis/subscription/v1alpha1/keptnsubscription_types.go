package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KeptnSubscriptionSpec defines the desired state of KeptnSubscription
// +k8s:openapi-gen=true
type KeptnSubscriptionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Topics string `json:"topics"`
	Listener string `json:"listener"`
}

// KeptnSubscriptionStatus defines the observed state of KeptnSubscription
// +k8s:openapi-gen=true
type KeptnSubscriptionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KeptnSubscription is the Schema for the keptnsubscriptions API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type KeptnSubscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeptnSubscriptionSpec   `json:"spec,omitempty"`
	Status KeptnSubscriptionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KeptnSubscriptionList contains a list of KeptnSubscription
type KeptnSubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeptnSubscription `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeptnSubscription{}, &KeptnSubscriptionList{})
}
