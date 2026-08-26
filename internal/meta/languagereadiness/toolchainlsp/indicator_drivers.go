package toolchainlsp

func driverIndicators(s Summary, resolution string) []Indicator {
	values := []struct {
		name, proof   string
		value, target int
	}{
		{"executed-cases.v1", "COHERENCE", s.CasesSatisfied, 22},
		{"protocol-cases.v1", "FOUNDATION", s.ProtocolCases, 16},
		{"coupling-cases.v1", "COHERENCE", s.CouplingCases, 6},
		{"advertised-capabilities.v1", "FOUNDATION", s.AdvertisedCapabilities, 8},
		{"read-features.v1", "COHERENCE", s.ReadFeatures, 7},
		{"diagnostic-paths.v1", "REGRESSION", s.DiagnosticPaths, 3},
		{"navigation-paths.v1", "COHERENCE", s.NavigationPaths, 3},
		{"symbol-paths.v1", "COHERENCE", s.SymbolPaths, 2},
		{"semantic-token-paths.v1", "COHERENCE", s.SemanticTokenPaths, 1},
		{"utf16-replays.v1", "FOUNDATION", s.UTF16Replays, 1},
		{"transcript-replays.v1", "REGRESSION", s.TranscriptReplays, 1},
		{"fail-closed-paths.v1", "REGRESSION", s.FailClosedPaths, 5},
		{"concept-bindings.v1", "FOUNDATION", s.ConceptBindings, 1},
		{"code-bindings.v1", "FOUNDATION", s.CodeBindings, 5},
		{"metric-bindings.v1", "COHERENCE", s.MetricBindings, 37},
		{"use-case-bindings.v1", "REGRESSION", s.UseCaseBindings, 3},
	}
	result := make([]Indicator, 0, len(values))
	for _, value := range values {
		result = append(result, indicator(value.name, "DRIVER", value.proof,
			value.value, value.target, "greater_or_equal", resolution))
	}
	return result
}
