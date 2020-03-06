package controller

import (
	"github.com/keptn/keptn-operator/pkg/controller/keptnsubscription"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, keptnsubscription.Add)
}
