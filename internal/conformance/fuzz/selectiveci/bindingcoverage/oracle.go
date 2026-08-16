package bindingcoverage

import (
	"encoding/hex"
	"sort"
	"strings"
)

const (
	SchemaV1 = "binding-coverage/v1"

	DecisionExact      = "EXACT"
	DecisionIncomplete = "INCOMPLETE"
	DecisionUnknown    = "UNKNOWN"
)

const (
	KindExactValue    = "EXACT_VALUE"
	KindExactDigest   = "EXACT_DIGEST"
	KindSetEqual      = "SET_EQUAL"
	KindDerivedDigest = "DERIVED_DIGEST"

	PolarityMatch    = "MATCH"
	PolarityMismatch = "MISMATCH"
)

// Input is the complete binding-coverage observation. It contains no
// execution or CI authorization fields because those concerns are out of
// scope for this oracle.
type Input struct {
	Schema           string       `json:"schema"`
	SnapshotDigest   string       `json:"snapshot_digest"`
	ExpectedDigest   string       `json:"expected_snapshot_digest"`
	Precedence       []Precedence `json:"precedence"`
	RequiredBindings []Binding    `json:"required_bindings"`
	Partitions       []Partition  `json:"partitions"`
}

type Precedence struct {
	Rank   int64  `json:"rank"`
	Stage  string `json:"stage"`
	Reason string `json:"reason"`
}

type Binding struct {
	ID             string `json:"id"`
	FromFieldID    string `json:"from_field_id"`
	ToFieldID      string `json:"to_field_id"`
	Kind           string `json:"kind"`
	ExpectedStage  string `json:"expected_stage"`
	ExpectedReason string `json:"expected_reason"`
}

type Partition struct {
	BindingID string `json:"binding_id"`
	Polarity  string `json:"polarity"`
	Stage     string `json:"stage"`
	Reason    string `json:"reason"`
}

// Vector is the normalized semantic result. The two authorization values are
// deliberately fixed false: authorization is not part of this partition.
type Vector struct {
	Decision               string   `json:"decision"`
	Reason                 string   `json:"reason"`
	RequiredBindingCount   int64    `json:"required_binding_count"`
	PartitionCount         int64    `json:"partition_count"`
	EndpointReferenceCount int64    `json:"endpoint_reference_count"`
	InputBytes             int64    `json:"input_bytes"`
	WorkUnits              int64    `json:"work_units"`
	MissingMatch           []string `json:"missing_match"`
	MissingMismatch        []string `json:"missing_mismatch"`
	ExecutionAuthorized    bool     `json:"execution_authorized"`
	CIAuthorized           bool     `json:"ci_authorized"`
}

type Result struct {
	Vector
	CanonicalDigest string `json:"canonical_digest"`
}

func Evaluate(input Input) Result {
	vector, overflow := baseVector(input)
	if overflow {
		return finish(vector, DecisionUnknown, "COUNT_OVERFLOW")
	}
	if reason := validateInput(input); reason != "" {
		return finish(vector, DecisionUnknown, reason)
	}
	if len(input.RequiredBindings) == 0 {
		return finish(vector, DecisionIncomplete, "ZERO_DENOMINATOR")
	}
	missingMatch, missingMismatch := missingEvidence(input)
	vector.MissingMatch = missingMatch
	vector.MissingMismatch = missingMismatch
	if len(missingMatch) != 0 || len(missingMismatch) != 0 {
		return finish(vector, DecisionIncomplete, incompleteReason(missingMatch, missingMismatch))
	}
	return finish(vector, DecisionExact, "COMPLETE")
}

func baseVector(input Input) (Vector, bool) {
	endpointCount, overflow := endpointReferenceCount(input.RequiredBindings)
	work, workOverflow := safeAdd(int64(len(input.RequiredBindings)), int64(len(input.Partitions)))
	overflow = overflow || workOverflow
	if !overflow {
		work, overflow = safeAdd(work, int64(endpointCount))
	}
	canonical, marshalErr := canonicalInputJSON(input)
	if marshalErr != nil {
		overflow = true
	}
	return Vector{
		RequiredBindingCount:   int64(len(input.RequiredBindings)),
		PartitionCount:         int64(len(input.Partitions)),
		EndpointReferenceCount: endpointCount,
		InputBytes:             int64(len(canonical)),
		WorkUnits:              work,
		MissingMatch:           []string{},
		MissingMismatch:        []string{},
		ExecutionAuthorized:    false,
		CIAuthorized:           false,
	}, overflow
}

func validateInput(input Input) string {
	if input.Schema != SchemaV1 {
		return "UNKNOWN_SCHEMA"
	}
	if !validDigest(input.SnapshotDigest) || input.SnapshotDigest != input.ExpectedDigest {
		return "STALE_OR_BAD_DIGEST"
	}
	precedence, reason := validatePrecedence(input.Precedence)
	if reason != "" {
		return reason
	}
	bindings, reason := validateBindings(input.RequiredBindings, precedence)
	if reason != "" {
		return reason
	}
	return validatePartitions(input.Partitions, bindings, precedence)
}

