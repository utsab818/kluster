package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{
	Group:   "utsab.dev",
	Version: "v1alpha1",
}

// SchemeBuilder from apimachinery/runtime is loaded to register in scheme
var (
	SchemeBuilder runtime.SchemeBuilder
	AddToScheme   = SchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// The init function calls itself at first which registers the function addKnownTypes
func init() {
	SchemeBuilder.Register(addKnownTypes)
}

// This function is registerd under SchemeBuilder.
func addKnownTypes(scheme *runtime.Scheme) error {
	// for the &kluster{} to be implemented in addKnownTypes,
	// the Kluster struct should implement some behavious
	// deepcopyobject, set and get GV for the type.
	// We can use codegenerator for this.
	// Get and Set GV is implemented using typemeta and objectmeta
	// in Kluster and KlusterList struct
	// Deepcopy of the object should be added.
	scheme.AddKnownTypes(SchemeGroupVersion, &Kluster{}, &KlusterList{})

	// add the Group-Version
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
