package operationconformance

import (
	"fmt"
	"sort"
	"strconv"
)

func observeImports(evidence SplitGoEvidence) Decision {
	before, err := importSignature(evidence.Source)
	if err != nil || len(evidence.Candidates) == 0 {
		return DecisionFail
	}
	after := make([]string, 0)
	for _, candidate := range evidence.Candidates {
		items, importErr := importSignature(candidate)
		if importErr != nil {
			return DecisionFail
		}
		after = append(after, items...)
	}
	sort.Strings(after)
	if !sameStrings(before, after) {
		return DecisionFail
	}
	return DecisionPass
}

func importSignature(file FileEvidence) ([]string, error) {
	_, parsed, err := parseEvidence(file)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		path, unquoteErr := strconv.Unquote(spec.Path.Value)
		if unquoteErr != nil {
			return nil, unquoteErr
		}
		alias := ""
		if spec.Name != nil {
			alias = spec.Name.Name
		}
		result = append(result, fmt.Sprintf("%s\x00%s", alias, path))
	}
	sort.Strings(result)
	return result, nil
}
