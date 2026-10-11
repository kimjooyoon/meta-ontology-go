package toolchainrelease

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type checkpointDetails struct {
	Initial struct {
		Preparations []struct {
			InputSHA   string `json:"input_source_sha256"`
			Generation struct {
				Report struct {
					Activity string
					Search   *struct {
						Original   string                                              `json:"original_source_digest"`
						Plan       string                                              `json:"ir_plan_sha256"`
						Selected   string                                              `json:"selected_candidate_id"`
						Calls      *int                                                `json:"provider_operations"`
						Pool       *bodycodegen.RetainedSearchCandidateSet             `json:"initial_candidate_set"`
						Generation *bodycodegen.IRBodySearchCandidateGenerationReceipt `json:"candidate_generation"`
					} `json:"body_search"`
				}
			}
		}
	}
	Attempts []struct {
		Search []bodycodegen.SearchCandidate `json:"search_candidates"`
	}
}

func validateSearchCheckpoint(raw, source, feedback, evaluation []byte, replay bool, retained string) (string, error) {
	var r jointSmokeOutput
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", err
	}
	var envelope struct {
		Construction json.RawMessage
		Evaluation   struct {
			Runtime struct {
				Separation struct {
					Status   string
					Unknown  *int `json:"unknown_inputs"`
					Overlap  *int `json:"overlapping_inputs"`
					Disjoint *int `json:"disjoint_inputs"`
				} `json:"input_separation"`
			}
		}
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return "", err
	}
	constructionSHA, err := checkpointConstructionDigest(envelope.Construction)
	if err != nil || replay && (retained == "" || retained != constructionSHA) {
		return "", fmt.Errorf("search checkpoint replay changed the saved construction")
	}
	c, e := r.Construction, r.Evaluation
	if !jointSmokeBool(r.Generated, !replay) || c.Schema != "gooo/joint-construction/v2" ||
		c.Stage != "COMPLETE" || c.Decision != "COMPLETE_FINITE" || c.Failure != "" ||
		c.OriginalSHA != fmt.Sprintf("sha256:%x", sha256.Sum256(source)) || c.Source == "" || c.Selected.SHA == "" ||
		!jointSmokeInt(c.Budget, 8) || c.Space != "14" || !jointSmokeInt(c.SelectedAttempt, 0) || len(c.Attempts) != 1 ||
		!slices.Equal(c.Kinds, []string{"record_mask", "source_search_index"}) || !jointSmokeBool(c.Initial.Model.Loaded, false) ||
		!jointSmokeBool(e.Replayed, replay) || !jointSmokeInt(e.Calls, 0) {
		return "", fmt.Errorf("search checkpoint source, candidate space or inference differs")
	}
	a := c.Attempts[0]
	if a.Rejection != nil || !slices.Equal(a.Masks, []int{1, 0}) || !jointSmokeInt(a.Passed, 2) || !jointSmokeInt(a.Total, 2) ||
		a.Runtime.SHA != c.Selected.SHA || e.Runtime.SHA != c.Selected.SHA {
		return "", fmt.Errorf("search checkpoint selected a different helper combination")
	}
	construction, err := jointSmokeCases(feedback, 1)
	if err != nil {
		return "", err
	}
	cases, err := jointSmokeCases(evaluation, 3)
	if err != nil {
		return "", err
	}
	for i, runtime := range []jointSmokeRuntime{a.Runtime, e.Runtime} {
		rows := construction
		if i == 1 {
			rows = cases
		}
		if err := validateSearchRuntimeActivity(runtime, rows, false, "searchcheckpoint://activity/main"); err != nil {
			return "", err
		}
	}
	separation := envelope.Evaluation.Runtime.Separation
	if separation.Status != "UNKNOWN" || !jointSmokeInt(separation.Unknown, 3) ||
		!jointSmokeInt(separation.Overlap, 0) || !jointSmokeInt(separation.Disjoint, 0) {
		return "", fmt.Errorf("search checkpoint erased its incomplete input-exposure scope")
	}
	var detail checkpointDetails
	if err := json.Unmarshal(envelope.Construction, &detail); err != nil {
		return "", err
	}
	if err := validateCheckpointPool(detail, c.OriginalSHA); err != nil {
		return "", err
	}
	return constructionSHA, nil
}

