package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func splitGoBinding() (generation.Binding, error) {
	for _, binding := range generation.DefaultRegistry() {
		if binding.Operation == sourcepolicy.OperationSplitGo {
			return binding, nil
		}
	}
	return generation.Binding{}, fmt.Errorf(
		"registry binding %q is missing", sourcepolicy.OperationSplitGo,
	)
}
