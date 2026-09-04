package publicworkflowlineage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type markerValues map[string][]string

func Load(filename string, source []byte) (Policy, error) {
	if len(source) == 0 {
		return Policy{}, errors.New("workflow lineage source is empty")
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return Policy{}, errors.New("workflow lineage source has syntax diagnostics")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Policy{}, fmt.Errorf("lower workflow lineage source: %w", err)
	}
	if ir.Package != "cieffortobservation" || ir.Namespace.String() != "cieffortobservation" {
		return Policy{}, errors.New("workflow lineage source is not the canonical CI effort observation source")
	}
	markers := markerValues{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.Name != "ObserveCIWorkflowWindow" || node.ValueProgram == "" {
			continue
		}
		for key, values := range parseMarkers(node.ValueProgram) {
			markers[key] = append(markers[key], values...)
		}
	}
	policy := Policy{
		Schema:                  PolicySchema,
		EvaluatorSchema:         EvaluatorSchema,
		SourceDigest:            cache.HashBytes(source).String(),
		SemanticDigest:          ir.StableHash(),
		EvaluatorDigest:         evaluatorDigest(),
		Package:                 ir.Package,
		Namespace:               ir.Namespace.String(),
		Activity:                "ObserveCIWorkflowWindow",
		Name:                    first(markers, "partial-lineage-policy"),
		SourceWorkflow:          first(markers, "partial-lineage-source-workflow"),
		ConsumerWorkflow:        first(markers, "partial-lineage-consumer-workflow"),
		SourceIdentity:          splitCSV(first(markers, "partial-lineage-source-identity")),
		ArtifactIdentityFields:  splitCSV(first(markers, "partial-lineage-artifact-identity")),
		ConsumerIdentity:        splitCSV(first(markers, "partial-lineage-consumer-identity")),
		LineageStates:           splitCSV(first(markers, "partial-lineage-lineage-states")),
		CausalFields:            splitCSV(first(markers, "partial-lineage-causal-fields")),
		LineageEdges:            append([]string(nil), markers["partial-lineage-edge"]...),
		RefutedDominatesUnknown: first(markers, "partial-lineage-refuted-dominates-unknown") == "true",
		Metrics:                 parseMetrics(markers["partial-lineage-metric"]),
		Cases:                   parseCases(markers["partial-lineage-case"]),
	}
	if err := policy.Validate(); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func (policy Policy) Validate() error {
	if policy.Schema != PolicySchema || policy.EvaluatorSchema != EvaluatorSchema || policy.Name != PolicyName ||
		policy.Package != "cieffortobservation" || policy.Namespace != "cieffortobservation" || policy.Activity != "ObserveCIWorkflowWindow" ||
		policy.SourceWorkflow != "CI" || policy.ConsumerWorkflow != "CI effort observation" || !policy.RefutedDominatesUnknown ||
		len(policy.CausalFields) != 6 || len(policy.LineageEdges) != LineageEdgeCount || len(policy.Cases) != CaseCount ||
		policy.SourceDigest == "" || policy.SemanticDigest == "" || policy.EvaluatorDigest == "" {
		return errors.New("workflow lineage policy identity or denominator is invalid")
	}
	wantFields := []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"}
	if !sameStrings(policy.SourceIdentity, []string{"workflow", "run_id", "run_attempt", "subject_sha", "subject_ref"}) || !sameStrings(policy.ArtifactIdentityFields, []string{"name", "id", "digest", "subject_sha"}) || !sameStrings(policy.ConsumerIdentity, []string{"workflow", "run_id", "run_attempt", "subject_sha", "ref"}) || !sameStrings(policy.LineageStates, []string{StateExact, StateStale, StateDirectMissing, StateMismatch, StateTampered, StateCurrentDevFallback}) {
		return errors.New("workflow lineage identity or state fields are not canonical")
	}
	if !sameStrings(policy.CausalFields, wantFields) {
		return errors.New("workflow lineage causal fields are not canonical")
	}
	wantEdges := []string{"consumer>source-run", "source-run>subject", "source-run>artifact", "artifact>attempt", "artifact>digest", "consumer>consumer-subject"}
	if !sameStrings(policy.LineageEdges, wantEdges) {
		return errors.New("workflow lineage edges are not canonical")
	}
	wantMetrics := map[string]int{"stale_misattributed_before": 2, "stale_misattributed_after": 0, "stale_unknown": 2, "exact_subject_bindings": 2, "unknown_classifications": 2, "mismatch_detections": 2, "fallback_attempts": 1, "fallback_accepted": 0, "fallback_rejected": 1, "source_artifact_resolutions": 2, "source_receipts": SourceReceiptCount, "consumer_receipts": ConsumerReceiptCount, "evidence_artifacts": EvidenceArtifactCount}
	if !sameMetrics(policy.Metrics, wantMetrics) {
		return errors.New("workflow lineage metrics are not canonical")
	}
	counts := map[string]int{}
	seen := map[string]bool{}
	for _, item := range policy.Cases {
		if item.ID == "" || seen[item.ID] || item.Decision == "" || item.LineageState == "" || item.SourceSubject == "" || item.SourceRunID <= 0 {
			return errors.New("workflow lineage case table is invalid")
		}
		if item.Decision == DecisionUnknown && item.UnknownClass == "" {
			return errors.New("workflow lineage UNKNOWN case has no class")
		}
		seen[item.ID] = true
		counts[item.Decision]++
	}
	if counts[DecisionClosed] != ClosedCaseCount || counts[DecisionUnknown] != UnknownCaseCount || counts[DecisionRefuted] != RefutedCaseCount {
		return errors.New("workflow lineage cases are not 2/2/2")
	}
	return nil
}

func evaluatorDigest() string {
	return cache.HashBytes([]byte(EvaluatorSchema + "\x00exact-subject-run-attempt-artifact-v1")).String()
}

func parseMarkers(value string) markerValues {
	result := markerValues{}
	for part := range strings.SplitSeq(value, ";") {
		fields := strings.SplitN(part, "=", 2)
		if len(fields) == 2 && strings.TrimSpace(fields[0]) != "" {
			key := strings.TrimSpace(fields[0])
			result[key] = append(result[key], strings.TrimSpace(fields[1]))
		}
	}
	return result
}

func first(values markerValues, key string) string {
	if len(values[key]) == 0 {
		return ""
	}
	return values[key][0]
}

func parseCases(values []string) []CaseSpec {
	result := make([]CaseSpec, 0, len(values))
	for _, value := range values {
		fields := strings.Split(value, "|")
		if len(fields) != 6 {
			continue
		}
		runID, _ := strconv.ParseInt(fields[5], 10, 64)
		result = append(result, CaseSpec{ID: fields[0], Decision: fields[1], LineageState: fields[2], UnknownClass: fields[3], SourceSubject: fields[4], SourceRunID: runID})
	}
	return result
}

func parseMetrics(values []string) map[string]int {
	result := map[string]int{}
	for _, value := range values {
		fields := strings.SplitN(value, "|", 2)
		if len(fields) != 2 {
			continue
		}
		parsed, _ := strconv.Atoi(fields[1])
		result[fields[0]] = parsed
	}
	return result
}

func splitCSV(value string) []string {
	result := []string{}
	for item := range strings.SplitSeq(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sameMetrics(left, right map[string]int) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range right {
		if left[key] != value {
			return false
		}
	}
	return true
}
