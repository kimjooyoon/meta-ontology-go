package transformationeffect

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func validateSplitGoEffects(effects []Effect) error {
	for index, effect := range effects {
		artifact := effect.SplitGoEvaluation
		if artifact == nil {
			continue
		}
		if effect.Operation != string(sourcepolicy.OperationSplitGo) {
			return fmt.Errorf("effect %d attaches SplitGo evidence to %s", index, effect.Operation)
		}
		if effect.EvaluatorEvidence != hashJSON(*artifact) {
			return fmt.Errorf("effect %d SplitGo evaluator evidence digest is not bound", index)
		}
		if err := ValidateSplitGoEvaluation(*artifact); err != nil {
			return fmt.Errorf("effect %d SplitGo evaluator replay: %w", index, err)
		}
	}
	return nil
}
