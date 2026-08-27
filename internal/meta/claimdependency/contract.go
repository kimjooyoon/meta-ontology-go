package claimdependency

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	ValidatorContractSchema = "gooo.meta.claim-dependency-validator-contract/v1"
	FailureReceiptSchema    = "gooo.meta.claim-dependency-failure-receipt/v1"
	failureProcedure        = "CI_NONZERO_EXIT_FAILURE_ANTECEDENT_V1"
)

type sourceGraph struct {
	IR    semantic.IR
	Graph Graph
}

// GraphFromSource reconstructs executable propositions and typed dependency
// edges from syntax.ParseFile -> bidir.Lower canonical IR. Case names do not
// participate in the graph or in the final state decision.
func GraphFromSource(source []byte, sourcePath string) (Graph, error) {
	parsed, err := graphFromSource(source, sourcePath)
	if err != nil {
		return Graph{}, err
	}
	return parsed.Graph, nil
}

func graphFromSource(source []byte, sourcePath string) (sourceGraph, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return sourceGraph{}, fmt.Errorf("source parse failed: %s", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceGraph{}, fmt.Errorf("source lower failed: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return sourceGraph{}, fmt.Errorf("canonical IR invalid: %w", err)
	}
	activities := map[string]semantic.Node{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Activity {
			activities[node.Name] = node
		}
	}
	generatedBy, usedBy := map[string]string{}, map[string][]string{}
	for _, fact := range ir.Graph.AllFacts() {
		switch fact.Predicate {
		case semantic.WasGeneratedBy:
			generatedBy[fact.Subject.String()] = fact.Object.String()
		case semantic.Used:
			usedBy[fact.Subject.String()] = append(usedBy[fact.Subject.String()], fact.Object.String())
		}
	}
	claims, activityIndex := make([]Claim, 0, ClaimTotal), map[string]int{}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		node, ok := activities[activity.Name]
		if !ok || node.ValueProgram == "" {
			return sourceGraph{}, fmt.Errorf("activity %q is not a semantic value claim", activity.Name)
		}
		inputs := append([]string(nil), usedBy[node.ID.String()]...)
		sort.Strings(inputs)
		output := ""
		for entityID, activityID := range generatedBy {
			if activityID == node.ID.String() {
				output = entityID
				break
			}
		}
		artifact := "gooo://claim-dependency/artifact/" + strings.ToLower(activity.Name)
		proposition := normalizedProposition(node.ID.String(), inputs, output, artifact, node.ValueProgram)
		activityIndex[node.ID.String()] = len(claims)
		claims = append(claims, Claim{Ordinal: len(claims) + 1, Axis: strings.ToLower(activity.Name), ClaimID: node.ID.String(), ActivityID: node.ID.String(), ActivityName: activity.Name, Proposition: proposition, PropositionDigest: digestBytes([]byte(proposition)), Target: TargetAddress{Inputs: inputs, Output: output, Artifact: artifact}, ValueProgram: node.ValueProgram, Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperationID, ProofChoice: ProofChoice, Coordinate: Coordinate{Stage: "CLAIM", Step: activity.Name, Reason: "NORMALIZED_EXECUTION_PROPOSITION"}})
	}
	if len(claims) != ClaimTotal {
		return sourceGraph{}, fmt.Errorf("source must contain exactly %d activity claims, got %d", ClaimTotal, len(claims))
	}
	digests := map[string]bool{}
	for _, claim := range claims {
		digests[claim.PropositionDigest] = true
	}
	if len(digests) != ClaimTotal {
		return sourceGraph{}, fmt.Errorf("claims do not contain %d distinct observed propositions", ClaimTotal)
	}
	type edgeCandidate struct {
		from, to int
		kind     EdgeKind
	}
	candidates := make([]edgeCandidate, 0, EdgeTotal)
	for downstreamID, entities := range usedBy {
		to, ok := activityIndex[downstreamID]
		if !ok {
			continue
		}
		for _, entityID := range entities {
			upstreamID, ok := generatedBy[entityID]
			if !ok {
				continue
			}
			from, ok := activityIndex[upstreamID]
			if !ok || from == to {
				continue
			}
			kind, ok := edgeKind(claims[to].ValueProgram)
			if !ok {
				return sourceGraph{}, fmt.Errorf("activity %q does not declare a typed dependency edge", claims[to].ActivityName)
			}
			candidates = append(candidates, edgeCandidate{from: from, to: to, kind: kind})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].from != candidates[j].from {
			return candidates[i].from < candidates[j].from
		}
		if candidates[i].to != candidates[j].to {
			return candidates[i].to < candidates[j].to
		}
		return candidates[i].kind < candidates[j].kind
	})
	if len(candidates) != EdgeTotal {
		return sourceGraph{}, fmt.Errorf("semantic activity relations must reconstruct exactly %d edges, got %d", EdgeTotal, len(candidates))
	}
	edges := make([]Edge, len(candidates))
	for i, c := range candidates {
		edges[i] = Edge{EdgeID: fmt.Sprintf("E%02d", i+1), FromClaimID: claims[c.from].ClaimID, ToClaimID: claims[c.to].ClaimID, Kind: c.kind, ActivationPredicate: activationPredicate(c.kind), SemanticBasis: "prov:wasGeneratedBy + prov:used + source-derived value-program edge predicate"}
	}
	graph := Graph{Schema: GraphSchema, Authority: "CANONICAL_IR_FROM_SYNTAX_PARSE_AND_BIDIR_LOWER", Completeness: "CLOSED_WORLD_SOURCE_RECONSTRUCTED", CanonicalIRDigest: prefixedDigest(ir.StableHash()), NodeTotal: len(claims), EdgeTotal: len(edges), Nodes: claims, Edges: edges}
	graph.Digest, err = graphDigest(graph)
	if err != nil {
		return sourceGraph{}, err
	}
	return sourceGraph{IR: ir, Graph: graph}, nil
}

