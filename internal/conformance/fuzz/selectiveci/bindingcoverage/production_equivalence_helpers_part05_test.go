package bindingcoverage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/bindingcoverage"
	"sort"
	"strings"
)

func authorizationFieldsAbsent(output production.Output) (bool, error) {
	data, err := json.Marshal(output)
	if err != nil {
		return false, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return false, err
	}
	for _, field := range []string{"execution_authorized", "ci_authorized"} {
		if _, exists := fields[field]; exists {
			return false, fmt.Errorf("production output contains %q", field)
		}
	}
	return true, nil
}
func mapProductionIDs(values []string) ([]string, error) {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !strings.HasPrefix(value, "urn:bindingcoverage:") {
			return nil, fmt.Errorf("unexpected production ID %q", value)
		}
		result = append(result, "sid:"+strings.TrimPrefix(value, "urn:bindingcoverage:"))
	}
	sort.Strings(result)
	return result, nil
}
func receiptDigest(receipts []productionReceipt) (string, error) {
	ordered := append([]productionReceipt{}, receipts...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].CaseName < ordered[j].CaseName })
	data, err := json.Marshal(ordered)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
