package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const packageCallerKey = "app/retry:Main"
const packageCallerID = "gooo-workspace://activity/gooo-package1activity-main"
const packageBudgetName = "GoooPackage0ActivityPlanBudget"
const packageBudgetID = "gooo-workspace://activity/gooo-package0activity-plan-budget"

type packageConstructionEnvelope struct {
	Schema, Decision, Error string
	Manifest                string `json:"manifest_digest"`
	Cases                   string `json:"cases_digest"`
	From                    string `json:"replayed_from_sha256"`
	Result                  json.RawMessage
}
type packageConstructionResult struct {
	Schema                   string
	Program                  json.RawMessage
	Cases                    json.RawMessage `json:"construction_cases"`
	Construction, Evaluation json.RawMessage
	From                     string `json:"replayed_from_sha256"`
}

func validatePackageConstructionSmoke(raw []byte, reference packageConstructionReference, budget int, saved []byte) error {
	var envelope packageConstructionEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return err
	}
	var result packageConstructionResult
	if err := json.Unmarshal(envelope.Result, &result); err != nil {
		return err
	}
	decision := "COMPLETE_FINITE"
	if budget == 5 {
		decision = "PARTIAL_FINITE"
	}
	if envelope.Schema != "gooo/workspace-caller-construction-receipt/v1" || envelope.Decision != decision || envelope.Error != "" ||
		envelope.Manifest != packageSmokeSHA(reference.Manifest) || envelope.Cases != packageSmokeSHA(reference.Evaluation) ||
		result.Schema != "gooo/workspace-caller-construction/v1" || !samePackageSourceValue(result.Program, reference.Program) ||
		!samePackageSourceValue(result.Cases, reference.Feedback) {
		return fmt.Errorf("package construction envelope or original package binding differs")
	}
	selected, err := validatePackageConstructionReplay(envelope, result, saved)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(map[string]any{"generated_now": len(saved) == 0, "construction": result.Construction, "evaluation": result.Evaluation})
	if err != nil {
		return err
	}
	var joint jointSmokeOutput
	if err = json.Unmarshal(encoded, &joint); err != nil {
		return err
	}
	if err = validateNativeArithmeticIdentity(joint, []byte(reference.Source), budget, len(saved) > 0, selected); err != nil {
		return err
	}
	if err = validateJointFillInitialIDs(joint, []string{"late_unbounded", "early_wrong_cap", "caller_zero_divisor", "bounded"}); err != nil {
		return err
	}
	feedback, lowered, err := packageConstructionCases(reference.Feedback, 1)
	if err != nil {
		return err
	}
	evaluation, _, err := packageConstructionCases(reference.Evaluation, 4)
	if err != nil {
		return err
	}
	var native struct {
		Cases    json.RawMessage `json:"construction_cases"`
		Attempts []struct{ Runtime nativeSmokeRuntime }
	}
	if err = json.Unmarshal(result.Construction, &native); err != nil {
		return err
	}
	if !samePackageSourceValue(native.Cases, lowered) {
		return fmt.Errorf("lowered feedback differs from original package cases")
	}
	if err = validatePackageConstructionAttempts(joint, native.Attempts[4].Runtime, feedback); err != nil {
		return err
	}
	c, e := joint.Construction, joint.Evaluation
	index := 5
	var actual []json.RawMessage
	if budget == 5 {
		index = 0
		actual = []json.RawMessage{[]byte("2"), []byte("-12"), []byte("5"), []byte("9007199254740994")}
	}
	if c.Attempts[index].Runtime.SHA != c.Selected.SHA || e.Runtime.SHA != c.Selected.SHA ||
		c.Attempts[index].FillCandidates[0].SelectedSHA != packageSmokeSHA([]byte(c.Source)) {
		return fmt.Errorf("package selected source differs from execution")
	}
	if err = validateJointFillRuntimeFor(e.Runtime, evaluation, actual, packageCallerKey, packageCallerID); err != nil {
		return err
	}
	return validateNativeSmokeRuns(e.Runtime.Runs)
}