func activationPredicate(kind EdgeKind) string {
	switch kind {
	case Requires:
		return "UPSTREAM_DISCHARGED_AND_LOCAL_EVIDENCE"
	case Contradicts:
		return "UPSTREAM_DISCHARGED_PROPOSITION"
	case FailureEntailment:
		return "OBSERVED_FAILURE_ANTECEDENT_AND_UPSTREAM_REFUTED"
	default:
		return "UPSTREAM_UNKNOWN_BLOCKS_ONLY"
	}
}

func normalizedProposition(activityID string, inputs []string, output, artifact, value string) string {
	return fmt.Sprintf("execute(activity=%s,inputs=[%s],output=%s,artifact=%s,value=%s)", activityID, strings.Join(inputs, ","), output, artifact, value)
}

func edgeKind(program string) (EdgeKind, bool) {
	const prefix = "claim.edge:"
	if !strings.HasPrefix(program, prefix) {
		return "", false
	}
	value := strings.TrimPrefix(program, prefix)
	if pipe := strings.IndexByte(value, '|'); pipe >= 0 {
		value = value[:pipe]
	}
	switch value {
	case "supports":
		return Supports, true
	case "requires":
		return Requires, true
	case "contradicts":
		return Contradicts, true
	case "failure-entailment":
		return FailureEntailment, true
	}
	return "", false
}

