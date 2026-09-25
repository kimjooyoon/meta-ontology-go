package lanefrontier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
)

func productionSHA(value string) string {
	if value == "" {
		return ""
	}
	if len(value) == 40 || len(value) == 64 {
		if _, err := hex.DecodeString(value); err == nil {
			return value
		}
	}
	sum := sha256.Sum256([]byte("sha:" + value))
	return hex.EncodeToString(sum[:20])
}
func productionLaneID(value string) string {
	if value == "" {
		return ""
	}
	return "urn:lane-frontier:translated/" + productionDigest(value)[:16]
}
func normalizedSchema(schema string) string {
	if schema == SchemaV1 {
		return production.SchemaVersion
	}
	return schema
}
func validatePartition(vector semanticVector) error {
	unknownReasons := map[string]bool{string(UnknownSchema): true, string(MissingInput): true, string(InvalidCount): true, string(AmbiguousOwner): true}
	if vector.Decision == string(Unknown) && !unknownReasons[vector.Reason] {
		return fmt.Errorf("UNKNOWN reason %q is outside the four allowed classes", vector.Reason)
	}
	if vector.Decision == string(Eligible) && vector.Reason != string(production.ReasonEligible) {
		return fmt.Errorf("ELIGIBLE reason %q is not ELIGIBLE", vector.Reason)
	}
	if vector.Decision != string(Unknown) && vector.Decision != string(Ineligible) && vector.Decision != string(Eligible) {
		return fmt.Errorf("unknown decision %q", vector.Decision)
	}
	return nil
}
func permutedCase(fixture Case) Case {
	permuted := fixture
	permuted.Input.OwnedPathPrefixes = reverseCopy(fixture.Input.OwnedPathPrefixes)
	permuted.Input.ChangedPaths = reverseCopy(fixture.Input.ChangedPaths)
	return permuted
}
func reverseCopy(values []string) []string {
	result := append([]string(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
func vectorJSON(vector semanticVector) string {
	data, _ := json.Marshal(vector)
	return string(data)
}
func pairedReceiptDigest(receipts []pairedReceipt) string {
	data, _ := json.Marshal(struct {
		Cases []pairedReceipt `json:"cases"`
	}{receipts})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
