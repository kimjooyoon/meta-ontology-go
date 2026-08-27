package causalci

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	CIEvidenceObservationSchema          = "gooo/causal-ci-selection-ci-evidence-observation/v1"
	CIEvidenceAdjudicationSchema         = "gooo/causal-ci-selection-ci-evidence-adjudication/v1"
	CIEvidenceAdjudicationScope          = "CI_CAUSAL_EVIDENCE"
	CIEvidenceProcessExited              = "EXITED"
	CIEvidenceOutcomeObserved            = "OBSERVED_HTTP_RESPONSE"
	CIEvidenceOutcomeOpen                = "OPEN"
	CIEvidenceReasonHTTPObserved         = "HTTP_RESPONSE_OBSERVED"
	CIEvidenceReasonPermission           = "CI_PERMISSION_DENIED"
	CIEvidenceReasonMalformed            = "MALFORMED_GITHUB_RESPONSE"
	CIEvidenceReasonMissing              = "MISSING_DOWNSTREAM_ARTIFACT"
	CIEvidenceCausalChild                = "CAUSAL_CHILD_OF_API_FAILURE"
	CIEvidenceFixtureScenario            = "FIXTURE_SCENARIO"
	CIEvidenceResolutionExact            = "EXACT"
	CIEvidenceResolutionLowered          = "LOWER_RESOLUTION"
	CIEvidenceDecisionObserved           = "OBSERVED"
	CIEvidenceDecisionFailClosed         = "FAIL_CLOSED"
	CIEvidenceEndpointRunsList           = "ACTIONS_WORKFLOW_RUNS_LIST"
	CIEvidenceEndpointArtifacts          = "ACTIONS_RUN_ARTIFACTS_LIST"
	CIEvidenceArtifactKind               = "ARTIFACT"
	CIEvidenceUploadStepKind             = "UPLOAD_STEP"
	CIEvidenceSourceCurrent              = "CURRENT_EVIDENCE"
	CIEvidenceSourceHistorical           = "HISTORICAL_FIXTURE"
	CIEvidenceSourceSynthetic            = "SYNTHETIC_FIXTURE"
	CIEvidenceReasonCurrentUnavailable   = "CURRENT_CI_EVIDENCE_UNAVAILABLE"
	CIEvidenceReasonPaginationIncomplete = "WORKFLOW_RUN_PAGINATION_INCOMPLETE"
	CIEvidenceCausalRelationUnknown      = "CAUSAL_RELATION_UNKNOWN"
)

// CIEvidenceObservation is a raw process observation. It contains response
// bytes and process facts, but no semantic decision, resolution, or reason.
// This keeps API transport failures from becoming producer conclusions.
type CIEvidenceObservation struct {
	Schema      string           `json:"schema"`
	Repository  string           `json:"repository"`
	RunID       int64            `json:"run_id"`
	JobID       int64            `json:"job_id"`
	JobName     string           `json:"job_name"`
	SourceClass string           `json:"source_class"`
	Cases       []CIEvidenceCase `json:"cases"`
}

type CIEvidenceCase struct {
	RunID              int64                `json:"run_id"`
	JobID              int64                `json:"job_id"`
	JobName            string               `json:"job_name"`
	ID                 string               `json:"id"`
	Endpoint           string               `json:"endpoint"`
	EndpointClass      string               `json:"endpoint_class"`
	SourceClass        string               `json:"source_class"`
	RequiredPermission string               `json:"required_permission"`
	HTTPStatus         int                  `json:"http_status"`
	ProcessExitCode    int                  `json:"process_exit_code"`
	ProcessStatus      string               `json:"process_status"`
	ResponseBody       string               `json:"response_body"`
	Stdout             string               `json:"stdout"`
	Stderr             string               `json:"stderr"`
	Artifacts          []CIEvidenceArtifact `json:"artifacts,omitempty"`
	Pages              []CIEvidencePage     `json:"pages,omitempty"`
	Provenance         string               `json:"provenance"`
}

type CIEvidencePage struct {
	EndpointClass  string `json:"endpoint_class"`
	URL            string `json:"url"`
	PageNumber     int    `json:"page_number"`
	HTTPStatus     int    `json:"http_status"`
	BodyDigest     string `json:"body_digest"`
	BodyBytes      int    `json:"body_bytes"`
	NextLinkDigest string `json:"next_link_digest"`
	Observed       bool   `json:"observed"`
}