func validatePackageConstructionReplay(envelope packageConstructionEnvelope, result packageConstructionResult, saved []byte) (string, error) {
	if len(saved) == 0 {
		if envelope.From != "" || result.From != "" {
			return "", fmt.Errorf("fresh package construction claims saved replay")
		}
		return "", nil
	}
	var old packageConstructionEnvelope
	if err := json.Unmarshal(saved, &old); err != nil {
		return "", err
	}
	var prior packageConstructionResult
	if err := json.Unmarshal(old.Result, &prior); err != nil {
		return "", err
	}
	// The inner digest follows the compiler's typed saved-result encoding; the
	// outer digest binds the original receipt bytes, including historical timing.
	var typed workspaceexecution.ConstructionResult
	if err := json.Unmarshal(old.Result, &typed); err != nil {
		return "", err
	}
	canonical, err := json.Marshal(typed)
	if err != nil {
		return "", err
	}
	if envelope.From != packageSmokeSHA(saved) || result.From != packageSmokeSHA(canonical) ||
		!samePackageSourceValue(result.Construction, prior.Construction) {
		return "", fmt.Errorf("package replay changed original receipt or saved attempt history")
	}
	return typed.Construction.Selected.GeneratedSHA256, nil
}

func packageConstructionCases(raw []byte, count int) ([]jointSmokeCase, []byte, error) {
	var suite struct {
		Schema string           `json:"schema"`
		Cases  []jointSmokeCase `json:"cases"`
	}
	if err := json.Unmarshal(raw, &suite); err != nil {
		return nil, nil, err
	}
	if suite.Schema != "gooo/body-composition-cases/v1" || len(suite.Cases) != count {
		return nil, nil, fmt.Errorf("package reference cases differ")
	}
	original := suite.Cases
	rows := make([]map[string]any, len(original))
	for i, row := range original {
		if len(row.Inputs) != 1 || len(row.Expected) != 1 || len(row.Inputs[packageCallerKey]) == 0 || len(row.Expected[packageCallerKey]) == 0 {
			return nil, nil, fmt.Errorf("package reference needs the original named caller")
		}
		rows[i] = map[string]any{"inputs": map[string]json.RawMessage{"GoooPackage1ActivityMain": row.Inputs[packageCallerKey]}, "expected": map[string]json.RawMessage{"GoooPackage1ActivityMain": row.Expected[packageCallerKey]}}
	}
	lowered, err := json.Marshal(map[string]any{"schema": suite.Schema, "cases": rows})
	return original, lowered, err
}

func validatePackageConstructionAttempts(joint jointSmokeOutput, fault nativeSmokeRuntime, cases []jointSmokeCase) error {
	c := joint.Construction
	plan := c.Initial.Preparations[0].Generation.Report.Fill.PlanSHA
	for i, a := range c.Attempts {
		if len(a.FillCandidates) != 1 || a.FillCandidates[0].PlanSHA != plan {
			return fmt.Errorf("package candidate plan differs")
		}
		if i == 1 || i == 2 {
			if err := validateJointRejectedFillFor(a, i, 6, c.OriginalSHA, packageBudgetName, packageBudgetID); err != nil {
				return err
			}
			continue
		}
		local := map[int]int{0: 0, 3: 1, 4: 3, 5: 2}[i]
		if err := validateJointFillLocalFor(a, local, i, 6, c.OriginalSHA, packageBudgetName, packageBudgetID); err != nil {
			return err
		}
		if i == 4 {
			if fault.Source != a.FillCandidates[0].SelectedSHA {
				return fmt.Errorf("package fault source differs")
			}
			if err := validateNativeSmokeRuntime(fault, [5]int{0, 0, 1, 0, 0}, 1); err != nil {
				return err
			}
			if len(fault.Traces[0].Deliveries) != 1 {
				return fmt.Errorf("package caller fault lost its delivery")
			}
			d := fault.Traces[0].Deliveries[0]
			if d.ID != packageCallerID || !samePackageSourceValue(d.Input, cases[0].Inputs[packageCallerKey]) ||
				!samePackageSourceValue(d.Expected, cases[0].Expected[packageCallerKey]) || !jointSmokeBool(d.Passed, false) {
				return fmt.Errorf("package caller fault changed original values")
			}
			if err := validateNativeSmokeFault(fault, d, packageBudgetID, 8); err != nil {
				return err
			}
			continue
		}
		actual := []json.RawMessage{json.RawMessage([]string{"9", "-7", "-8"}[local])}
		if err := validateJointFillRuntimeFor(a.Runtime, cases, actual, packageCallerKey, packageCallerID); err != nil {
			return err
		}
		if err := validateNativeSmokeRuns(a.Runtime.Runs); err != nil {
			return err
		}
	}
	return nil
}

func packageSmokeSHA(raw []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) }
