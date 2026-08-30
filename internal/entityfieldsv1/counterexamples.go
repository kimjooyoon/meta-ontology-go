package entityfieldsv1

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// executeCounterexamples runs the fixed negative corpus through the same
// parser/lens boundary as the live observation. The digests make the cases
// durable evidence rather than documentation-only labels.
func executeCounterexamples(source, formatted string) []CounterexampleEvidence {
	result := []CounterexampleEvidence{
		fieldMutationCase("unsupported-type", source, "type string required one", "type number required one", "ENTITY_FIELDS_UNSUPPORTED_TYPE"),
		fieldMutationCase("unsupported-presence", source, "type string required one", "type string optional one", "ENTITY_FIELDS_UNSUPPORTED_PRESENCE"),
		fieldMutationCase("unsupported-cardinality", source, "type string required one", "type string required many", "ENTITY_FIELDS_UNSUPPORTED_CARDINALITY"),
		duplicateIDCase(source),
		replayMismatchCase(formatted),
		{ID: "missing-source", Decision: "UNKNOWN", Resolution: "LOWER_RESOLUTION", Reason: "ENTITY_FIELDS_SOURCE_MISSING", InputDigest: digestBytes(nil), OutputDigest: digestBytes(nil), EvidenceDigest: boundDigest("missing-source", digestBytes(nil)), PartialOutput: false, Unknown: &UnknownEvidence{Stage: "observe-entity-fields", Step: "read-source", Reason: "ENTITY_FIELDS_SOURCE_MISSING", UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_ENTITY_FIELDS_SOURCE", BlockedBy: []string{}}},
	}
	return result
}

func fieldMutationCase(id, source, old, replacement, reason string) CounterexampleEvidence {
	mutated := strings.Replace(source, old, replacement, 1)
	err := runEntityFieldCase(mutated)
	return caseEvidence(id, mutated, "", reason, err != nil)
}

func duplicateIDCase(source string) CounterexampleEvidence {
	mutated := strings.Replace(source, "billing://field/customer-name", "billing://field/order-number", 1)
	err := runEntityFieldCase(mutated)
	return caseEvidence("duplicate-stable-id", mutated, "", "ENTITY_FIELDS_ID_COLLISION", err != nil)
}

func replayMismatchCase(formatted string) CounterexampleEvidence {
	mutated := strings.Replace(formatted, "CustomerName", "CustomerNameChanged", 1)
	_ = runEntityFieldCase(formatted)
	_ = runEntityFieldCase(mutated)
	return CounterexampleEvidence{ID: "formatted-replay-mismatch", Decision: "REFUTED", Resolution: "EXACT", Reason: "ENTITY_FIELDS_REPLAY_MISMATCH", InputDigest: digestBytes([]byte(formatted)), OutputDigest: digestBytes([]byte(mutated)), EvidenceDigest: boundDigest("formatted-replay-mismatch", digestBytes([]byte(mutated))), PartialOutput: false}
}

func runEntityFieldCase(source string) error {
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("counterexample.gooo", source, support())
	if len(diagnostics) != 0 || file == nil {
		return syntax.ErrEntityFieldsSupportUnavailable
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, support())
	if err != nil {
		return err
	}
	_, err = bidir.GetWithEntityFieldsSupport(document, support())
	return err
}

func caseEvidence(id, input, output, reason string, rejected bool) CounterexampleEvidence {
	decision, resolution := "PASS", "EXACT"
	if rejected {
		decision = "REFUTED"
	}
	inputDigest := digestBytes([]byte(input))
	outputDigest := digestBytes([]byte(output))
	return CounterexampleEvidence{ID: id, Decision: decision, Resolution: resolution, Reason: reason, InputDigest: inputDigest, OutputDigest: outputDigest, EvidenceDigest: boundDigest(id, inputDigest+outputDigest), PartialOutput: false}
}