type CIEvidenceArtifact struct {
	ID                 string `json:"id"`
	Path               string `json:"path"`
	Kind               string `json:"kind"`
	Expected           bool   `json:"expected"`
	Present            bool   `json:"present"`
	DependencyObserved bool   `json:"dependency_observed"`
	DependencyEvidence string `json:"dependency_evidence,omitempty"`
}

// CIEvidenceRow is the derived, independently adjudicated representation of
// one API-process observation. Its outcome is deliberately separate from the
// causal CI plan conformance decision.
type CIEvidenceRow struct {
	RunID                 int64            `json:"run_id"`
	JobID                 int64            `json:"job_id"`
	JobName               string           `json:"job_name"`
	CaseID                string           `json:"case_id"`
	Endpoint              string           `json:"endpoint"`
	EndpointClass         string           `json:"endpoint_class"`
	SourceClass           string           `json:"source_class"`
	RequiredPermission    string           `json:"required_permission"`
	ObservedHTTPStatus    int              `json:"observed_http_status"`
	ObservedProcessExit   int              `json:"observed_process_exit_code"`
	ObservedProcessStatus string           `json:"observed_process_status"`
	ResponseDigest        string           `json:"response_digest"`
	ResponseBytes         int              `json:"response_bytes"`
	ObservedStdoutDigest  string           `json:"observed_stdout_digest"`
	ObservedStdoutBytes   int              `json:"observed_stdout_bytes"`
	ObservedStderrDigest  string           `json:"observed_stderr_digest"`
	ObservedStderrBytes   int              `json:"observed_stderr_bytes"`
	Outcome               string           `json:"outcome"`
	Decision              string           `json:"decision"`
	Resolution            string           `json:"resolution"`
	Coordinate            Coordinate       `json:"coordinate"`
	EvidenceDigest        string           `json:"evidence_digest"`
	Provenance            string           `json:"provenance"`
	PageCount             int              `json:"page_count"`
	PageInventory         []CIEvidencePage `json:"page_inventory"`
	PageInventoryComplete bool             `json:"page_inventory_complete"`
}

type CIEvidenceArtifactObservation struct {
	ArtifactID            string `json:"artifact_id"`
	Path                  string `json:"path"`
	Kind                  string `json:"kind"`
	Expected              bool   `json:"expected"`
	Present               bool   `json:"present"`
	ParentCaseID          string `json:"parent_case_id"`
	CausalRelation        string `json:"causal_relation"`
	DependencyObserved    bool   `json:"dependency_observed"`
	DependencyEvidence    string `json:"dependency_evidence,omitempty"`
	ObservedProcessExit   int    `json:"observed_process_exit_code"`
	ObservedProcessStatus string `json:"observed_process_status"`
	EvidenceDigest        string `json:"evidence_digest"`
	Provenance            string `json:"provenance"`
}

