package languagedebugexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"

const RuntimeReceiptSchema = "gooo/language-debug-runtime-receipt/v1"

type Input struct {
	SubjectSHA          string                `json:"subject_sha"`
	ExecutableDigest    string                `json:"executable_digest"`
	Contract            Contract              `json:"contract"`
	First               languagedebug.Receipt `json:"first"`
	Second              languagedebug.Receipt `json:"second"`
	UnknownBreakpoint   languagedebug.Receipt `json:"unknown_breakpoint"`
	RuntimeObservations []RuntimeObservation  `json:"runtime_observations"`
	Build               Measurement           `json:"build"`
	EvaluatorBuild      Measurement           `json:"evaluator_build"`
	Test                Measurement           `json:"test"`
	Graph               GraphObservation      `json:"graph"`
}