func TruthTableCases() []TruthTableCase {
	return []TruthTableCase{
		{Schema: TruthTableSchema, CaseID: "SUPPORTS-POSITIVE", Kind: Supports, Direction: "direction-matching", UpstreamState: "DISCHARGED", LocalPredicate: "EVIDENCE_ACCEPTED", ExpectedState: "OPEN", Positive: true, SemanticBasis: "support never discharges or refutes by itself"},
		{Schema: TruthTableSchema, CaseID: "SUPPORTS-REVERSED", Kind: Supports, Direction: "reversed-direction", UpstreamState: "REFUTED", LocalPredicate: "UNKNOWN", ExpectedState: "OPEN", Positive: false, SemanticBasis: "support does not refute"},
		{Schema: TruthTableSchema, CaseID: "REQUIRES-POSITIVE", Kind: Requires, Direction: "direction-matching", UpstreamState: "DISCHARGED", LocalPredicate: "EVIDENCE_ACCEPTED", ExpectedState: "DISCHARGED", Positive: true, SemanticBasis: "upstream and local requirement hold"},
		{Schema: TruthTableSchema, CaseID: "REQUIRES-UNKNOWN", Kind: Requires, Direction: "direction-matching", UpstreamState: "DISCHARGED", LocalPredicate: "UNKNOWN", ExpectedState: "OPEN", Positive: true, SemanticBasis: "local requirement evidence is unknown"},
		{Schema: TruthTableSchema, CaseID: "CONTRADICTS-POSITIVE", Kind: Contradicts, Direction: "direction-matching", UpstreamState: "DISCHARGED", LocalPredicate: "UNKNOWN", ExpectedState: "REFUTED", Positive: true, ContradictionObserved: true, SemanticBasis: "established proposition and observed contradiction agree in the declared direction"},
		{Schema: TruthTableSchema, CaseID: "CONTRADICTS-REVERSED", Kind: Contradicts, Direction: "reversed-direction", UpstreamState: "DISCHARGED", LocalPredicate: "UNKNOWN", ExpectedState: "OPEN", Positive: false, ContradictionObserved: true, SemanticBasis: "reversed contradiction direction cannot refute"},
		{Schema: TruthTableSchema, CaseID: "FAILURE_ENTAILMENT-POSITIVE", Kind: FailureEntailment, Direction: "direction-matching", UpstreamState: "REFUTED", LocalPredicate: "UNKNOWN", ExpectedState: "REFUTED", Positive: true, FailureAntecedentObserved: true, SemanticBasis: "declared failure antecedent is observed and entails target failure"},
		{Schema: TruthTableSchema, CaseID: "FAILURE_ENTAILMENT-UNKNOWN", Kind: FailureEntailment, Direction: "direction-matching", UpstreamState: "REFUTED", LocalPredicate: "UNKNOWN", ExpectedState: "OPEN", Positive: true, FailureAntecedentObserved: false, SemanticBasis: "upstream refutation alone does not activate failure entailment"},
	}
}

type relationOutcome string

const (
	relationOpen       relationOutcome = "OPEN"
	relationDischarged relationOutcome = "DISCHARGED"
	relationRefuted    relationOutcome = "REFUTED"
)

// edgeRelation is the single state relation used by both the executable
// truth-table cases and the graph classifier. directionMatches is explicit so
// the negative cases exercise the same relation with a reversed edge.
func edgeRelation(kind EdgeKind, upstreamState string, local ObservationPredicate, directionMatches, contradictionObserved, failureAntecedentObserved bool) relationOutcome {
	if !directionMatches {
		return relationOpen
	}
	switch kind {
	case Requires:
		if upstreamState == "DISCHARGED" && local == ObservationEvidence {
			return relationDischarged
		}
	case Contradicts:
		if upstreamState == "DISCHARGED" && contradictionObserved {
			return relationRefuted
		}
	case FailureEntailment:
		if upstreamState == "REFUTED" && failureAntecedentObserved {
			return relationRefuted
		}
	case Supports:
		// SUPPORTS can block an OPEN claim when its upstream is unresolved,
		// but it can never discharge or refute the target by itself.
	}
	return relationOpen
}

// validateTruthTable executes the closed edge algebra over each positive and
// negative case before the table is emitted. The table is therefore an
// executable counterexample set, not a list of unchecked labels.
func validateTruthTable(cases []TruthTableCase) error {
	if len(cases) != 2*len(EdgeKinds()) {
		return fmt.Errorf("truth table has %d cases, want %d", len(cases), 2*len(EdgeKinds()))
	}
	seen := map[EdgeKind]int{}
	for _, test := range cases {
		actual := edgeRelation(test.Kind, test.UpstreamState, ObservationPredicate(test.LocalPredicate), test.Positive, test.ContradictionObserved, test.FailureAntecedentObserved)
		if string(actual) != test.ExpectedState {
			return fmt.Errorf("truth table case %q computed %s, expected %s", test.CaseID, actual, test.ExpectedState)
		}
		seen[test.Kind]++
	}
	for _, kind := range EdgeKinds() {
		if seen[kind] != 2 {
			return fmt.Errorf("truth table edge kind %s has %d cases", kind, seen[kind])
		}
	}
	return nil
}