type CIEvidenceAdjudication struct {
	Schema                           string                          `json:"schema"`
	Scope                            string                          `json:"scope"`
	ObservationDigest                string                          `json:"observation_digest"`
	ExpectedCaseIDs                  []string                        `json:"expected_case_ids"`
	ObservedCaseIDs                  []string                        `json:"observed_case_ids"`
	Rows                             []CIEvidenceRow                 `json:"rows"`
	ArtifactObservations             []CIEvidenceArtifactObservation `json:"artifact_observations"`
	ActualCaseID                     string                          `json:"actual_case_id"`
	ActualRootCauseNumerator         int                             `json:"actual_root_cause_numerator"`
	ActualRootCauseDenominator       int                             `json:"actual_root_cause_denominator"`
	ExpectedDownstreamArtifactIDs    []string                        `json:"expected_downstream_artifact_ids"`
	ObservedDownstreamArtifactIDs    []string                        `json:"observed_downstream_artifact_ids"`
	DownstreamObservationNumerator   int                             `json:"downstream_artifact_observation_numerator"`
	DownstreamObservationDenominator int                             `json:"downstream_artifact_observation_denominator"`
	DownstreamMissingNumerator       *int                            `json:"downstream_missing_artifact_numerator"`
	DownstreamMissingDenominator     *int                            `json:"downstream_missing_artifact_denominator"`
	ActualResolution                 string                          `json:"actual_resolution"`
	ActualDecision                   string                          `json:"actual_decision"`
	ActualCoordinate                 Coordinate                      `json:"actual_coordinate"`
	CurrentOutcome                   string                          `json:"current_outcome"`
	CurrentPermissionDenominator     int                             `json:"current_permission_denial_denominator"`
	CurrentPermissionNumerator       int                             `json:"current_permission_denial_numerator"`
	HistoricalPermissionDenominator  int                             `json:"historical_permission_denial_denominator"`
	HistoricalPermissionNumerator    int                             `json:"historical_permission_denial_numerator"`
	CurrentSourceNumerator           int                             `json:"current_source_numerator"`
	CurrentSourceDenominator         int                             `json:"current_source_denominator"`
	HistoricalSourceNumerator        int                             `json:"historical_source_numerator"`
	HistoricalSourceDenominator      int                             `json:"historical_source_denominator"`
	SyntheticSourceNumerator         int                             `json:"synthetic_source_numerator"`
	SyntheticSourceDenominator       int                             `json:"synthetic_source_denominator"`
	CurrentPageCount                 int                             `json:"current_page_count"`
	CurrentPageInventoryComplete     bool                            `json:"current_page_inventory_complete"`
	Digest                           string                          `json:"digest"`
}

// ValidateCIEvidenceAdjudication re-seals an independently produced receipt.
// Consumers must verify the digest after decoding instead of trusting a
// producer-provided decision or root-cause count.
func ValidateCIEvidenceAdjudication(value CIEvidenceAdjudication) error {
	if value.Schema != CIEvidenceAdjudicationSchema || value.Scope != CIEvidenceAdjudicationScope || value.Digest == "" {
		return fmt.Errorf("malformed CI evidence adjudication")
	}
	stored := value.Digest
	value.Digest = ""
	computed, err := digestJSON(value)
	if err != nil || computed != stored {
		return fmt.Errorf("CI evidence adjudication digest mismatch")
	}
	return nil
}

