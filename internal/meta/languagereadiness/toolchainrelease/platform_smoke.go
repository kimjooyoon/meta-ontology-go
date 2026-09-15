package toolchainrelease

import (
	"encoding/json"
	"fmt"
)

func smokeBinary(path, root string) (SmokeEvidence, error) {
	output, err := commandOutput(root, nil, path, "version", "--json")
	if err != nil {
		return SmokeEvidence{}, err
	}
	var smoke SmokeEvidence
	if err := json.Unmarshal(output, &smoke); err != nil {
		return SmokeEvidence{}, err
	}
	if smoke.SchemaVersion != "gooo-version/v1" ||
		smoke.Language != "gooo" ||
		smoke.Version == "" ||
		smoke.Status != "development" {
		return SmokeEvidence{}, fmt.Errorf("TOOLCHAIN_RELEASE_NATIVE_SMOKE_MISMATCH")
	}
	return smoke, nil
}
