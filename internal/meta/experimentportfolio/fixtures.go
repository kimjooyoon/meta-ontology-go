package experimentportfolio

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	receiptProducer = "portfolio-receipt-producer"
	receiptConsumer = "portfolio-adjudicator"
)

type candidateFixture struct {
	Contract          CandidateContract
	Coordinates       []Coordinate
	Counterexamples   []Counterexample
	UnknownLocations  []UnknownLocation
	ExtensionEvidence []ExtensionEvidence
}

func fixtureFor(candidateID string) (candidateFixture, error) {
	contract, ok := candidateContract(ExpectedContract(), candidateID)
	if !ok {
		return candidateFixture{}, fmt.Errorf("unknown portfolio candidate %q", candidateID)
	}
	coordinate := func(id string, numerator, denominator int, status, stage, step, reason string) Coordinate {
		return Coordinate{ID: id, Producer: receiptProducer, Consumer: receiptConsumer,
			MetaOperation: contract.MetaOperation, ProofChoice: contract.ProofChoice,
			Numerator: numerator, Denominator: denominator, Status: status,
			Stage: stage, Step: step, Reason: reason}
	}
	fixture := candidateFixture{Contract: contract}
	switch candidateID {
	case "derive":
		fixture.Coordinates = []Coordinate{
			coordinate("source-replay", 1, 1, "DISCHARGED", "FOUNDATION", "bind-source", "source digest and activity shape agree"),
			coordinate("receipt-independence", 1, 1, "DISCHARGED", "FOUNDATION", "seal-receipt", "producer receipt is independently sealed"),
			coordinate("counterexample-boundary", 2, 2, "DISCHARGED", "COMPARISON", "replay-counterexamples", "both fixed counterexample slots are explicit"),
			coordinate("unknown-localization", 2, 2, "DISCHARGED", "COMPARISON", "locate-unknowns", "both unknown positions retain stage and reason"),
			coordinate("extension-evidence", 0, 1, "OPEN", "EXTENSION", "await-consumer", "no downstream extension receipt yet"),
			coordinate("read-only-effects", 1, 1, "DISCHARGED", "GUARDRAIL", "inspect-effects", "receipt declares zero repository writes"),
			coordinate("source-semantic-causality", 0, 3, "REFUTED", "CAUSALITY", "run-3-case-contract", "three semantic interventions are directly refuted as digest-only bindings"),
		}
		fixture.Counterexamples = []Counterexample{
			{ID: "derive-ce-01", Location: "derive/input/ambiguous-binding", Claim: "a missing binding must not become a pass", Stage: "COMPARISON", Step: "replay-counterexamples", Reason: "ambiguous binding is retained"},
			{ID: "derive-ce-02", Location: "derive/output/unknown-extension", Claim: "an unproven extension stays open", Stage: "EXTENSION", Step: "await-consumer", Reason: "extension evidence is absent"},
		}
		fixture.UnknownLocations = []UnknownLocation{
			{ID: "derive-unknown-01", Path: "derive/input/ambiguous-binding", Stage: "COMPARISON", Step: "locate-unknowns", Reason: "binding source is not unique"},
			{ID: "derive-unknown-02", Path: "derive/output/unknown-extension", Stage: "EXTENSION", Step: "await-consumer", Reason: "consumer proof is not present"},
		}
		fixture.ExtensionEvidence = []ExtensionEvidence{
			{ID: "derive-extension", Claim: "a consumer can add a coordinate without rewriting old ones", Status: "OPEN", Evidence: "none", Stage: "EXTENSION", Step: "await-consumer", Reason: "extension has not been exercised"},
		}
	case "replay":
		fixture.Coordinates = []Coordinate{
			coordinate("source-replay", 1, 1, "DISCHARGED", "FOUNDATION", "bind-source", "source digest and activity shape agree"),
			coordinate("receipt-independence", 0, 1, "REFUTED", "FOUNDATION", "seal-receipt", "replay path reuses the producer boundary"),
			coordinate("counterexample-boundary", 1, 2, "OPEN", "COMPARISON", "replay-counterexamples", "one fixed counterexample needs an independent replay"),
			coordinate("unknown-localization", 1, 2, "OPEN", "COMPARISON", "locate-unknowns", "one unknown position is unresolved"),
			coordinate("extension-evidence", 1, 1, "DISCHARGED", "EXTENSION", "replay-consumer", "consumer replay records one extension"),
			coordinate("read-only-effects", 1, 1, "DISCHARGED", "GUARDRAIL", "inspect-effects", "receipt declares zero repository writes"),
			coordinate("source-semantic-causality", 0, 3, "REFUTED", "CAUSALITY", "run-3-case-contract", "three semantic interventions are directly refuted as digest-only bindings"),
		}
		fixture.Counterexamples = []Counterexample{
			{ID: "replay-ce-01", Location: "replay/output/unknown-extension", Claim: "a reused receipt must expose its boundary", Stage: "COMPARISON", Step: "replay-counterexamples", Reason: "independence is not discharged"},
		}
		fixture.UnknownLocations = []UnknownLocation{
			{ID: "replay-unknown-01", Path: "replay/output/unknown-extension", Stage: "COMPARISON", Step: "locate-unknowns", Reason: "one extension result is unresolved"},
		}
		fixture.ExtensionEvidence = []ExtensionEvidence{
			{ID: "replay-extension", Claim: "a consumer can add a coordinate without rewriting old ones", Status: "DISCHARGED", Evidence: "consumer replay receipt", Stage: "EXTENSION", Step: "replay-consumer", Reason: "one extension was replayed"},
		}
	case "reflect":
		fixture.Coordinates = []Coordinate{
			coordinate("source-replay", 1, 1, "DISCHARGED", "FOUNDATION", "bind-source", "source digest and activity shape agree"),
			coordinate("receipt-independence", 1, 1, "DISCHARGED", "FOUNDATION", "seal-receipt", "producer receipt is independently sealed"),
			coordinate("counterexample-boundary", 0, 2, "REFUTED", "COMPARISON", "replay-counterexamples", "no counterexample was retained"),
			coordinate("unknown-localization", 0, 2, "REFUTED", "COMPARISON", "locate-unknowns", "unknown positions are not recorded"),
			coordinate("extension-evidence", 0, 1, "REFUTED", "EXTENSION", "reflect-consumer", "extension claim has no evidence"),
			coordinate("read-only-effects", 1, 1, "DISCHARGED", "GUARDRAIL", "inspect-effects", "receipt declares zero repository writes"),
			coordinate("source-semantic-causality", 0, 3, "REFUTED", "CAUSALITY", "run-3-case-contract", "three semantic interventions are directly refuted as digest-only bindings"),
		}
	default:
		return candidateFixture{}, fmt.Errorf("unknown portfolio candidate %q", candidateID)
	}
	return fixture, nil
}