// AdjudicateCIEvidence derives transport and causal classifications only
// from raw observations. In particular, HTTP 403 is never a semantic
// contradiction or fixed point, and child artifact absence cannot become a
// second root cause.
func AdjudicateCIEvidence(observations []CIEvidenceObservation, actualCaseID string) (CIEvidenceAdjudication, error) {
	expected := []string{"historical-pagination-incomplete", "malformed-http-200", "missing-artifact", "normal-http-200", "permission-denied-403"}
	expectedDownstreamArtifacts := expectedDownstreamArtifactIDs()
	cases := make([]CIEvidenceCase, 0, len(observations))
	var current *CIEvidenceCase
	var observationRaw []byte
	for _, observation := range observations {
		if observation.Schema != CIEvidenceObservationSchema || observation.Repository != "kimjooyoon/meta-ontology-go" || observation.JobName == "" || len(observation.Cases) != 1 {
			return CIEvidenceAdjudication{}, fmt.Errorf("malformed CI evidence observation")
		}
		value := observation.Cases[0]
		expectedRunID, expectedJobID, expectedJobName := int64(33088310894), int64(98574425650), "language-concept-artifact"
		if observation.SourceClass == CIEvidenceSourceCurrent || (observation.SourceClass == CIEvidenceSourceHistorical && value.ID == "historical-pagination-incomplete") {
			expectedRunID, expectedJobID = 33098087709, 98608698224
		}
		if observation.RunID != expectedRunID || observation.JobID != expectedJobID || observation.JobName != expectedJobName || observation.RunID != value.RunID || observation.JobID != value.JobID || observation.JobName != value.JobName || observation.SourceClass != value.SourceClass {
			return CIEvidenceAdjudication{}, fmt.Errorf("CI evidence top-level identity mismatch for %q", value.ID)
		}
		observationDigest, err := digestJSON(observation)
		if err != nil {
			return CIEvidenceAdjudication{}, err
		}
		observationRaw = append(observationRaw, []byte(observationDigest)...)
		switch value.SourceClass {
		case CIEvidenceSourceCurrent:
			if current != nil {
				return CIEvidenceAdjudication{}, fmt.Errorf("duplicate current CI evidence case")
			}
			copy := value
			current = &copy
		case CIEvidenceSourceHistorical, CIEvidenceSourceSynthetic:
			cases = append(cases, value)
		default:
			return CIEvidenceAdjudication{}, fmt.Errorf("unknown CI evidence source class for %q", value.ID)
		}
	}
	if len(cases) != len(expected) || !sameIDs(expected, caseIDs(cases)) {
		return CIEvidenceAdjudication{}, fmt.Errorf("CI evidence case inventory mismatch")
	}
	seen := map[string]struct{}{}
	rows := make([]CIEvidenceRow, 0, len(cases)+1)
	children := make([]CIEvidenceArtifactObservation, 0)
	for _, value := range cases {
		if err := validateCIEvidenceCase(value); err != nil {
			return CIEvidenceAdjudication{}, err
		}
		if _, exists := seen[value.ID]; exists {
			return CIEvidenceAdjudication{}, fmt.Errorf("duplicate CI evidence case %q", value.ID)
		}
		seen[value.ID] = struct{}{}
		rows = append(rows, deriveCIEvidenceRow(value))
		for _, artifact := range value.Artifacts {
			child, ok, err := deriveCIEvidenceArtifact(value, artifact)
			if err != nil {
				return CIEvidenceAdjudication{}, err
			}
			if ok {
				children = append(children, child)
			}
		}
	}
	if actualCaseID != "" && (current == nil || current.ID != actualCaseID) {
		return CIEvidenceAdjudication{}, fmt.Errorf("actual CI evidence case is not current evidence")
	}
	var actual CIEvidenceRow
	if current == nil {
		actual = CIEvidenceRow{Outcome: CIEvidenceOutcomeOpen, Decision: CIEvidenceDecisionFailClosed, Resolution: CIEvidenceResolutionLowered, Coordinate: Coordinate{Stage: "proposal-promotion", Step: "fetch-github-evidence", Reason: CIEvidenceReasonCurrentUnavailable}}
	} else {
		if err := validateCurrentCIEvidenceCase(*current); err != nil {
			return CIEvidenceAdjudication{}, err
		}
		actual = deriveCIEvidenceRow(*current)
		rows = append(rows, actual)
		for _, artifact := range current.Artifacts {
			child, ok, err := deriveCIEvidenceArtifact(*current, artifact)
			if err != nil {
				return CIEvidenceAdjudication{}, err
			}
			if ok {
				children = append(children, child)
			}
		}
	}
	historicalPermission := 0
	historicalSources := 0
	syntheticSources := 0
	for _, row := range rows {
		switch row.SourceClass {
		case CIEvidenceSourceHistorical:
			historicalSources++
		case CIEvidenceSourceSynthetic:
			syntheticSources++
		}
		if row.SourceClass == CIEvidenceSourceHistorical && row.ObservedHTTPStatus == 403 && row.Coordinate.Reason == CIEvidenceReasonPermission {
			historicalPermission++
		}
	}
	var currentMissingNumerator *int
	var currentMissingDenominator *int
	observedDownstreamArtifacts := []string{}
	if current != nil {
		observedDownstreamArtifacts = observedDownstreamArtifactIDs(current.Artifacts)
		currentMissing := 0
		for _, artifact := range current.Artifacts {
			if artifact.Expected && !artifact.Present {
				currentMissing++
			}
		}
		currentMissingNumerator = &currentMissing
		currentMissingDenominator = intPointer(len(expectedDownstreamArtifacts))
	}
	result := CIEvidenceAdjudication{
		Schema: CIEvidenceAdjudicationSchema, Scope: CIEvidenceAdjudicationScope,
		ObservationDigest: digestBytes(observationRaw), ExpectedCaseIDs: expected,
		ObservedCaseIDs: caseIDs(cases), Rows: rows, ArtifactObservations: children,
		ActualRootCauseDenominator: 1, ExpectedDownstreamArtifactIDs: expectedDownstreamArtifacts,
		ObservedDownstreamArtifactIDs:    observedDownstreamArtifacts,
		DownstreamObservationNumerator:   len(observedDownstreamArtifacts),
		DownstreamObservationDenominator: len(expectedDownstreamArtifacts),
		DownstreamMissingNumerator:       currentMissingNumerator, DownstreamMissingDenominator: currentMissingDenominator,
		ActualResolution: actual.Resolution,
		ActualDecision:   actual.Decision, ActualCoordinate: actual.Coordinate,
		CurrentOutcome: actual.Outcome, CurrentPermissionDenominator: 1,
		HistoricalPermissionNumerator: historicalPermission, HistoricalPermissionDenominator: 1,
		CurrentSourceNumerator: boolInt(current != nil), CurrentSourceDenominator: 1,
		HistoricalSourceNumerator: historicalSources, HistoricalSourceDenominator: 2,
		SyntheticSourceNumerator: syntheticSources, SyntheticSourceDenominator: 3,
	}
	if current != nil {
		result.ActualCaseID = current.ID
		result.ActualRootCauseNumerator = 1
		result.CurrentPermissionNumerator = boolInt(currentRowIsPermissionDenied(actual))
		result.CurrentPageCount = len(current.Pages)
		result.CurrentPageInventoryComplete = currentPagesComplete(*current)
	}
	result.Digest, _ = digestJSON(result)
	return result, nil
}

