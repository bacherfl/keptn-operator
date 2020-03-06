package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type UniformService struct {
	Name string `json:"name"`
	Image string `json:"image"`
	Events []string `json:"events"`
	Env []corev1.EnvVar `json:"env,omitempty"`
}
// UniformSpec defines the desired state of Uniform
type UniformSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Services []*UniformService `json:"services"`
}

// UniformStatus defines the observed state of Uniform
type UniformStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Uniform is the Schema for the uniforms API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=uniforms,scope=Namespaced
type Uniform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UniformSpec   `json:"spec,omitempty"`
	Status UniformStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UniformList contains a list of Uniform
type UniformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Uniform `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Uniform{}, &UniformList{})
}
