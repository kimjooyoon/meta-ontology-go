package toolchainrelease

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func validateRelativeV3Smoke(raw, source, feedback, cases []byte, preflight *bodycodegen.TypedPathContextExport,
	head string, replay bool, selected string) (string, error) {
	var r typedSmokeOutput
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", err
	}
	if err := validateRelativeV2Header(r, source, feedback, replay, selected); err != nil {
		return "", err
	}
	c, e := r.Construction, r.Evaluation
	p, err := validateRelativeV3Binding(c.Initial.Preparations[0].Generation.Report, preflight, head)
	if err != nil {
		return "", err
	}
	if err = validateRelativeV2Candidates(p, source); err != nil {
		return "", err
	}
	if c.Attempts[0].Runtime.SHA != c.Selected.GeneratedSHA256 || e.Runtime.SHA != c.Selected.GeneratedSHA256 {
		return "", fmt.Errorf("V2 native program identity differs")
	}
	if err = validateRelativeV2Envelopes(raw, c, replay); err != nil {
		return "", err
	}
	if err = validateRelativeV3Runtime(e.Runtime, cases, 8); err != nil {
		return "", err
	}
	if err = validateRelativeV3Runtime(c.Attempts[0].Runtime, feedback, 1); err != nil {
		return "", err
	}
	return c.Selected.GeneratedSHA256, nil
}
