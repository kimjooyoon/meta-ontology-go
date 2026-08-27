package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"runtime"

	sandbox "github.com/kimjooyoon/meta-ontology-go/internal/meta/reflectivequerysandbox"
)

const (
	producerName      = "reflective-query-sandbox.producer"
	consumerName      = "reflective-query-sandbox.independent-verifier"
	structureActivity = "gooo://reflective-query-sandbox/activity/reflect-structure"
	claimsActivity    = "gooo://reflective-query-sandbox/activity/reflect-claims"
	metricsActivity   = "gooo://reflective-query-sandbox/activity/reflect-metrics"
	mutationActivity  = "gooo://reflective-query-sandbox/activity/attempt-mutation"
	structureTarget   = "gooo://reflective-query-sandbox/entity/program-structure"
	claimsTarget      = "gooo://reflective-query-sandbox/entity/claim-catalog"
	metricsTarget     = "gooo://reflective-query-sandbox/entity/metric-catalog"
	mutationTarget    = "gooo://reflective-query-sandbox/entity/mutation-request"
	unknownTarget     = "gooo://reflective-query-sandbox/entity/unknown-target"
)

type expectedAttempt struct {
	id, class, operation, root, relation, target, meta, proof, stage, step string
	decision, resolution, reason                                           string
	matched                                                                int
}

func main() {
	input := flag.String("input", "", "producer observation")
	source := flag.String("source", "", "Gooo source")
	subject := flag.String("subject-sha", "", "exact subject commit")
	output := flag.String("output", "", "independent receipt")
	flag.Parse()
	if *input == "" || *source == "" || *subject == "" || *output == "" {
		fail("usage: consumer -input FILE -source FILE -subject-sha SHA -output FILE")
	}
	var observation sandbox.Observation
	data, err := os.ReadFile(*input)
	if err != nil {
		fail("read observation: %v", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&observation); err != nil {
		fail("decode observation: %v", err)
	}
	if err := validateObservation(observation, *source, *subject); err != nil {
		fail("independent validation: %v", err)
	}
	receipt := buildReceipt(observation)
	receipt.Digest = receiptDigest(receipt)
	encoded, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		fail("encode receipt: %v", err)
	}
	if err := os.WriteFile(*output, append(encoded, '\n'), 0o644); err != nil {
		fail("write receipt: %v", err)
	}
	fmt.Printf("consumer verdict: %s %d/%d writes=%d mutation_authority=%t\n", receipt.Decision, receipt.Coordinates.Satisfied, receipt.Coordinates.Total, receipt.Effects.RepositoryWrites, receipt.Effects.MutationAuthority)
}

