package selfimprovementloop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteArtifacts writes only to the caller-selected output directory.
func WriteArtifacts(outputDir string, artifacts Artifacts) error {
	if filepath.Clean(outputDir) == "." || outputDir == "" {
		return fmt.Errorf("caller-owned temporary output directory is required")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	files := map[string]any{
		"report.json": artifacts.Report, "patch-proposal.json": artifacts.PatchProposal,
		"evidence.json": artifacts.Evidence, "dossier.json": artifacts.Dossier,
	}
	for name, value := range files {
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal %s: %w", name, err)
		}
		data = append(data, '\n')
		if err := os.WriteFile(filepath.Join(outputDir, name), data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}
