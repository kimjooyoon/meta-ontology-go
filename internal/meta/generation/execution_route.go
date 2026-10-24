package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

// ExecutionRoute identifies a native handler in scripts/meta-execution. An
// authored operation can exist before that consumer implements its handler.
// Routes live in the canonical registry and are covered by its plan digest.
type ExecutionRoute string

const (
	RouteCollapseAssign  ExecutionRoute = "collapse-assign-return"
	RouteSplitGo         ExecutionRoute = "split-go-declarations"
	RouteExtractFunction ExecutionRoute = "extract-function"
	RouteRegisterSyntax  ExecutionRoute = "register-syntax"
)

func NativeExecutionRoute(operation sourcepolicy.Operation) ExecutionRoute {
	for _, binding := range DefaultRegistry() {
		if binding.Operation == operation {
			return binding.ExecutionRoute
		}
	}
	return ""
}
