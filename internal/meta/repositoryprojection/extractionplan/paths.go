package extractionplan

import (
	"bytes"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/moduleboundary"
	projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
	"path/filepath"
	"slices"
	"strings"
)

func extractionPath(root, logical string) (string, error) {
	if logical == "" || filepath.IsAbs(logical) || strings.ContainsRune(logical, 0) {
		return "", fmt.Errorf("unsafe extraction path %q", logical)
	}
	clean := filepath.Clean(filepath.FromSlash(logical))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe extraction path %q", logical)
	}
	return filepath.Join(root, clean), nil
}

func validateRecipeModule(root, logical string, recipe extractionRecipe) error {
	source, err := extractionPath(root, logical)
	if err != nil {
		return err
	}
	targets := make([]string, 0, len(recipe.Edits)+len(recipe.Creates))
	for _, edit := range recipe.Edits {
		target, err := extractionPath(root, edit.Path)
		if err != nil {
			return err
		}
		targets = append(targets, target)
	}
	for _, creation := range recipe.Creates {
		target, err := extractionPath(root, creation.Path)
		if err != nil {
			return err
		}
		targets = append(targets, target)
	}
	return moduleboundary.CheckWriteSet(root, source, targets...)
}

func extractionLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func editText(lines []string) []byte {
	return []byte(strings.Join(lines, "\n"))
}

func appendUnique(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}
	return append(values, value)
}

func stageGenericExtraction(root, logical string, buffers map[string][]byte, created map[string]bool, changed, made map[string][]string) ([]string, []projectionextractor.StrategyEvidence, error) {
	result, err := projectionextractor.ExtractWithResult(root, logical)
	if err != nil {
		return nil, nil, err
	}
	if len(result.Operations) == 0 {
		return nil, nil, fmt.Errorf("generic extraction performed no operation")
	}
	for path, data := range result.Generated {
		buffers[path] = data
		changed[logical] = appendUnique(changed[logical], path)
		if path != logical {
			created[path] = true
			made[logical] = appendUnique(made[logical], path)
		}
	}
	return append([]string{}, result.Operations...), append([]projectionextractor.StrategyEvidence{}, result.Evidence...), nil
}
