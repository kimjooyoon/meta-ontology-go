package ambiguitybudget

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func decodeEffects(raw []byte) (Effects, error) {
	var artifact Effects
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&artifact); err != nil {
		return Effects{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Effects{}, fmt.Errorf("workspace effects have trailing JSON")
		}
		return Effects{}, err
	}
	if artifact.Schema != EffectsSchema || artifact.Version != "v1" || !artifact.TrackedAndUntracked ||
		!validDigest(artifact.SnapshotBeforeDigest) || !validDigest(artifact.SnapshotAfterDigest) ||
		artifact.RepositoryWrites != 0 || !artifact.WriteSetEqual ||
		artifact.MutationAuthority != "UNKNOWN" || artifact.MutationAuthorityResolution != "NOT_OBSERVED" {
		return Effects{}, fmt.Errorf("workspace effects are not an observed zero-write artifact")
	}
	artifact.ArtifactDigest = digestBytes(raw)
	return artifact, nil
}
