package metrictransition

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"
)

type Options struct {
	MetricsPath, EffectPath, ReceiptsPath, ProvenancePath, PatchPath string
	ExpectedSHA, CIRunID                                             string
}

type inputSet struct {
	metrics, effect, receipts, provenance, patch []byte
	report                                       linecaps.LineMetricsReport
	effectLedger                                 transformationeffect.Ledger
	receiptReport                                generation.ReceiptReport
	provenanceReport                             generation.ArtifactProvenance
}

func loadInputs(options Options) (inputSet, error) {
	paths := []string{options.MetricsPath, options.EffectPath, options.ReceiptsPath, options.ProvenancePath, options.PatchPath}
	raw := make([][]byte, len(paths))
	for index, path := range paths {
		if path == "" {
			return inputSet{}, fmt.Errorf("metric transition input path must not be empty")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return inputSet{}, err
		}
		raw[index] = data
	}
	inputs := inputSet{metrics: raw[0], effect: raw[1], receipts: raw[2], provenance: raw[3], patch: raw[4]}
	if err := json.Unmarshal(inputs.metrics, &inputs.report); err != nil {
		return inputSet{}, fmt.Errorf("decode source metrics: %w", err)
	}
	if err := json.Unmarshal(inputs.effect, &inputs.effectLedger); err != nil {
		return inputSet{}, fmt.Errorf("decode effect ledger: %w", err)
	}
	if err := json.Unmarshal(inputs.receipts, &inputs.receiptReport); err != nil {
		return inputSet{}, fmt.Errorf("decode effect receipts: %w", err)
	}
	if err := json.Unmarshal(inputs.provenance, &inputs.provenanceReport); err != nil {
		return inputSet{}, fmt.Errorf("decode effect provenance: %w", err)
	}
	return inputs, validateBinding(options, inputs)
}

func validateBinding(options Options, inputs inputSet) error {
	if options.ExpectedSHA == "" || options.CIRunID == "" || inputs.report.CommitSHA != options.ExpectedSHA {
		return fmt.Errorf("metric transition exact-head binding is invalid")
	}
	_, err := validateEffectOutcome(inputs)
	return err
}
