package foundationseed

func indicators(source Source) []Indicator {
	deltaValue, deltaKnown := 0, false
	if source.ReadinessDeltaClaims != nil {
		deltaValue, deltaKnown = *source.ReadinessDeltaClaims, true
	}
	return []Indicator{
		indicator("resolution-contract-valid", "DRIVER", "FOUNDATION",
			boolInt(source.ResolutionValid), 1, source.ResolutionValid),
		indicator("current-head-bound", "DRIVER", "COHERENCE",
			boolInt(source.HeadBound), 1, source.HeadBound),
		indicator("search-limit-complete", "DRIVER", "FOUNDATION",
			source.ObservedAttempts, source.SearchLimit, source.SearchComplete),
		indicator("missing-attempts-complete", "DRIVER", "FOUNDATION",
			source.MissingAttempts, source.SearchLimit, source.MissingComplete),
		indicator("contiguous-parent-links", "DRIVER", "COHERENCE",
			boolInt(source.Contiguous), 1, source.Contiguous),
		indicator("selected-ancestors-zero", "GUARDRAIL", "FOUNDATION",
			source.SelectedAncestors, 0, source.SelectedAncestors == 0),
		indicator("valid-candidates-zero", "GUARDRAIL", "FOUNDATION",
			source.ValidCandidates, 0, source.ValidCandidates == 0),
		indicator("ambiguous-candidates-zero", "GUARDRAIL", "REGRESSION",
			source.AmbiguousCandidates, 0, source.AmbiguousCandidates == 0),
		indicator("repository-writes-zero", "GUARDRAIL", "REGRESSION",
			source.RepositoryWrites, 0, source.RepositoryWrites == 0),
		indicator("readiness-delta-claims-zero", "OUTCOME", "REGRESSION",
			deltaValue, 0, deltaKnown && deltaValue == 0),
		indicator("seed-scope-exact", "OUTCOME", "FOUNDATION",
			boolInt(source.ExactExhaustion), 1, source.ExactExhaustion),
		indicator("authority-denied", "GUARDRAIL", "REGRESSION",
			boolInt(source.AuthorityDenied), 1, source.AuthorityDenied),
	}
}

func indicator(id, class, choice string, value, target int, passed bool) Indicator {
	return Indicator{ID: id, Class: class, Choice: choice,
		Value: value, Target: target, Passed: passed}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
