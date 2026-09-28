package selfimprovementtransport

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

var resolutionPolicyKeys = []string{
	"resolution-schema", "resolution-states", "resolution-causal-fields",
	"resolution-artifact-identity", "resolution-refuted-dominates-unknown",
}

func parseResolutionPolicy(ir *semantic.IR) (ResolutionPolicy, error) {
	markers, err := resolutionPolicyMarkers(ir)
	if err != nil {
		return ResolutionPolicy{}, err
	}
	policy, err := buildResolutionPolicy(markers)
	if err != nil {
		return ResolutionPolicy{}, err
	}
	if err := policy.Validate(); err != nil {
		return ResolutionPolicy{}, err
	}
	return policy, nil
}

func resolutionPolicyMarkers(ir *semantic.IR) (map[string][]string, error) {
	markers := map[string][]string{}
	if ir != nil {
		for _, node := range ir.Graph.Nodes() {
			if node.Kind != semantic.Activity || node.Name != "ResolveConsumerSubject" || node.ValueProgram == "" {
				continue
			}
			for part := range strings.SplitSeq(node.ValueProgram, ";") {
				fields := strings.SplitN(part, "=", 2)
				if len(fields) == 2 && strings.TrimSpace(fields[0]) != "" {
					key := strings.TrimSpace(fields[0])
					markers[key] = append(markers[key], strings.TrimSpace(fields[1]))
				}
			}
		}
	}
	for _, key := range resolutionPolicyKeys {
		if len(markers[key]) != 1 {
			return nil, fmt.Errorf("transport resolution policy marker %q is not unique", key)
		}
	}
	return markers, nil
}

func buildResolutionPolicy(markers map[string][]string) (ResolutionPolicy, error) {
	policy := ResolutionPolicy{
		Schema:                  markers["resolution-schema"][0],
		States:                  splitResolutionCSV(markers["resolution-states"][0]),
		CausalFields:            splitResolutionCSV(markers["resolution-causal-fields"][0]),
		ArtifactIdentity:        splitResolutionCSV(markers["resolution-artifact-identity"][0]),
		RefutedDominatesUnknown: markers["resolution-refuted-dominates-unknown"][0] == "true",
		Transitions:             append([]string(nil), markers["resolution-transition"]...),
		Metrics:                 map[string]int{},
	}
	if err := appendResolutionPolicyMetrics(&policy, markers["resolution-metric"]); err != nil {
		return ResolutionPolicy{}, err
	}
	if err := appendResolutionPolicyCases(&policy, markers["resolution-case"]); err != nil {
		return ResolutionPolicy{}, err
	}
	return policy, nil
}

func appendResolutionPolicyMetrics(policy *ResolutionPolicy, values []string) error {
	for _, value := range values {
		fields := strings.SplitN(value, "|", 2)
		if len(fields) != 2 || fields[0] == "" {
			return errors.New("transport resolution metric marker is invalid")
		}
		var parsed int
		if _, err := fmt.Sscanf(fields[1], "%d", &parsed); err != nil {
			return fmt.Errorf("transport resolution metric %q is invalid", fields[0])
		}
		if _, exists := policy.Metrics[fields[0]]; exists {
			return fmt.Errorf("transport resolution metric %q is duplicated", fields[0])
		}
		policy.Metrics[fields[0]] = parsed
	}
	return nil
}

func appendResolutionPolicyCases(policy *ResolutionPolicy, values []string) error {
	for _, value := range values {
		fields := strings.Split(value, "|")
		if len(fields) != 6 {
			return errors.New("transport resolution case marker is invalid")
		}
		policy.Cases = append(policy.Cases, ResolutionCase{ID: fields[0], State: fields[1], Stage: fields[2], Step: fields[3], UnknownClass: fields[4], Reason: fields[5]})
	}
	policy.CaseDenominator = len(policy.Cases)
	for _, item := range policy.Cases {
		switch item.State {
		case ResolutionClosed:
			policy.ClosedCases++
		case ResolutionUnknown:
			policy.UnknownCases++
		case ResolutionRefuted:
			policy.RefutedCases++
		}
	}
	return nil
}

