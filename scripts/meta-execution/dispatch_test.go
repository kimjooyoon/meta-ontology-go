package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestAdvertisedNativeRoutesHaveExecutionHandlers(t *testing.T) {
	advertised := make(map[generation.ExecutionRoute]bool)
	for _, binding := range generation.DefaultRegistry() {
		if binding.ExecutionRoute == "" {
			continue
		}
		if nativeExecutors[binding.ExecutionRoute] == nil {
			t.Fatal("planner advertises a route without a native handler", binding)
		}
		advertised[binding.ExecutionRoute] = true
	}
	for route := range nativeExecutors {
		if route == "" || !advertised[route] {
			t.Fatal("native handler has no canonical registry binding", route)
		}
	}
}

func TestUnconnectedGoooOperationCannotReachMaterialization(t *testing.T) {
	_, failure := executeAction(t.TempDir(), "", "", generation.Plan{},
		generation.Action{Operation: sourcepolicy.OperationSplitGooo}, metaExecutionTrace{})
	if failure == nil || failure.reason != "UNSUPPORTED_SELECTED_OPERATION" {
		t.Fatal("unconnected operation reached execution", failure)
	}
}
