package bindingcoverage

import (
	"math"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateHeader(input Input) Reason {
	if input.ContractID == "" || input.SnapshotDigest == "" || input.ExpectedSnapshotDigest == "" {
		return ReasonMissingInput
	}
	if !validStableID(input.ContractID) {
		return ReasonInvalidID
	}
	if !validDigest(input.SnapshotDigest) || !validDigest(input.ExpectedSnapshotDigest) {
		return ReasonInvalidDigest
	}
	if input.SnapshotDigest != input.ExpectedSnapshotDigest {
		return ReasonSnapshotMismatch
	}
	return ""
}

func validatePrecedence(entries []PrecedenceEntry) (map[string]struct{}, Reason) {
	ranks := make(map[uint64]struct{}, len(entries))
	pairs := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if validateStageToken(entry.Stage) != "" || validateReasonToken(entry.Reason) != "" {
			return nil, ReasonInvalidPrecedence
		}
		if _, exists := ranks[entry.Rank]; exists {
			return nil, ReasonDuplicatePrecedence
		}
		pair := expectedPair(entry.Stage, entry.Reason)
		if _, exists := pairs[pair]; exists {
			return nil, ReasonDuplicatePrecedence
		}
		ranks[entry.Rank] = struct{}{}
		pairs[pair] = struct{}{}
	}
	return pairs, ""
}

func validateBindings(bindings []RequiredBinding, precedencePairs map[string]struct{}) (map[string]string, uint64, Reason) {
	ids := make(map[string]struct{}, len(bindings))
	bindingPairs := make(map[string]string, len(bindings))
	endpointReferences := uint64(0)
	for _, binding := range bindings {
		if reason := validateID(binding.BindingID); reason != "" {
			return nil, 0, reason
		}
		if _, exists := ids[binding.BindingID]; exists {
			return nil, 0, ReasonDuplicateID
		}
		if reason := validateID(binding.FromFieldID); reason != "" {
			return nil, 0, reason
		}
		if reason := validateID(binding.ToFieldID); reason != "" {
			return nil, 0, reason
		}
		if !validKind(binding.Kind) {
			return nil, 0, ReasonInvalidEnum
		}
		if reason := validateStageToken(binding.ExpectedStage); reason != "" {
			return nil, 0, reason
		}
		if reason := validateReasonToken(binding.ExpectedReason); reason != "" {
			return nil, 0, reason
		}
		if binding.FromFieldID == binding.ToFieldID {
			return nil, 0, ReasonSelfLink
		}
		pair := expectedPair(binding.ExpectedStage, binding.ExpectedReason)
		if _, registered := precedencePairs[pair]; !registered {
			return nil, 0, ReasonUnregisteredPair
		}
		var ok bool
		endpointReferences, ok = addUint64(endpointReferences, 2)
		if !ok {
			return nil, 0, ReasonWorkOverflow
		}
		ids[binding.BindingID] = struct{}{}
		bindingPairs[binding.BindingID] = pair
	}
	return bindingPairs, endpointReferences, ""
}

func validatePartitions(partitions []Partition, bindingPairs map[string]string) (map[string]struct{}, map[string]struct{}, Reason) {
	ids := make(map[string]struct{}, len(partitions))
	match := make(map[string]struct{}, len(bindingPairs))
	mismatch := make(map[string]struct{}, len(bindingPairs))
	polarity := make(map[string]struct{}, len(partitions))
	for _, partition := range partitions {
		if reason := validateID(partition.PartitionID); reason != "" {
			return nil, nil, reason
		}
		if _, exists := bindingPairs[partition.PartitionID]; exists {
			return nil, nil, ReasonDuplicateID
		}
		if _, exists := ids[partition.PartitionID]; exists {
			return nil, nil, ReasonDuplicateID
		}
		if reason := validateID(partition.BindingID); reason != "" {
			return nil, nil, reason
		}
		expected, exists := bindingPairs[partition.BindingID]
		if !exists {
			return nil, nil, ReasonUnknownReference
		}
		if !validPolarity(partition.Polarity) {
			return nil, nil, ReasonInvalidEnum
		}
		if reason := validateStageToken(partition.ExpectedStage); reason != "" {
			return nil, nil, reason
		}
		if reason := validateReasonToken(partition.ExpectedReason); reason != "" {
			return nil, nil, reason
		}
		if expected != expectedPair(partition.ExpectedStage, partition.ExpectedReason) {
			return nil, nil, ReasonStalePartition
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

func workUnits(required, partitions, endpointReferences uint64) (uint64, bool) {
	total, ok := addUint64(required, partitions)
	if !ok {
		return 0, false
	}
	return addUint64(total, endpointReferences)
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

func validateStageToken(value string) Reason {
	return validateWireToken(value, "stage:")
}

func validateReasonToken(value string) Reason {
	return validateWireToken(value, "reason:")
}

func validateWireToken(value, prefix string) Reason {
	if value == "" {
		return ReasonMissingInput
	}
	if !strings.HasPrefix(value, prefix) {
		return ReasonInvalidToken
	}
	token := value[len(prefix):]
	if token == "" || token[0] == '-' || token[len(token)-1] == '-' {
		return ReasonInvalidToken
	}
	for index := 0; index < len(token); index++ {
		char := token[index]
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
			return ReasonInvalidToken
		}
		if index > 0 && token[index-1] == '-' && char == '-' {
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

func expectedPair(stage, reason string) string { return stage + "\x00" + reason }
