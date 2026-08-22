package artifact

import (
	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/improvement"
)

func Build(conceptArtifact []byte, headSHA string) (Receipt, error) {
	snapshot, err := readiness.Evaluate(conceptArtifact)
	if err != nil {
		return Receipt{}, err
	}
	input := improvement.FromReadiness(snapshot)
	receipt := seal(Receipt{
		Schema:          Schema,
		HeadSHA:         headSHA,
		Snapshot:        snapshot,
		TransitionInput: input,
		FixedPoint:      improvement.Evaluate(input, input),
	})
	if err := Validate(receipt); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}
