package languagesemantic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	syntaxreplay "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

func validateUpstreamSyntaxArtifact(raw []byte, expectedHead, root string) (languagesyntax.Report, error) {
	var report languagesyntax.Report
	if err := json.Unmarshal(raw, &report); err != nil {
		return report, fmt.Errorf("decode upstream syntax evidence: %w", err)
	}
	if err := languagesyntax.Validate(report, expectedHead); err != nil {
		return report, fmt.Errorf("upstream syntax contract: %w", err)
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return report, fmt.Errorf("resolve upstream syntax source root: %w", err)
	}
	observed, err := syntaxreplay.Observe(os.DirFS(absoluteRoot))
	if err != nil {
		return report, fmt.Errorf("observe upstream syntax source: %w", err)
	}
	if err := compareSyntaxSources(report.Source.GoooFiles, observed); err != nil {
		return report, err
	}
	return report, nil
}

func compareSyntaxSources(receipt, observed []syntaxreplay.FileObservation) error {
	if len(receipt) != len(observed) {
		return fmt.Errorf("upstream syntax source inventory contains %d files, want %d", len(receipt), len(observed))
	}
	receiptByPath, err := syntaxSourceObservationsByPath(receipt)
	if err != nil {
		return fmt.Errorf("upstream syntax receipt source inventory: %w", err)
	}
	observedByPath, err := syntaxSourceObservationsByPath(observed)
	if err != nil {
		return fmt.Errorf("upstream syntax source inventory: %w", err)
	}
	if len(receiptByPath) != len(observedByPath) {
		return fmt.Errorf("upstream syntax source inventory has duplicate or missing paths")
	}
	for path, receiptFile := range receiptByPath {
		observedFile, ok := observedByPath[path]
		if !ok {
			return fmt.Errorf("upstream syntax receipt omits observed source %q", path)
		}
		if receiptFile.GoooLines != observedFile.GoooLines || receiptFile.SourceDigest != observedFile.SourceDigest {
			return fmt.Errorf("upstream syntax source binding mismatch for %q", path)
		}
	}
	return nil
}

func syntaxSourceObservationsByPath(observations []syntaxreplay.FileObservation) (map[string]syntaxreplay.FileObservation, error) {
	byPath := make(map[string]syntaxreplay.FileObservation, len(observations))
	for _, observation := range observations {
		path := filepath.ToSlash(filepath.Clean(observation.Path))
		if path == "." || observation.Path != path {
			return nil, fmt.Errorf("source path is not canonical: %q", observation.Path)
		}
		if _, exists := byPath[path]; exists {
			return nil, fmt.Errorf("source path is duplicated: %q", path)
		}
		byPath[path] = observation
	}
	return byPath, nil
}