func validateCIEvidenceCase(value CIEvidenceCase) error {
	if value.RunID < 1 || value.JobID < 1 || value.JobName == "" || value.ID == "" || value.Endpoint == "" || value.RequiredPermission != "actions: read" || value.HTTPStatus < 100 || value.HTTPStatus > 599 || value.ProcessStatus != CIEvidenceProcessExited || value.ProcessExitCode < 0 || value.Provenance == "" {
		return fmt.Errorf("malformed CI evidence case %q", value.ID)
	}
	if value.SourceClass == CIEvidenceSourceHistorical {
		if value.ID == "historical-pagination-incomplete" {
			if value.RunID != 33098087709 || value.JobID != 98608698224 || value.JobName != "language-concept-artifact" {
				return fmt.Errorf("historical CI evidence identity mismatch for %q", value.ID)
			}
		} else if value.RunID != 33088310894 || value.JobID != 98574425650 || value.JobName != "language-concept-artifact" {
			return fmt.Errorf("historical CI evidence identity mismatch for %q", value.ID)
		}
	}
	if value.SourceClass == CIEvidenceSourceCurrent &&
		(value.RunID != 33098087709 || value.JobID != 98608698224 || value.JobName != "language-concept-artifact" || value.ID != "current-pagination-incomplete") {
		return fmt.Errorf("current CI evidence identity mismatch for %q", value.ID)
	}
	if value.SourceClass == CIEvidenceSourceCurrent &&
		(!strings.HasPrefix(value.Provenance, "CURRENT_GITHUB_ACTIONS_OBSERVATION ") ||
			!strings.Contains(value.Provenance, "run=33098087709") ||
			!strings.Contains(value.Provenance, "job=98608698224") || !currentPagesComplete(value)) {
		return fmt.Errorf("current CI evidence is not an observed API/process capture for %q", value.ID)
	}
	if value.SourceClass == CIEvidenceSourceHistorical && !strings.HasPrefix(value.Provenance, "HISTORICAL_") {
		return fmt.Errorf("historical CI evidence provenance mismatch for %q", value.ID)
	}
	if value.SourceClass == CIEvidenceSourceSynthetic && !strings.HasPrefix(value.Provenance, "FIXED_") && !strings.HasPrefix(value.Provenance, "SYNTHETIC_") {
		return fmt.Errorf("synthetic CI evidence provenance mismatch for %q", value.ID)
	}
	if value.SourceClass == CIEvidenceSourceHistorical && value.ID != "permission-denied-403" && value.ID != "historical-pagination-incomplete" {
		return fmt.Errorf("historical CI evidence case mismatch for %q", value.ID)
	}
	if value.SourceClass == CIEvidenceSourceSynthetic && value.ID == "permission-denied-403" {
		return fmt.Errorf("permission case cannot be synthetic evidence")
	}
	expectedEndpointClass := CIEvidenceEndpointRunsList
	if strings.Contains(value.Endpoint, "/actions/runs/") && strings.Contains(value.Endpoint, "/artifacts?") {
		expectedEndpointClass = CIEvidenceEndpointArtifacts
	}
	if value.EndpointClass != expectedEndpointClass {
		return fmt.Errorf("CI evidence endpoint class mismatch for %q", value.ID)
	}
	for _, page := range value.Pages {
		if page.EndpointClass != value.EndpointClass || page.URL == "" || page.PageNumber < 1 || page.HTTPStatus < 100 || page.HTTPStatus > 599 || page.BodyBytes < 0 || page.BodyDigest == "" || page.NextLinkDigest == "" {
			return fmt.Errorf("malformed CI evidence page for %q", value.ID)
		}
		if page.Observed && (!strings.HasPrefix(page.BodyDigest, "sha256:") || !strings.HasPrefix(page.NextLinkDigest, "sha256:")) {
			return fmt.Errorf("observed CI evidence page digest is invalid for %q", value.ID)
		}
	}
	return nil
}