// JSON object ordering may normalize when a saved report is decoded. Preserve
// every value, including exact integer tokens, while comparing the full history.
func checkpointConstructionDigest(raw []byte) (string, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return "", err
	}
	return digestValue(value)
}

func validateCheckpointPool(d checkpointDetails, original string) error {
	if len(d.Initial.Preparations) != 2 || len(d.Attempts) != 1 || len(d.Attempts[0].Search) != 1 {
		return fmt.Errorf("search checkpoint observation is missing")
	}
	step := d.Initial.Preparations[1]
	s := step.Generation.Report.Search
	if step.Generation.Report.Activity != "Build" || s == nil || s.Pool == nil || s.Generation == nil ||
		step.InputSHA == "" || step.InputSHA == original || step.InputSHA != s.Original ||
		s.Plan == "" || s.Pool.PlanSHA256 != s.Plan || s.Pool.AttemptBudget != 8 || !jointSmokeInt(s.Calls, 0) {
		return fmt.Errorf("search pool lost its prepared helper checkpoint")
	}
	generationSHA, err := digestValue(s.Generation)
	if err != nil || generationSHA != s.Pool.GenerationSHA256 {
		return fmt.Errorf("search generation reference differs")
	}
	expressions := []string{"1", "input + -1", "input", "6", "-input", "input + 4", "input - -4"}
	if len(s.Pool.Candidates) != len(expressions) {
		return fmt.Errorf("search initial pool size differs")
	}
	for i, candidate := range s.Pool.Candidates {
		sum := sha256.Sum256([]byte(expressions[i]))
		id := fmt.Sprintf("generated_%x", sum[:6])
		if candidate.Expression != expressions[i] || candidate.ID != id {
			return fmt.Errorf("search initial pool order differs")
		}
	}
	x := d.Attempts[0].Search[0]
	if x.Activity != "Build" || x.PlanningSourceSHA256 != step.InputSHA || x.InputSourceSHA256 != step.InputSHA ||
		x.PlanSHA256 != s.Plan || s.Selected != s.Pool.Candidates[0].ID || x.Attempt.CandidateID != s.Selected || x.Attempt.Expression != "1" ||
		!x.Attempt.TypecheckPassed || !x.Attempt.ScoringCompleted || x.Attempt.TestCasesPassed != 1 || x.Attempt.TestCasesTotal != 1 ||
		len(x.Attempt.CaseResults) != 1 || x.LocalCalls == nil || x.LocalCalls.Status != "RECORDED" || len(x.LocalCalls.Evaluations) != 1 {
		return fmt.Errorf("caller search candidate lost its original pool or current local evaluation")
	}
	row := x.Attempt.CaseResults[0]
	if row.Input != 2 || row.Expected != 6 || row.Actual != 6 || !row.Passed {
		return fmt.Errorf("local helper case differs")
	}
	pass := x.LocalCalls.Evaluations[0]
	if pass == nil || pass.Status != "RECORDED" || pass.Phase != "candidate" || pass.CandidateID != s.Selected || !pass.EvaluationCompleted ||
		pass.CasesStarted != 1 || pass.CasesCompleted != 1 || len(pass.Calls) != 1 ||
		pass.Calls[0].CaseIndex != 0 || pass.Calls[0].ActivityID != "searchcheckpoint://activity/seed" ||
		!samePackageSourceValue(pass.Calls[0].Inputs, []byte("[3]")) {
		return fmt.Errorf("caller search probe history differs from the actual selected evaluation")
	}
	return nil
}
