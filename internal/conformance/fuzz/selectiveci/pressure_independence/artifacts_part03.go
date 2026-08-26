package pressureindependence

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"sort"
)

func canonicalRegistryMatchesArtifact(registry registryArtifact) bool {
	data, err := canonicalRegistryBytes(registry.Records, registry)
	return err == nil && string(data) == string(pressureRegistryArtifactBytes)
}
func registryBindingDigest(records []PressureRecord, contract registryArtifact) string {
	data, _ := canonicalRegistryBytes(records, contract)
	return digestBytes(data)
}
func canonicalRegistryBytes(records []PressureRecord, contract registryArtifact) ([]byte, error) {
	records = append([]PressureRecord(nil), records...)
	sort.Slice(records, func(left, right int) bool { return pressureKey(records[left]) < pressureKey(records[right]) })
	data, err := json.Marshal(registryArtifact{
		Schema: contract.Schema, FixtureID: contract.FixtureID, InputSchema: contract.InputSchema, Records: records,
	})
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
func artifactDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