func validateCurrentCIEvidenceCase(value CIEvidenceCase) error {
	if value.SourceClass != CIEvidenceSourceCurrent || value.HTTPStatus != 200 || value.ProcessExitCode != 1 || !strings.Contains(value.Stderr, "workflow run pagination incomplete") || len(value.Pages) == 0 {
		return fmt.Errorf("current CI evidence is not the observed pagination failure")
	}
	if err := validateDownstreamArtifactInventory(value.Artifacts); err != nil {
		return err
	}
	return validateCIEvidenceCase(value)
}

func expectedDownstreamArtifactIDs() []string {
	return []string{
		"language-readiness-predecessor-selection-a",
		"language-readiness-baseline-reference-a",
		"language-readiness-foundation-seed-a",
	}
}

func intPointer(value int) *int { return &value }

func observedDownstreamArtifactIDs(artifacts []CIEvidenceArtifact) []string {
	observed := make(map[string]bool, len(artifacts))
	for _, artifact := range artifacts {
		if artifact.Expected {
			observed[artifact.ID] = true
		}
	}
	result := make([]string, 0, len(expectedDownstreamArtifactIDs()))
	for _, id := range expectedDownstreamArtifactIDs() {
		if observed[id] {
			result = append(result, id)
		}
	}
	return result
}

func validateDownstreamArtifactInventory(artifacts []CIEvidenceArtifact) error {
	expected := expectedDownstreamArtifactIDs()
	if len(artifacts) != len(expected) {
		return fmt.Errorf("downstream artifact inventory mismatch")
	}
	expectedSet := make(map[string]bool, len(expected))
	for _, id := range expected {
		expectedSet[id] = true
	}
	seen := make(map[string]bool, len(artifacts))
	for _, artifact := range artifacts {
		if !artifact.Expected || !expectedSet[artifact.ID] || seen[artifact.ID] {
			return fmt.Errorf("downstream artifact inventory mismatch for %q", artifact.ID)
		}
		seen[artifact.ID] = true
	}
	for _, id := range expected {
		if !seen[id] {
			return fmt.Errorf("downstream artifact inventory missing %q", id)
		}
	}
	return nil
}

