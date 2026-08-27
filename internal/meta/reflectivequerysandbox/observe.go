package reflectivequerysandbox

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
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

func Observe(path string, subjectSHA string) (Observation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Observation{}, fmt.Errorf("read source: %w", err)
	}
	file, diagnostics := syntax.ParseFile(path, string(data))
	if diagnostics.HasErrors() {
		return Observation{}, diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Observation{}, fmt.Errorf("lower source: %w", err)
	}
	graph, err := query.FromSemanticIR(ir)
	if err != nil {
		return Observation{}, fmt.Errorf("project query view: %w", err)
	}
	semanticDigest := ir.StableHash()
	source := Snapshot{Path: path, SourceDigest: semantic.StableHash(data), SemanticDigest: semanticDigest, NodeCount: len(graph.Nodes()), FactCount: len(graph.AllFacts()), GoooLines: countLines(data)}
	attempts := []Attempt{
		exactAttempt(graph, semanticDigest, "reflect.structure", "OUTCOME", structureActivity, structureTarget, "query-self-structure", "FOUNDATION", "QUERY", "match-structure"),
		exactAttempt(graph, semanticDigest, "reflect.claims", "OUTCOME", claimsActivity, claimsTarget, "query-self-claims", "COHERENCE", "QUERY", "match-claims"),
		exactAttempt(graph, semanticDigest, "reflect.metrics", "OUTCOME", metricsActivity, metricsTarget, "query-self-metrics", "REGRESSION", "QUERY", "match-metrics"),
		mutationAttempt(graph, semanticDigest),
		exactAttempt(graph, semanticDigest, "unknown.target", "GUARDRAIL", metricsActivity, unknownTarget, "preserve-unknown-target", "REGRESSION", "UNKNOWN", "reject-unknown-target"),
	}
	return Observation{
		Schema: Schema, SubjectSHA: subjectSHA, Contract: CanonicalContract(), Source: source,
		Attempts: attempts, Claims: buildClaimTransitions(), Effects: Effects{},
		Producer: "reflective-query-sandbox.producer",
	}, nil
}

func exactAttempt(graph *query.Graph, semanticDigest, id, class, root, target, operation, proof, stage, step string) Attempt {
	before := graph.StableHash()
	attempt := Attempt{ID: id, Class: class, Operation: "query", Root: root, Relation: "used", Target: target, MetaOperation: operation, Producer: "reflective-query-sandbox.producer", Consumer: "reflective-query-sandbox.independent-verifier", ProofChoice: proof, Stage: stage, Step: step, SemanticDigestBefore: semanticDigest, GraphDigestBefore: before}
	result, err := graph.ExactMatch(query.NewExactQuery(query.ID(root), query.Used, query.ID(target)))
	after := graph.StableHash()
	attempt.SemanticDigestAfter, attempt.GraphDigestAfter = semanticDigest, after
	if err != nil {
		attempt.Decision, attempt.Resolution, attempt.Reason = "UNKNOWN", "LOWER_RESOLUTION", "UNKNOWN_TARGET"
		if !errors.Is(err, query.ErrUnknownEndpoint) {
			attempt.Reason = "QUERY_REJECTED"
		}
		return attempt
	}
	attempt.MatchedFacts = len(result.All())
	if attempt.MatchedFacts == 1 {
		attempt.Decision, attempt.Resolution, attempt.Reason = "PASS", "EXACT", "EXACT_RELATION_MATCH"
		return attempt
	}
	attempt.Decision, attempt.Resolution, attempt.Reason = "UNKNOWN", "LOWER_RESOLUTION", "RELATION_NOT_OBSERVED"
	return attempt
}

func mutationAttempt(graph *query.Graph, semanticDigest string) Attempt {
	digest := graph.StableHash()
	return Attempt{ID: "mutation.attempt", Class: "OUTCOME", Operation: "mutate", Root: mutationActivity, Relation: "set", Target: mutationTarget, MetaOperation: "deny-mutation-request", Producer: "reflective-query-sandbox.producer", Consumer: "reflective-query-sandbox.independent-verifier", ProofChoice: "FOUNDATION", Stage: "BOUNDARY", Step: "reject-mutation-operation", Decision: "DENIED", Resolution: "INVARIANT_ONLY", Reason: "READ_ONLY_QUERY_ONLY", SemanticDigestBefore: semanticDigest, SemanticDigestAfter: semanticDigest, GraphDigestBefore: digest, GraphDigestAfter: digest}
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	return strings.Count(string(data), "\n")
}