func ProduceReceipt(subjectSHA, sourcePath string, source []byte, candidateID string) (Receipt, error) {
	fixture, err := fixtureFor(candidateID)
	if err != nil {
		return Receipt{}, err
	}
	if !strings.HasSuffix(filepath.ToSlash(sourcePath), fixture.Contract.SourcePath) {
		return Receipt{}, fmt.Errorf("candidate %q source path does not match %q", candidateID, fixture.Contract.SourcePath)
	}
	canonicalPath := canonicalSourcePath(sourcePath)
	receipt := Receipt{
		Schema:            ReceiptSchema,
		SubjectSHA:        subjectSHA,
		CandidateID:       candidateID,
		SourcePath:        canonicalPath,
		SourceDigest:      sha256Digest(source),
		Producer:          receiptProducer,
		Consumer:          receiptConsumer,
		MetaOperation:     fixture.Contract.MetaOperation,
		ProofChoice:       fixture.Contract.ProofChoice,
		SemanticValue:     "",
		Decision:          "",
		ClaimTransitions:  []ClaimTransition{},
		CoordinateVector:  fixture.Coordinates,
		Counterexamples:   fixture.Counterexamples,
		UnknownLocations:  fixture.UnknownLocations,
		ExtensionEvidence: fixture.ExtensionEvidence,
		RepositoryWrites:  0,
		MutationAuthority: false,
	}
	receipt.FactsDigest = receiptFactsDigest(receipt)
	return sealReceipt(receipt), nil
}

func canonicalSourcePath(path string) string {
	path = filepath.ToSlash(path)
	const marker = "examples/experiment-portfolio/"
	if index := strings.Index(path, marker); index >= 0 {
		return path[index:]
	}
	return path
}

func sha256Digest(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}
