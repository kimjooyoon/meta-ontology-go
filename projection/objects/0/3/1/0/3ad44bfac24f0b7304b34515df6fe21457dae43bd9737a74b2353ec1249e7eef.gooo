package semanticdelta

import (
	"fmt"
	"io"
)

// Evaluate decodes a request, writes a deterministic report, and returns a
// ScopeError when the semantic delta is outside the allowed scope.
func Evaluate(input io.Reader, output io.Writer, outputFormat Format) (Report, error) {
	if input == nil || output == nil {
		return Report{}, fmt.Errorf("semanticdelta input and output are required")
	}
	data, err := io.ReadAll(input)
	if err != nil {
		return Report{}, fmt.Errorf("read semanticdelta input: %w", err)
	}
	request, err := Decode(data)
	if err != nil {
		return Report{}, err
	}
	report, err := Detect(request.Delta, request.Allowed)
	if err != nil {
		return Report{}, err
	}
	var encoded []byte
	switch outputFormat {
	case FormatText:
		encoded = EncodeReportText(report)
	case FormatJSON, "":
		encoded, err = EncodeReportJSON(report)
		if err != nil {
			return Report{}, err
		}
	default:
		return Report{}, fmt.Errorf("unsupported semanticdelta report format %q", outputFormat)
	}
	if _, err := output.Write(encoded); err != nil {
		return Report{}, fmt.Errorf("write semanticdelta report: %w", err)
	}
	if !report.Passes() {
		return report, &ScopeError{Report: report}
	}
	return report, nil
}