func (policy ResolutionPolicy) Validate() error {
	if policy.Schema != ResolutionPolicySchema ||
		!sameResolutionStrings(policy.States, []string{ResolutionClosed, ResolutionUnknown, ResolutionRefuted}) ||
		!sameResolutionStrings(policy.CausalFields, []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"}) ||
		!sameResolutionStrings(policy.Transitions, []string{"producer-declaration>transport-index", "transport-index>consumer-resolution"}) ||
		!sameResolutionStrings(policy.ArtifactIdentity, []string{"artifact_id", "artifact_name", "artifact_digest", "producer_run_id", "producer_run_attempt", "producer_subject_sha", "producer_declaration_digest", "producer_payload_name", "producer_payload_digest"}) ||
		!policy.RefutedDominatesUnknown || policy.CaseDenominator != ResolutionCaseCount ||
		policy.ClosedCases != 3 || policy.UnknownCases != 3 || policy.RefutedCases != 4 ||
		!sameResolutionMetrics(policy.Metrics, canonicalResolutionMetrics()) {
		return errors.New("transport resolution policy is not canonical")
	}
	seen := map[string]bool{}
	counts := map[string]int{}
	for _, item := range policy.Cases {
		if item.ID == "" || seen[item.ID] || item.Stage == "" || item.Step == "" || item.Reason == "" {
			return errors.New("transport resolution case table is invalid")
		}
		if item.State == ResolutionUnknown && item.UnknownClass == "" {
			return errors.New("transport resolution UNKNOWN case has no class")
		}
		if item.State != ResolutionClosed && item.State != ResolutionUnknown && item.State != ResolutionRefuted {
			return errors.New("transport resolution case state is invalid")
		}
		seen[item.ID] = true
		counts[item.State]++
	}
	if counts[ResolutionClosed] != 3 || counts[ResolutionUnknown] != 3 || counts[ResolutionRefuted] != 4 {
		return errors.New("transport resolution case counts are not canonical")
	}
	if !sameResolutionCases(policy.Cases, canonicalResolutionCases()) {
		return errors.New("transport resolution case table is not canonical")
	}
	return nil
}

func canonicalResolutionCases() []ResolutionCase {
	return []ResolutionCase{
		{ID: "EXACT_PRODUCER_DECLARATION", State: ResolutionClosed, Stage: "CONSUME", Step: "resolve-producer-subject", Reason: "EXACT_PRODUCER_SUBJECT_PAYLOAD_MATCH"},
		{ID: "EXACT_PAYLOAD_SUBJECT", State: ResolutionClosed, Stage: "CONSUME", Step: "resolve-producer-subject", Reason: "EXACT_PAYLOAD_SUBJECT_MATCH"},
		{ID: "EXACT_NONCURRENT_SUBJECT", State: ResolutionClosed, Stage: "CONSUME", Step: "resolve-producer-subject", Reason: "EXACT_NONCURRENT_SUBJECT_MATCH"},
		{ID: "EXPIRED_ARTIFACT", State: ResolutionUnknown, Stage: "LOCATE", Step: "read-artifact-metadata", UnknownClass: "EXPIRED", Reason: "ARTIFACT_EXPIRED"},
		{ID: "MISSING_PRODUCER_DECLARATION", State: ResolutionUnknown, Stage: "CONSUME", Step: "resolve-producer-declaration", UnknownClass: "DIRECT_MISSING", Reason: "PRODUCER_DECLARATION_MISSING"},
		{ID: "MISSING_PAYLOAD", State: ResolutionUnknown, Stage: "CONSUME", Step: "resolve-payload", UnknownClass: "DIRECT_MISSING", Reason: "PRODUCER_PAYLOAD_MISSING"},
		{ID: "DUPLICATE_DECLARATION", State: ResolutionRefuted, Stage: "CONSUME", Step: "resolve-producer-declaration", Reason: "DUPLICATE_PRODUCER_DECLARATION"},
		{ID: "PAYLOAD_SUBJECT_CONTRADICTION", State: ResolutionRefuted, Stage: "CONSUME", Step: "resolve-payload", Reason: "PRODUCER_PAYLOAD_SUBJECT_MISMATCH"},
		{ID: "PAYLOAD_DIGEST_MISMATCH", State: ResolutionRefuted, Stage: "CONSUME", Step: "resolve-payload", Reason: "PRODUCER_PAYLOAD_DIGEST_MISMATCH"},
		{ID: "REPOSITORY_WORKFLOW_CONTRADICTION", State: ResolutionRefuted, Stage: "CONSUME", Step: "resolve-producer-identity", Reason: "PRODUCER_REPOSITORY_WORKFLOW_MISMATCH"},
	}
}

func canonicalResolutionMetrics() map[string]int {
	return map[string]int{
		"active_root_before": 1, "active_root_after": 0,
		"exact_resolutions_before": 0, "exact_resolutions_after": 1,
		"unknown_six_field_before": 0, "unknown_six_field_after": 3,
		"refuted_contradictions_before": 0, "refuted_contradictions_after": 4,
		"fallback_accepted_before": 0, "fallback_accepted_after": 0,
		"artifact_instances_before": 1, "artifact_instances_after": 1,
		"artifact_types_before": 1, "artifact_types_after": 1,
		"independent_replay_comparisons_before": 0, "independent_replay_comparisons_after": 1,
	}
}

func splitResolutionCSV(value string) []string {
	result := []string{}
	for item := range strings.SplitSeq(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func sameResolutionStrings(left, right []string) bool {
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

func sameResolutionMetrics(left, right map[string]int) bool {
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

func sameResolutionCases(left, right []ResolutionCase) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range right {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
