package metarecognition

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
)

func evaluateGooo(c Case) Outcome {
	switch c.Baseline.Subject {
	case SubjectBinding:
		return evaluateBinding(c.Baseline)
	case SubjectGraph:
		return evaluateGraph(c.Baseline)
	case SubjectSoundness:
		return evaluateSoundness(c.Baseline)
	case SubjectPath:
		return evaluatePath(c.Baseline)
	case SubjectResource:
		return evaluateResource(c.Baseline)
	default:
		return productionOutcome(UnknownFullSuiteRequired, ReasonSourceMapRegistry, nil, Work{})
	}
}
func evaluateBinding(b BaselineConfig) Outcome {
	declaration := b.DeclarationName
	if declaration == "" {
		declaration = "Order"
	}
	directive := "//gooo:bind id=\"billing://order\" role=\"HANDWRITTEN_IMPL\"\n"
	if !b.DirectivePresent {
		directive = ""
	}
	source := []byte("package billing\n\n" + directive + "type " + declaration + " struct{}\n")
	sources := []semanticbinding.SourceFile{{Filename: b.ObservedFile, PackagePath: "billing", Source: source}}
	if b.Ambiguous {
		sources = append(sources, sources[0])
	}
	registered := []string{b.StableID}
	if !b.RegistryPresent {
		registered = []string{"billing://other"}
	}
	result, _ := semanticbinding.Extract(semanticbinding.Input{Sources: sources, RegisteredIDs: registered})
	work := Work{Full: len(sources), Selected: len(result.Bindings), ProvRecords: len(result.Bindings) + len(result.Obligations), ProvPaths: len(result.Bindings)}
	if result.Status == semanticbinding.StatusUnknown {
		return productionOutcome(UnknownFullSuiteRequired, ReasonSourceMapRegistry, []string{b.StableID}, work)
	}
	if len(result.Bindings) == 0 {
		return productionOutcome(UnknownFullSuiteRequired, ReasonBlobWithoutID, []string{b.ObservedFile}, work)
	}
	if len(result.Bindings) != 1 || result.Bindings[0].ID != b.BoundID {
		return productionOutcome(FailClosedUnsound, ReasonBlobWithoutID, []string{b.StableID}, work)
	}
	if b.ExpectedFile != b.ObservedFile || b.ExpectedBlob != b.ObservedBlob {
		return productionOutcome(FailClosedUnsound, ReasonBlobWithoutID, []string{b.StableID}, work)
	}
	reason := ReasonExactBinding
	if result.Bindings[0].DeclarationKey != "Order" {
		reason = ReasonRenameBinding
	}
	return productionOutcome(ClosedSound, reason, []string{b.StableID}, work)
}
