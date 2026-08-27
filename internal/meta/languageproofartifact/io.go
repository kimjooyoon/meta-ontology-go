package languageproofartifact

import (
	"encoding/json"
	"os"
)

func Write(path string, artifact Artifact) error {
	raw, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
