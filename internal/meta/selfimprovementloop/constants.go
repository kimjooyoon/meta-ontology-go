package selfimprovementloop

const (
	Schema             = "gooo/self-improvement-minimal-loop/v1"
	GraphSchemaVersion = "gooo-graph/v1"
	DecisionClosed     = "CLOSED"
	DecisionUnknown    = "UNKNOWN"
	DecisionRefuted    = "REFUTED"
)

var fixedCells = [...]string{
	"OBSERVE_BASELINE", "DECLARE_TARGET", "PIN_SCOPE", "BIND_META_ACTIVITY",
	"PROPOSE_TRANSFORMATION", "PREDICT_EFFECT", "BUILD_COUNTEREXAMPLE",
	"EXECUTE_CI", "CAPTURE_RECEIPT", "COMPARE_EXACT_PAIR", "HUMAN_DECISION",
	"PROPAGATE_OR_REFUTE",
}

// SemanticCells returns the immutable order of the loop's meaning cells.
func SemanticCells() []string { return append([]string(nil), fixedCells[:]...) }
