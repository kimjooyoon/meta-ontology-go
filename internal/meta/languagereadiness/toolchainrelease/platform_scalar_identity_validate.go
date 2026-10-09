package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const scalarDescribeID = "urn:gooo:scalar-identity:describe"

func validateScalarIdentitySmoke(raw []byte, names [3]string, model, replay bool, selected string) (string, error) {
	generated, err := validateLanguageSmoke(raw, 4, replay, selected)
	if err != nil {
		return "", err
	}
	var result struct {
		Composition bodyexecution.Composition        `json:"composition"`
		Runtime     bodyexecution.CompositionRuntime `json:"runtime"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	c := result.Composition
	if c.Plan.EntryActivity != "Describe" || len(c.Plan.Activities) != 1 || len(c.Steps) != 1 {
		return "", fmt.Errorf("scalar identity declaration or assembly count differs")
	}
	a := c.Plan.Activities[0]
	if a.Name != "Describe" || a.ID != scalarDescribeID || len(a.Inputs) != 3 {
		return "", fmt.Errorf("scalar identity activity or input count differs")
	}
	for i, port := range a.Inputs {
		if port.Port != fmt.Sprintf("input%d", i) || port.Type != names[i] || port.EntityID != scalarInputIDs[i] || port.From != -1 {
			return "", fmt.Errorf("scalar identity authored name or native type differs")
		}
	}
	for _, boundary := range []string{"start", "end"} {
		if strings.Count(c.Source, fmt.Sprintf("//gooo:generated:%s id=%q kind=\"activity\"", boundary, scalarDescribeID)) != 1 {
			return "", fmt.Errorf("scalar identity generated marker differs")
		}
	}
	if err := validateScalarIdentityAssembly(c, model); err != nil {
		return "", err
	}
	return generated, validateScalarIdentityTraces(result.Runtime.Traces)
}

func validateScalarIdentityAssembly(c bodyexecution.Composition, model bool) error {
	r := c.Steps[0].Generation.Report.RecordAssembly
	if r == nil || r.Status != "COMPLETE_FINITE" || r.SelectedMask != 7 || r.Passed != 4 || r.Total != 4 ||
		r.FieldsPassed != 12 || r.FieldsTotal != 12 || len(r.Choices) != 3 || r.ModelRequested != model {
		return fmt.Errorf("scalar identity finite assembly differs")
	}
	if !model {
		if r.ModelCalls != 0 || len(r.Attempts) != 8 {
			return fmt.Errorf("scalar identity fixed order differs")
		}
		return nil
	}
	if r.ModelCalls != 1 || r.Prediction == nil || len(r.Ranking) != 8 || len(r.Attempts) != 6 ||
		r.Ranking[0] != r.Attempts[0].Mask || r.Attempts[0].FieldsPassed != 7 || r.Attempts[0].FieldsTotal != 12 ||
		r.Context == nil || r.Context.Status != "ENCODED" || r.Model == nil || !r.Model.Loaded ||
		r.Model.MetadataSHA256 != "3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202" ||
		r.Model.WeightsSHA256 != "76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f" {
		return fmt.Errorf("scalar identity own-model inference or retained identity differs")
	}
	return nil
}
