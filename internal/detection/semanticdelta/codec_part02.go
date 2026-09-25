package semanticdelta

import (
	"encoding/json"
	"fmt"
)

// EncodeReportJSON returns canonical JSON for a detection report.
func EncodeReportJSON(report Report) ([]byte, error) {
	report.Normalize()
	encoded, err := json.MarshalIndent(jsonReport(report), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode semanticdelta report: %w", err)
	}
	return append(encoded, '\n'), nil
}
