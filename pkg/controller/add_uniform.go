package controller

import (
	"github.com/keptn/keptn-operator/pkg/controller/uniform"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, uniform.Add)
}