func deriveCIEvidenceRow(value CIEvidenceCase) CIEvidenceRow {
	pageInventory := append([]CIEvidencePage(nil), value.Pages...)
	row := CIEvidenceRow{RunID: value.RunID, JobID: value.JobID, JobName: value.JobName, CaseID: value.ID, Endpoint: value.Endpoint, EndpointClass: value.EndpointClass, SourceClass: value.SourceClass, RequiredPermission: value.RequiredPermission, ObservedHTTPStatus: value.HTTPStatus, ObservedProcessExit: value.ProcessExitCode, ObservedProcessStatus: value.ProcessStatus, ResponseDigest: digestBytes([]byte(value.ResponseBody)), ResponseBytes: len([]byte(value.ResponseBody)), ObservedStdoutDigest: digestBytes([]byte(value.Stdout)), ObservedStdoutBytes: len([]byte(value.Stdout)), ObservedStderrDigest: digestBytes([]byte(value.Stderr)), ObservedStderrBytes: len([]byte(value.Stderr)), Provenance: value.Provenance, PageCount: len(pageInventory), PageInventory: pageInventory, PageInventoryComplete: currentPagesComplete(value)}
	if strings.Contains(value.Stderr, "workflow run pagination incomplete") {
		row.Outcome = CIEvidenceOutcomeOpen
		if value.SourceClass != CIEvidenceSourceCurrent {
			row.Outcome = ExecutionUnknown
		}
		row.Decision = CIEvidenceDecisionFailClosed
		row.Resolution = CIEvidenceResolutionLowered
		row.Coordinate = Coordinate{Stage: "proposal-promotion", Step: "fetch-github-evidence", Reason: CIEvidenceReasonPaginationIncomplete}
	} else if value.HTTPStatus == 403 {
		row.Outcome = ExecutionUnknown
		row.Decision = CIEvidenceDecisionFailClosed
		row.Resolution = CIEvidenceResolutionLowered
		if isPermissionDeniedEvidence(value) {
			row.Coordinate = Coordinate{Stage: "proposal-promotion", Step: "fetch-github-evidence", Reason: CIEvidenceReasonPermission}
		} else {
			row.Coordinate = Coordinate{Stage: "proposal-promotion", Step: "fetch-github-evidence", Reason: "GITHUB_HTTP_403_UNKNOWN"}
		}
	} else if value.HTTPStatus < 200 || value.HTTPStatus >= 300 {
		row.Outcome = ExecutionUnknown
		row.Decision = CIEvidenceDecisionFailClosed
		row.Resolution = CIEvidenceResolutionLowered
		row.Coordinate = Coordinate{Stage: "proposal-promotion", Step: "fetch-github-evidence", Reason: "GITHUB_HTTP_FAILURE"}
	} else if (value.EndpointClass == CIEvidenceEndpointRunsList && !looksLikeRunsListJSON(value.ResponseBody)) || (value.EndpointClass == CIEvidenceEndpointArtifacts && !looksLikeArtifactsListJSON(value.ResponseBody)) {
		row.Outcome = ExecutionUnknown
		row.Decision = CIEvidenceDecisionFailClosed
		row.Resolution = CIEvidenceResolutionLowered
		row.Coordinate = Coordinate{Stage: "proposal-promotion", Step: "decode-github-evidence", Reason: CIEvidenceReasonMalformed}
	} else if hasMissingArtifact(value.Artifacts) {
		row.Outcome = ExecutionUnknown
		row.Decision = CIEvidenceDecisionFailClosed
		row.Resolution = CIEvidenceResolutionLowered
		row.Coordinate = Coordinate{Stage: "proposal-promotion", Step: "bind-github-artifact", Reason: CIEvidenceReasonMissing}
	} else {
		row.Outcome = CIEvidenceOutcomeObserved
		row.Decision = CIEvidenceDecisionObserved
		row.Resolution = CIEvidenceResolutionExact
		row.Coordinate = Coordinate{Stage: "proposal-promotion", Step: "fetch-github-evidence", Reason: CIEvidenceReasonHTTPObserved}
	}
	pageDigest, _ := digestJSON(value.Pages)
	row.EvidenceDigest, _ = digestJSON(struct {
		CaseID, Endpoint, EndpointClass, SourceClass, Permission, ResponseDigest, StdoutDigest, StderrDigest, PageDigest, Outcome, Decision, Resolution string
		HTTPStatus, ProcessExit, ResponseBytes, StdoutBytes, StderrBytes, PageCount                                                                     int
	}{row.CaseID, row.Endpoint, row.EndpointClass, row.SourceClass, row.RequiredPermission, row.ResponseDigest, row.ObservedStdoutDigest, row.ObservedStderrDigest, pageDigest, row.Outcome, row.Decision, row.Resolution, row.ObservedHTTPStatus, row.ObservedProcessExit, row.ResponseBytes, row.ObservedStdoutBytes, row.ObservedStderrBytes, row.PageCount})
	return row
}

