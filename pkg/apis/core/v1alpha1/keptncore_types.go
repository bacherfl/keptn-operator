package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KeptnCoreSpec defines the desired state of KeptnCore
// +k8s:openapi-gen=true
type KeptnCoreSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Platform string `json:"platform"`
	Version string `json:"version"`
}

// KeptnCoreStatus defines the observed state of KeptnCore
// +k8s:openapi-gen=true
type KeptnCoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KeptnCore is the Schema for the keptncores API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type KeptnCore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeptnCoreSpec   `json:"spec,omitempty"`
	Status KeptnCoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KeptnCoreList contains a list of KeptnCore
type KeptnCoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeptnCore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeptnCore{}, &KeptnCoreList{})
}