func graphDigest(graph Graph) (string, error)       { graph.Digest = ""; return digestJSON(graph) }
func receiptDigest(receipt Receipt) (string, error) { receipt.Digest = ""; return digestJSON(receipt) }
func evidenceReceiptDigest(receipt EvidenceReceipt) (string, error) {
	receipt.Digest = ""
	return digestJSON(receipt)
}
func observationReceiptDigest(receipt ObservationReceipt) (string, error) {
	receipt.Digest = ""
	return digestJSON(receipt)
}
func observationBundleDigest(bundle ObservationBundle) (string, error) {
	bundle.Digest = ""
	// Profile is a fixture label only.  It is deliberately excluded from the
	// semantic digest so relabelling the same raw observations cannot alter a
	// decision.
	bundle.Profile = ""
	return digestJSON(bundle)
}
func failureReceiptDigest(receipt FailureReceipt) (string, error) {
	receipt.Digest = ""
	return digestJSON(receipt)
}
func readValidatorContract(path string) (ValidatorContract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ValidatorContract{}, fmt.Errorf("validator contract: %w", err)
	}
	var contract ValidatorContract
	if err := json.Unmarshal(data, &contract); err != nil {
		return ValidatorContract{}, fmt.Errorf("validator contract decode: %w", err)
	}
	if err := validateValidatorContract(contract); err != nil {
		return ValidatorContract{}, err
	}
	return contract, nil
}
func validateValidatorContract(contract ValidatorContract) error {
	if contract.Schema != ValidatorContractSchema || contract.ContractID == "" || contract.ExpectedArtifactPath == "" || contract.ExpectedArtifactDigest == "" || len(contract.Claims) != ClaimTotal {
		return fmt.Errorf("validator contract identity or denominator is invalid")
	}
	seen := map[string]bool{}
	targets := map[string]bool{}
	for _, claim := range contract.Claims {
		if claim.ActivityName == "" || claim.ExpectedValueProgram == "" || claim.ExpectedTarget.Artifact == "" || claim.ExpectedTarget.Output == "" || seen[claim.ActivityName] {
			return fmt.Errorf("validator contract claim material is invalid")
		}
		seen[claim.ActivityName] = true
		targetKey := strings.Join(claim.ExpectedTarget.Inputs, ",") + "|" + claim.ExpectedTarget.Output + "|" + claim.ExpectedTarget.Artifact
		if targets[targetKey] {
			return fmt.Errorf("validator contract targets are not distinct")
		}
		targets[targetKey] = true
	}
	if len(seen) != ClaimTotal {
		return fmt.Errorf("validator contract claims are not distinct")
	}
	return nil
}
func contractClaim(contract ValidatorContract, activityName string) (ValidatorClaim, bool) {
	for _, claim := range contract.Claims {
		if claim.ActivityName == activityName {
			return claim, true
		}
	}
	return ValidatorClaim{}, false
}
func claimEvidenceDigest(value EvidenceClaim) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}
func capabilityDigest(value CapabilityEvidence) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}
func transitionDigest(transition Transition) (string, error) {
	transition.TransitionDigest = ""
	return digestJSON(transition)
}
func digestJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// DigestBytesForObservation is the provider boundary's byte digest primitive.
// The observer uses it without parsing source or importing the evidence
// classifier's implementation details.
func DigestBytesForObservation(data []byte) string { return digestBytes(data) }
func prefixedDigest(raw string) string {
	if strings.HasPrefix(raw, "sha256:") {
		return raw
	}
	return "sha256:" + raw
}
func EdgeKinds() []EdgeKind { return []EdgeKind{Supports, Requires, Contradicts, FailureEntailment} }