func validateObservation(value sandbox.Observation, sourcePath, subject string) error {
	if value.Schema != sandbox.Schema || value.SubjectSHA != subject || value.Producer != producerName {
		return errors.New("observation identity is not exact")
	}
	if value.Digest != observationDigest(value) {
		return errors.New("observation digest mismatch")
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	if value.Source.Path != sandbox.ExpectedSourcePath || value.Source.SourceDigest != plainDigest(source) || value.Source.GoooLines != 11 || value.Source.NodeCount != 9 || value.Source.FactCount != 8 {
		return errors.New("source snapshot is not exact")
	}
	if runtime.Version() != "go"+sandbox.ExpectedGoVersion {
		return fmt.Errorf("runtime=%s want go%s", runtime.Version(), sandbox.ExpectedGoVersion)
	}
	if value.Effects != (sandbox.Effects{}) {
		return errors.New("producer reported an effect")
	}
	if err := validateContract(value.Contract); err != nil {
		return err
	}
	if err := validateAttempts(value.Attempts, value.Source.SemanticDigest); err != nil {
		return err
	}
	return validateTransitions(value.Claims)
}

func validateContract(value sandbox.Contract) error {
	wantClasses := []sandbox.Bucket{{Name: "OUTCOME", Total: 4}, {Name: "DRIVER", Total: 4}, {Name: "GUARDRAIL", Total: 4}}
	wantProofs := []sandbox.Bucket{{Name: "FOUNDATION", Total: 4}, {Name: "COHERENCE", Total: 4}, {Name: "REGRESSION", Total: 4}}
	if value.Schema != sandbox.Schema || value.MetricID != sandbox.MetricID || value.GoVersion != sandbox.ExpectedGoVersion || value.Denominator != 12 || !reflect.DeepEqual(value.Classes, wantClasses) || !reflect.DeepEqual(value.Proofs, wantProofs) || value.ExpectedNodes != 9 || value.ExpectedFacts != 8 || value.ExpectedAttempts != 5 || value.ExpectedSafe != 3 || value.ExpectedDenied != 1 || value.ExpectedUnknown != 1 || value.ExpectedTransitions != 24 {
		return errors.New("fixed denominator contract drifted")
	}
	return nil
}

func validateAttempts(attempts []sandbox.Attempt, semanticDigest string) error {
	want := []expectedAttempt{
		{"reflect.structure", "OUTCOME", "query", structureActivity, "used", structureTarget, "query-self-structure", "FOUNDATION", "QUERY", "match-structure", "PASS", "EXACT", "EXACT_RELATION_MATCH", 1},
		{"reflect.claims", "OUTCOME", "query", claimsActivity, "used", claimsTarget, "query-self-claims", "COHERENCE", "QUERY", "match-claims", "PASS", "EXACT", "EXACT_RELATION_MATCH", 1},
		{"reflect.metrics", "OUTCOME", "query", metricsActivity, "used", metricsTarget, "query-self-metrics", "REGRESSION", "QUERY", "match-metrics", "PASS", "EXACT", "EXACT_RELATION_MATCH", 1},
		{"mutation.attempt", "OUTCOME", "mutate", mutationActivity, "set", mutationTarget, "deny-mutation-request", "FOUNDATION", "BOUNDARY", "reject-mutation-operation", "DENIED", "INVARIANT_ONLY", "READ_ONLY_QUERY_ONLY", 0},
		{"unknown.target", "GUARDRAIL", "query", metricsActivity, "used", unknownTarget, "preserve-unknown-target", "REGRESSION", "UNKNOWN", "reject-unknown-target", "UNKNOWN", "LOWER_RESOLUTION", "UNKNOWN_TARGET", 0},
	}
	if len(attempts) != len(want) {
		return fmt.Errorf("attempt denominator=%d want=%d", len(attempts), len(want))
	}
	for index, got := range attempts {
		expect := want[index]
		if got.ID != expect.id || got.Class != expect.class || got.Operation != expect.operation || got.Root != expect.root || got.Relation != expect.relation || got.Target != expect.target || got.MetaOperation != expect.meta || got.Producer != producerName || got.Consumer != consumerName || got.ProofChoice != expect.proof || got.Stage != expect.stage || got.Step != expect.step || got.Decision != expect.decision || got.Resolution != expect.resolution || got.Reason != expect.reason || got.MatchedFacts != expect.matched || got.SemanticDigestBefore != semanticDigest || got.SemanticDigestAfter != semanticDigest || got.GraphDigestBefore != got.GraphDigestAfter {
			return fmt.Errorf("attempt[%d] is not an exact boundary record", index)
		}
	}
	return nil
}

func validateTransitions(values []sandbox.ClaimTransition) error {
	if len(values) != 24 {
		return fmt.Errorf("transition denominator=%d want=24", len(values))
	}
	want := []struct{ id, class, proof, meta string }{
		{"outcome.structure", "OUTCOME", "FOUNDATION", "query-self-structure"}, {"outcome.claims", "OUTCOME", "COHERENCE", "query-self-claims"}, {"outcome.metrics", "OUTCOME", "REGRESSION", "query-self-metrics"}, {"outcome.mutation-denied", "OUTCOME", "FOUNDATION", "deny-mutation-request"},
		{"driver.semantic-snapshot", "DRIVER", "COHERENCE", "bind-semantic-snapshot"}, {"driver.query-projection", "DRIVER", "REGRESSION", "project-read-only-query-view"}, {"driver.query-receipts", "DRIVER", "FOUNDATION", "seal-query-receipts"}, {"driver.claim-ledger", "DRIVER", "COHERENCE", "transition-claim-ledger"},
		{"guardrail.unknown-closed", "GUARDRAIL", "REGRESSION", "preserve-unknown-target"}, {"guardrail.graph-unchanged", "GUARDRAIL", "FOUNDATION", "compare-query-graph-digest"}, {"guardrail.repository-write-set", "GUARDRAIL", "COHERENCE", "observe-repository-write-set"}, {"guardrail.mutation-authority", "GUARDRAIL", "REGRESSION", "bind-mutation-authority"},
	}
	previous := ""
	for index, spec := range want {
		for offset, state := range []struct{ from, to, stage, step, reason string }{{"UNRECORDED", "OPEN", "DECLARE", "register-denominator-claim", "CLAIM_REGISTERED"}, {"OPEN", "DISCHARGED", "OBSERVE", "evaluate-read-only-boundary", "OBSERVATION_DISCHARGED"}} {
			value := values[index*2+offset]
			if value.Sequence != index*2+offset+1 || value.ClaimID != spec.id || value.Class != spec.class || value.ProofChoice != spec.proof || value.MetaOperation != spec.meta || value.Producer != producerName || value.Consumer != consumerName || value.Stage != state.stage || value.Step != state.step || value.Reason != state.reason || value.From != state.from || value.To != state.to || value.PreviousDigest != previous || value.Digest != transitionDigest(value) {
				return fmt.Errorf("transition[%d] is not an append-only exact event", index*2+offset)
			}
			previous = value.Digest
		}
	}
	return nil
}

func buildReceipt(observation sandbox.Observation) sandbox.Receipt {
	notClaimed := []string{"generic Go reflection API equivalence", "runtime capability isolation beyond this process", "source completeness beyond fixed declarations", "mutation safety against a hostile process", "runtime memory and performance bounds"}
	return sandbox.Receipt{
		Schema: sandbox.ReceiptSchema, SubjectSHA: observation.SubjectSHA, MetricID: sandbox.MetricID, Decision: "PASS", Resolution: "OBSERVATION_ONLY", Reason: "READ_ONLY_REFLECTION_BOUNDARY_PROVED", Producer: observation.Producer, Consumer: consumerName, Contract: observation.Contract, Source: observation.Source, Attempts: observation.Attempts, Claims: observation.Claims, Coordinates: sandbox.Coordinates{Satisfied: 12, Total: 12, BasisPoints: 10000}, Classes: []sandbox.Score{{Name: "OUTCOME", Satisfied: 4, Total: 4}, {Name: "DRIVER", Satisfied: 4, Total: 4}, {Name: "GUARDRAIL", Satisfied: 4, Total: 4}}, Proofs: []sandbox.Score{{Name: "FOUNDATION", Satisfied: 4, Total: 4}, {Name: "COHERENCE", Satisfied: 4, Total: 4}, {Name: "REGRESSION", Satisfied: 4, Total: 4}}, Effects: sandbox.Effects{}, PromotionCreditBPS: 0, RepositoryWrites: 0, MutationAuthority: false, NotClaimed: notClaimed,
	}
}

func observationDigest(value sandbox.Observation) string {
	value.Digest = ""
	return hashJSON(value)
}

func transitionDigest(value sandbox.ClaimTransition) string {
	value.Digest = ""
	return hashJSON(value)
}

func receiptDigest(value sandbox.Receipt) string {
	value.Digest = ""
	return hashJSON(value)
}

func hashJSON(value any) string {
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func plainDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
