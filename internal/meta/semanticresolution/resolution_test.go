package semanticresolution

import "testing"

func TestResolutionLadderIsFiniteAndMonotone(t *testing.T) {
	current := ResolutionExactOperation
	for _, want := range []Resolution{ResolutionOperationClass, ResolutionInvariantOnly} {
		next, ok := LowerSemanticResolution(current)
		if !ok || next != want {
			t.Fatalf("lower(%q) = %q, %v; want %q, true", current, next, ok, want)
		}
		current = next
	}
	if next, ok := LowerSemanticResolution(current); ok || next != "" {
		t.Fatalf("lowest resolution descended to %q, %v", next, ok)
	}
}

func TestConflictLowersExactlyOneResolution(t *testing.T) {
	transition := ResolveSemanticConflict(Conflict{SourceDecision: "UNKNOWN", CurrentResolution: ResolutionExactOperation})
	if transition.Decision != "LOWER_RESOLUTION" || transition.ToResolution != ResolutionOperationClass || transition.Descents != 1 {
		t.Fatalf("unexpected transition: %+v", transition)
	}
}

func TestResolutionBudgetFailsClosed(t *testing.T) {
	transition := ResolveSemanticConflict(Conflict{SourceDecision: "UNKNOWN", CurrentResolution: ResolutionInvariantOnly, Descents: MaxResolutionDescents})
	if transition.Decision != "FAIL_CLOSED" || transition.Reason != "SEMANTIC_RESOLUTION_BUDGET_EXHAUSTED" {
		t.Fatalf("unexpected transition: %+v", transition)
	}
}

func TestResolutionWriteEffectFailsClosed(t *testing.T) {
	transition := ResolveSemanticConflict(Conflict{SourceDecision: "UNKNOWN", CurrentResolution: ResolutionExactOperation, RepositoryWrites: 1})
	if transition.Decision != "FAIL_CLOSED" || transition.Reason != "SEMANTIC_RESOLUTION_WRITE_EFFECT" {
		t.Fatalf("unexpected transition: %+v", transition)
	}
}
