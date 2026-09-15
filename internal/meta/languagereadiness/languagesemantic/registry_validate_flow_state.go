package languagesemantic

type registry_validateFlowState struct {
	done    bool
	slot00  Registry
	slot01  map[string]struct{}
	slot02  int
	slot03  int
	slot04  int
	slot05  []string
	result0 error
}
