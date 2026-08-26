package actionability

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metabinding"
)

type documentDecoder struct{}

func (documentDecoder) Read[T any](path string) (T, []byte, error) {
	var document T
	data, err := os.ReadFile(path)
	if err != nil {
		return document, nil, err
	}
	if err := json.Unmarshal(data, &document); err != nil {
		return document, nil, err
	}
	return document, data, nil
}

func Load(metricsPath, bindingPath string) (input, error) {
	decoder := documentDecoder{}
	metrics, metricsData, err := decoder.Read[metricsDocument](metricsPath)
	if err != nil {
		return input{}, err
	}
	binding, bindingData, err := decoder.Read[metabinding.Report](bindingPath)
	if err != nil {
		return input{}, err
	}
	total := binding.Summary.BoundIndicators + binding.Summary.UnboundIndicators
	if len(metrics.CommitSHA) != 40 || metrics.Repository == "" || metrics.Meta.Schema == "" {
		return input{}, fmt.Errorf("source metrics identity is malformed")
	}
	if metrics.CommitSHA != binding.CommitSHA || metrics.Repository != binding.Repository {
		return input{}, fmt.Errorf("metrics and binding identities differ")
	}
	if binding.Schema != metabinding.Schema || total != len(metrics.Meta.Indicators) {
		return input{}, fmt.Errorf("binding coverage does not match indicator inventory")
	}
	return input{metrics: metrics, binding: binding,
		metricsDigest: digestBytes(metricsData), bindingDigest: digestBytes(bindingData)}, nil
}

func readAuthority(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(filepath.Clean(root), AuthorityPath))
	if err != nil {
		return "", err
	}
	required := []string{"package metaactionability", "activity PreserveProjectRootExemption(",
		"activity PreserveMetaBindingGuardrail(", "activity RegisterExecutableMetaOperation(",
		"activity ResolveIndicatorExecutor(", "activity SelectMissingExecutor(",
		"activity ReplayActionabilityReport("}
	for _, token := range required {
		if !strings.Contains(string(data), token) {
			return "", fmt.Errorf("actionability authority is missing %q", token)
		}
	}
	return digestBytes(data), nil
}
