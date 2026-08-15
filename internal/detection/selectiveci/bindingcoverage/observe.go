package bindingcoverage

import (
	"math"
	"strings"
	"unicode"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// Observe evaluates explicit binding partitions without side effects or
// inferred coverage. Classify is a descriptive alias for the same observer.
func Observe(input Input) Output {
	input = normalizeInput(input)
	canonical, err := input.CanonicalJSON()
	if err != nil {
		return seal(baseOutput(input, 0), DecisionUnknown, ReasonEvaluatorError)
	}
	result := baseOutput(input, uint64(len(canonical)))
	if input.SchemaVersion != SchemaVersion {
		return seal(result, DecisionUnknown, ReasonUnknownSchema)
	}
	if input.RequiredBindings == nil || input.Partitions == nil {
		return seal(result, DecisionUnknown, ReasonMissingInput)
	}
	if reason := validateHeader(input); reason != "" {
		return seal(result, DecisionUnknown, reason)
	}

	bindingIDs, endpoints, reason := validateBindings(input.RequiredBindings)
	if reason != "" {
		return seal(result, DecisionUnknown, reason)
	}
	match, mismatch, reason := validatePartitions(input.Partitions, bindingIDs)
	if reason != "" {
		return seal(result, DecisionUnknown, reason)
	}
	result.RequiredBindingCount = uint64(len(input.RequiredBindings))
	result.PartitionCount = uint64(len(input.Partitions))
	result.MissingMatchBindingIDs, result.MissingMismatchBindingIDs = missingBindings(input.RequiredBindings, match, mismatch)
	result.MatchCoveredCount = result.RequiredBindingCount - uint64(len(result.MissingMatchBindingIDs))
	result.MismatchCoveredCount = result.RequiredBindingCount - uint64(len(result.MissingMismatchBindingIDs))
	work, ok := workUnits(result.RequiredBindingCount, result.PartitionCount, uint64(len(endpoints)))
	if !ok {
		return seal(baseOutput(input, uint64(len(canonical))), DecisionUnknown, ReasonWorkOverflow)
	}
	result.DeterministicWorkUnits = work
	if result.RequiredBindingCount == 0 || len(result.MissingMatchBindingIDs) != 0 || len(result.MissingMismatchBindingIDs) != 0 {
		return seal(result, DecisionIncomplete, ReasonIncomplete)
	}
	return seal(result, DecisionExact, ReasonExact)
}

func Classify(input Input) Output { return Observe(input) }

func Evaluate(input Input) Output { return Observe(input) }

func baseOutput(input Input, inputBytes uint64) Output {
	return Output{SchemaVersion: input.SchemaVersion, ContractID: input.ContractID, SnapshotDigest: input.SnapshotDigest, InputBytes: inputBytes}
}

func seal(output Output, decision Decision, reason Reason) Output {
	output = normalizeOutput(output)
	output.Decision = decision
	output.Reason = reason
	output.CanonicalDigest = output.StableDigest()
	return output
}

func validateHeader(input Input) Reason {
	if input.ContractID == "" || input.SnapshotDigest == "" {
		return ReasonMissingInput
	}
	if !validStableID(input.ContractID) {
		return ReasonInvalidID
	}
	if !validDigest(input.SnapshotDigest) {
		return ReasonInvalidDigest
	}
	return ""
}

func validateBindings(bindings []RequiredBinding) (map[string]struct{}, map[string]struct{}, Reason) {
	ids := make(map[string]struct{}, len(bindings))
	endpoints := make(map[string]struct{}, len(bindings)*2)
	for _, binding := range bindings {
		if reason := validateID(binding.BindingID); reason != "" {
			return nil, nil, reason
		}
		if _, exists := ids[binding.BindingID]; exists {
			return nil, nil, ReasonDuplicateID
		}
		if reason := validateID(binding.FromFieldID); reason != "" {
			return nil, nil, reason
		}
		if reason := validateID(binding.ToFieldID); reason != "" {
			return nil, nil, reason
		}
		if !validKind(binding.Kind) {
			return nil, nil, ReasonInvalidEnum
		}
		ids[binding.BindingID] = struct{}{}
		endpoints[binding.FromFieldID] = struct{}{}
		endpoints[binding.ToFieldID] = struct{}{}
	}
	return ids, endpoints, ""
}

func validatePartitions(partitions []Partition, bindingIDs map[string]struct{}) (map[string]struct{}, map[string]struct{}, Reason) {
	ids := make(map[string]struct{}, len(partitions))
	match := make(map[string]struct{}, len(bindingIDs))
	mismatch := make(map[string]struct{}, len(bindingIDs))
	polarity := make(map[string]struct{}, len(partitions))
	for _, partition := range partitions {
		if reason := validateID(partition.PartitionID); reason != "" {
			return nil, nil, reason
		}
		if _, exists := bindingIDs[partition.PartitionID]; exists {
			return nil, nil, ReasonDuplicateID
		}
		if _, exists := ids[partition.PartitionID]; exists {
			return nil, nil, ReasonDuplicateID
		}
		if reason := validateID(partition.BindingID); reason != "" {
			return nil, nil, reason
		}
		if _, exists := bindingIDs[partition.BindingID]; !exists {
			return nil, nil, ReasonUnknownReference
		}
		if !validPolarity(partition.Polarity) {
			return nil, nil, ReasonInvalidEnum
		}
		if reason := validateToken(partition.ExpectedStage); reason != "" {
			return nil, nil, reason
		}
		if reason := validateToken(partition.ExpectedReason); reason != "" {
			return nil, nil, reason
		}
		key := partition.BindingID + "\x00" + string(partition.Polarity)
		if _, exists := polarity[key]; exists {
			return nil, nil, ReasonDuplicatePolarity
		}
		ids[partition.PartitionID] = struct{}{}
		polarity[key] = struct{}{}
		if partition.Polarity == PolarityMatch {
			match[partition.BindingID] = struct{}{}
		} else {
			mismatch[partition.BindingID] = struct{}{}
		}
	}
	return match, mismatch, ""
}

func missingBindings(bindings []RequiredBinding, match, mismatch map[string]struct{}) ([]string, []string) {
	missingMatch := make([]string, 0)
	missingMismatch := make([]string, 0)
	for _, binding := range bindings {
		if _, exists := match[binding.BindingID]; !exists {
			missingMatch = append(missingMatch, binding.BindingID)
		}
		if _, exists := mismatch[binding.BindingID]; !exists {
			missingMismatch = append(missingMismatch, binding.BindingID)
		}
	}
	return missingMatch, missingMismatch
}

func workUnits(required, partitions, endpoints uint64) (uint64, bool) {
	total, ok := addUint64(required, partitions)
	if !ok {
		return 0, false
	}
	return addUint64(total, endpoints)
}

func addUint64(left, right uint64) (uint64, bool) {
	if math.MaxUint64-left < right {
		return 0, false
	}
	return left + right, true
}

func validateID(value string) Reason {
	if value == "" {
		return ReasonMissingInput
	}
	if !validStableID(value) {
		return ReasonInvalidID
	}
	return ""
}

func validStableID(value string) bool {
	parsed, err := semantic.ParseIdentity(value)
	return err == nil && parsed.String() == value
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, char := range value {
		if !(char >= '0' && char <= '9') && !(char >= 'a' && char <= 'f') {
			return false
		}
	}
	return true
}

func validateToken(value string) Reason {
	if value == "" {
		return ReasonMissingInput
	}
	if strings.TrimSpace(value) != value {
		return ReasonInvalidToken
	}
	for _, char := range value {
		if unicode.IsSpace(char) || unicode.IsControl(char) {
			return ReasonInvalidToken
		}
	}
	return ""
}

func validKind(kind BindingKind) bool {
	return kind == KindExactValue || kind == KindExactDigest || kind == KindSetEqual || kind == KindDerivedDigest
}

func validPolarity(polarity Polarity) bool {
	return polarity == PolarityMatch || polarity == PolarityMismatch
}
