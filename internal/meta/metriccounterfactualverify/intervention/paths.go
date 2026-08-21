package metricintervention

import (
	"fmt"
	"path"
	"strings"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactual"
)

func manifestIndex(manifest metric.Manifest) (map[string]string, map[string]bool, error) {
	languages := make(map[string]string, len(manifest.Files))
	directories := map[string]bool{".": true}
	for _, file := range manifest.Files {
		if !validMetricPath(file.Path) || languages[file.Path] != "" {
			return nil, nil, fmt.Errorf("manifest path %q is invalid or duplicated", file.Path)
		}
		languages[file.Path] = file.Language
		for _, directory := range directoryChain(file.Path) {
			directories[directory] = true
		}
	}
	return languages, directories, nil
}

func directoryChain(file string) []string {
	directory := path.Dir(file)
	chain := []string{"."}
	if directory == "." {
		return chain
	}
	current := ""
	for _, part := range strings.Split(directory, "/") {
		current = path.Join(current, part)
		chain = append(chain, current)
	}
	return chain
}

func validMetricPath(value string) bool {
	return value != "" && value != "." && path.Clean(value) == value && !strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "../")
}

func languageOf(file string) string {
	switch path.Ext(file) {
	case ".go":
		return "go"
	case ".gooo":
		return "gooo"
	default:
		return "other"
	}
}

func addLanguage(delta *metric.Delta, language string, files, lines int) {
	if language == "go" {
		delta.GoFiles, delta.GoLines = delta.GoFiles+files, delta.GoLines+lines
	}
	if language == "gooo" {
		delta.GoooFiles, delta.GoooLines = delta.GoooFiles+files, delta.GoooLines+lines
	}
}
