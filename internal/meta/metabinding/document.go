package metabinding

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func Load(path string) (input, error) {
	var in input
	data, err := os.ReadFile(path)
	if err != nil {
		return in, err
	}
	if err := json.Unmarshal(data, &in.document); err != nil {
		return in, err
	}
	if err := json.Unmarshal(data, &in.raw); err != nil {
		return in, err
	}
	if in.document.Meta.Schema == "" || len(in.document.Meta.Indicators) == 0 {
		return in, fmt.Errorf("source metrics indicator report is empty")
	}
	if len(in.document.CommitSHA) != 40 {
		return in, fmt.Errorf("source metrics commit SHA is malformed")
	}
	for _, indicator := range in.document.Meta.Indicators {
		if string(indicator.MetricID) == MetricID {
			return in, fmt.Errorf("source metrics already contain the recursive indicator")
		}
	}
	in.sourceDigest = digestBytes(data)
	return in, nil
}

func Augment(in input, indicator sourcepolicy.Indicator) ([]byte, error) {
	meta := in.document.Meta
	meta.Indicators = append(append([]sourcepolicy.Indicator(nil), meta.Indicators...), indicator)
	encodedMeta, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	raw := make(map[string]json.RawMessage, len(in.raw))
	maps.Copy(raw, in.raw)
	raw["meta"] = encodedMeta
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func MarshalReport(report Report) ([]byte, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