func isPermissionDeniedEvidence(value CIEvidenceCase) bool {
	if value.HTTPStatus != 403 || value.EndpointClass != CIEvidenceEndpointRunsList || value.RequiredPermission != "actions: read" || value.ProcessStatus != CIEvidenceProcessExited || value.ProcessExitCode == 0 || !strings.Contains(value.Stderr, "403") {
		return false
	}
	var response struct {
		Message string `json:"message"`
	}
	return json.Unmarshal([]byte(value.ResponseBody), &response) == nil && response.Message == "Resource not accessible by integration"
}

func deriveCIEvidenceArtifact(value CIEvidenceCase, artifact CIEvidenceArtifact) (CIEvidenceArtifactObservation, bool, error) {
	if artifact.ID == "" || artifact.Path == "" || (artifact.Kind != CIEvidenceArtifactKind && artifact.Kind != CIEvidenceUploadStepKind) {
		return CIEvidenceArtifactObservation{}, false, fmt.Errorf("malformed CI evidence artifact")
	}
	if !artifact.Expected || artifact.Present {
		return CIEvidenceArtifactObservation{}, false, nil
	}
	relation := CIEvidenceFixtureScenario
	if value.SourceClass == CIEvidenceSourceCurrent {
		relation = CIEvidenceCausalRelationUnknown
	}
	if artifact.DependencyObserved && artifact.DependencyEvidence != "" {
		relation = CIEvidenceCausalChild
	}
	child := CIEvidenceArtifactObservation{ArtifactID: artifact.ID, Path: artifact.Path, Kind: artifact.Kind, Expected: artifact.Expected, Present: artifact.Present, ParentCaseID: value.ID, CausalRelation: relation, DependencyObserved: artifact.DependencyObserved, DependencyEvidence: artifact.DependencyEvidence, ObservedProcessExit: value.ProcessExitCode, ObservedProcessStatus: value.ProcessStatus, Provenance: value.Provenance}
	child.EvidenceDigest, _ = digestJSON(child)
	return child, true, nil
}

func currentRowIsPermissionDenied(value CIEvidenceRow) bool {
	return value.SourceClass == CIEvidenceSourceCurrent && value.ObservedHTTPStatus == 403 && value.Coordinate.Reason == CIEvidenceReasonPermission
}

func currentPagesComplete(value CIEvidenceCase) bool {
	if len(value.Pages) == 0 {
		return false
	}
	for _, page := range value.Pages {
		if !page.Observed {
			return false
		}
	}
	return true
}

func looksLikeRunsListJSON(value string) bool {
	var payload struct {
		TotalCount   *int            `json:"total_count"`
		WorkflowRuns json.RawMessage `json:"workflow_runs"`
	}
	if json.Unmarshal([]byte(value), &payload) != nil || payload.TotalCount == nil || *payload.TotalCount < 0 || len(payload.WorkflowRuns) == 0 || string(payload.WorkflowRuns) == "null" {
		return false
	}
	var runs []json.RawMessage
	return json.Unmarshal(payload.WorkflowRuns, &runs) == nil
}

func looksLikeArtifactsListJSON(value string) bool {
	var payload struct {
		TotalCount *int            `json:"total_count"`
		Artifacts  json.RawMessage `json:"artifacts"`
	}
	if json.Unmarshal([]byte(value), &payload) != nil || payload.TotalCount == nil || *payload.TotalCount < 0 || len(payload.Artifacts) == 0 || string(payload.Artifacts) == "null" {
		return false
	}
	var artifacts []json.RawMessage
	return json.Unmarshal(payload.Artifacts, &artifacts) == nil
}

func caseIDs(values []CIEvidenceCase) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.ID)
	}
	return result
}

func findCIEvidenceRow(values []CIEvidenceRow, id string) (CIEvidenceRow, bool) {
	for _, value := range values {
		if value.CaseID == id {
			return value, true
		}
	}
	return CIEvidenceRow{}, false
}

func indexOfCase(values []CIEvidenceCase, id string) int {
	for index, value := range values {
		if value.ID == id {
			return index
		}
	}
	return -1
}

func hasMissingArtifact(values []CIEvidenceArtifact) bool {
	for _, value := range values {
		if value.Expected && !value.Present {
			return true
		}
	}
	return false
}

func looksLikeJSON(value string) bool {
	trimmed := strings.TrimSpace(value)
	return json.Valid([]byte(trimmed)) && strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")
}
