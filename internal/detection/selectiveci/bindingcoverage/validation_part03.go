package bindingcoverage

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"math"
)

func populateShapeCounts(output *Output, required, partitions int) bool {
	output.RequiredBindingCount = uint64(required)
	output.PartitionCount = uint64(partitions)
	endpointReferences, ok := addUint64(output.RequiredBindingCount, output.RequiredBindingCount)
	if !ok {
		return false
	}
	output.EndpointReferenceCount = endpointReferences
	work, ok := workUnits(output.RequiredBindingCount, output.PartitionCount, endpointReferences)
	if !ok {
		return false
	}
	output.DeterministicWorkUnits = work
	return true
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
