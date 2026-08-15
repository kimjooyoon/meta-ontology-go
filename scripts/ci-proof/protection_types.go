package main

type requiredCheckBinding struct {
	Context string `json:"context"`
	AppID   int64  `json:"app_id"`
}

func requiredCheckBindingsFor(contexts []string) []requiredCheckBinding {
	bindings := make([]requiredCheckBinding, 0, len(contexts))
	for _, context := range contexts {
		bindings = append(bindings, requiredCheckBinding{Context: context, AppID: 15368})
	}
	return bindings
}

func validRequiredCheckBindings(bindings []requiredCheckBinding, contexts []string) bool {
	if len(bindings) != len(contexts) {
		return false
	}
	expected := make(map[string]bool, len(contexts))
	for _, context := range contexts {
		expected[context] = true
	}
	seen := make(map[string]bool, len(bindings))
	for _, binding := range bindings {
		if binding.AppID != 15368 || !expected[binding.Context] || seen[binding.Context] {
			return false
		}
		seen[binding.Context] = true
	}
	return len(seen) == len(expected)
}
