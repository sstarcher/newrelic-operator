package controller

import (
	"github.com/sstarcher/newrelic-operator/pkg/controller/dashboard"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, dashboard.Add)
}