func validatePrecedence(entries []Precedence) (map[string]bool, string) {
	seenRanks := make(map[int64]bool, len(entries))
	seenPairs := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry.Rank < 0 || !validToken(entry.Stage, "stage:") || !validToken(entry.Reason, "reason:") {
			return nil, "UNKNOWN_PRECEDENCE"
		}
		pair := evidenceKey(entry.Stage, entry.Reason)
		if seenRanks[entry.Rank] || seenPairs[pair] {
			return nil, "UNKNOWN_PRECEDENCE"
		}
		seenRanks[entry.Rank] = true
		seenPairs[pair] = true
	}
	return seenPairs, ""
}

func validateBindings(bindings []Binding, precedence map[string]bool) (map[string]Binding, string) {
	seen := make(map[string]Binding, len(bindings))
	for _, binding := range bindings {
		if !validStableID(binding.ID) || !validStableID(binding.FromFieldID) || !validStableID(binding.ToFieldID) || binding.FromFieldID == binding.ToFieldID {
			return nil, "UNKNOWN_BINDING"
		}
		if !validKind(binding.Kind) || !validToken(binding.ExpectedStage, "stage:") || !validToken(binding.ExpectedReason, "reason:") {
			return nil, "UNKNOWN_BINDING"
		}
		if !precedence[evidenceKey(binding.ExpectedStage, binding.ExpectedReason)] || seen[binding.ID].ID != "" {
			return nil, "UNKNOWN_BINDING"
		}
		seen[binding.ID] = binding
	}
	return seen, ""
}

func validatePartitions(partitions []Partition, bindings map[string]Binding, precedence map[string]bool) string {
	seen := make(map[string]bool, len(partitions))
	for _, partition := range partitions {
		binding, ok := bindings[partition.BindingID]
		if !ok {
			return "DANGLING_PARTITION"
		}
		if partition.Polarity != PolarityMatch && partition.Polarity != PolarityMismatch {
			return "UNKNOWN_PARTITION"
		}
		if !validToken(partition.Stage, "stage:") || !validToken(partition.Reason, "reason:") || !precedence[evidenceKey(partition.Stage, partition.Reason)] {
			return "UNKNOWN_PARTITION"
		}
		if partition.Stage != binding.ExpectedStage || partition.Reason != binding.ExpectedReason {
			return "STALE_PARTITION"
		}
		key := partition.BindingID + "\x00" + partition.Polarity
		if seen[key] {
			return "DUPLICATE_PARTITION_POLARITY"
		}
		seen[key] = true
	}
	return ""
}

func missingEvidence(input Input) ([]string, []string) {
	seen := make(map[string]map[string]bool, len(input.RequiredBindings))
	for _, binding := range input.RequiredBindings {
		seen[binding.ID] = map[string]bool{PolarityMatch: false, PolarityMismatch: false}
	}
	for _, partition := range input.Partitions {
		if seen[partition.BindingID] != nil {
			seen[partition.BindingID][partition.Polarity] = true
		}
	}
	missingMatch, missingMismatch := make([]string, 0), make([]string, 0)
	for id, polarities := range seen {
		if !polarities[PolarityMatch] {
			missingMatch = append(missingMatch, id)
		}
		if !polarities[PolarityMismatch] {
			missingMismatch = append(missingMismatch, id)
		}
	}
	sort.Strings(missingMatch)
	sort.Strings(missingMismatch)
	return missingMatch, missingMismatch
}

func incompleteReason(match, mismatch []string) string {
	if len(match) != 0 && len(mismatch) != 0 {
		return "MISSING_MATCH_AND_MISMATCH"
	}
	if len(match) != 0 {
		return "MISSING_MATCH"
	}
	return "MISSING_MISMATCH"
}

func finish(vector Vector, decision, reason string) Result {
	vector.Decision = decision
	vector.Reason = reason
	vector.MissingMatch = append([]string{}, vector.MissingMatch...)
	vector.MissingMismatch = append([]string{}, vector.MissingMismatch...)
	return Result{Vector: vector, CanonicalDigest: digestVector(vector)}
}

func endpointReferenceCount(bindings []Binding) (int64, bool) {
	return safeAdd(int64(len(bindings)), int64(len(bindings)))
}

func validKind(kind string) bool {
	return kind == KindExactValue || kind == KindExactDigest || kind == KindSetEqual || kind == KindDerivedDigest
}

func validStableID(value string) bool {
	if !strings.HasPrefix(value, "sid:") || len(value) < 7 {
		return false
	}
	body := value[4:]
	if body[0] == '-' || body[len(body)-1] == '-' || strings.Contains(body, "--") {
		return false
	}
	for _, char := range body {
		if !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9') && char != '-' {
			return false
		}
	}
	return true
}

func validToken(value, prefix string) bool {
	return strings.HasPrefix(value, prefix) && validStableID("sid:"+value[len(prefix):])
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}

func evidenceKey(stage, reason string) string { return stage + "\x00" + reason }

func safeAdd(left, right int64) (int64, bool) {
	const maxInt64 = int64(1<<63 - 1)
	if right > 0 && left > maxInt64-right {
		return 0, true
	}
	return left + right, false
}
