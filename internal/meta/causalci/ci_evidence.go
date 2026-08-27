package causalci

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	CIEvidenceObservationSchema  = "gooo/causal-ci-selection-ci-evidence-observation/v1"
	CIEvidenceAdjudicationSchema = "gooo/causal-ci-selection-ci-evidence-adjudication/v1"
	CIEvidenceAdjudicationScope  = "CI_CAUSAL_EVIDENCE"
	CIEvidenceProcessExited      = "EXITED"
	CIEvidenceOutcomeObserved    = "OBSERVED_HTTP_RESPONSE"
	CIEvidenceReasonHTTPObserved = "HTTP_RESPONSE_OBSERVED"
	CIEvidenceReasonPermission   = "CI_PERMISSION_DENIED"
	CIEvidenceReasonMalformed    = "MALFORMED_GITHUB_RESPONSE"
	CIEvidenceReasonMissing      = "MISSING_DOWNSTREAM_ARTIFACT"
	CIEvidenceCausalChild        = "CAUSAL_CHILD_OF_API_FAILURE"
	CIEvidenceFixtureScenario    = "FIXTURE_SCENARIO"
	CIEvidenceResolutionExact    = "EXACT"
	CIEvidenceResolutionLowered  = "LOWER_RESOLUTION"
)

// CIEvidenceObservation is a raw process observation. It contains response
// bytes and process facts, but no semantic decision, resolution, or reason.
// This keeps API transport failures from becoming producer conclusions.
type CIEvidenceObservation struct {
	Schema     string           `json:"schema"`
	Repository string           `json:"repository"`
	RunID      int64            `json:"run_id"`
	JobID      int64            `json:"job_id"`
	JobName    string           `json:"job_name"`
	Cases      []CIEvidenceCase `json:"cases"`
}

type CIEvidenceCase struct {
	RunID              int64                `json:"run_id"`
	JobID              int64                `json:"job_id"`
	JobName            string               `json:"job_name"`
	ID                 string               `json:"id"`
	Endpoint           string               `json:"endpoint"`
	RequiredPermission string               `json:"required_permission"`
	HTTPStatus         int                  `json:"http_status"`
	ProcessExitCode    int                  `json:"process_exit_code"`
	ProcessStatus      string               `json:"process_status"`
	ResponseBody       string               `json:"response_body"`
	Stdout             string               `json:"stdout"`
	Stderr             string               `json:"stderr"`
	Artifacts          []CIEvidenceArtifact `json:"artifacts,omitempty"`
	Provenance         string               `json:"provenance"`
}

type CIEvidenceArtifact struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Expected bool   `json:"expected"`
	Present  bool   `json:"present"`
}

// CIEvidenceRow is the derived, independently adjudicated representation of
// one API-process observation. Its outcome is deliberately separate from the
// causal CI plan conformance decision.
type CIEvidenceRow struct {
	RunID                 int64      `json:"run_id"`
	JobID                 int64      `json:"job_id"`
	JobName               string     `json:"job_name"`
	CaseID                string     `json:"case_id"`
	Endpoint              string     `json:"endpoint"`
	RequiredPermission    string     `json:"required_permission"`
	ObservedHTTPStatus    int        `json:"observed_http_status"`
	ObservedProcessExit   int        `json:"observed_process_exit_code"`
	ObservedProcessStatus string     `json:"observed_process_status"`
	ResponseDigest        string     `json:"response_digest"`
	ResponseBytes         int        `json:"response_bytes"`
	ObservedStdoutDigest  string     `json:"observed_stdout_digest"`
	ObservedStdoutBytes   int        `json:"observed_stdout_bytes"`
	ObservedStderrDigest  string     `json:"observed_stderr_digest"`
	ObservedStderrBytes   int        `json:"observed_stderr_bytes"`
	Outcome               string     `json:"outcome"`
	Resolution            string     `json:"resolution"`
	Coordinate            Coordinate `json:"coordinate"`
	EvidenceDigest        string     `json:"evidence_digest"`
	Provenance            string     `json:"provenance"`
}

type CIEvidenceArtifactObservation struct {
	ArtifactID            string `json:"artifact_id"`
	Path                  string `json:"path"`
	Expected              bool   `json:"expected"`
	Present               bool   `json:"present"`
	ParentCaseID          string `json:"parent_case_id"`
	CausalRelation        string `json:"causal_relation"`
	ObservedProcessExit   int    `json:"observed_process_exit_code"`
	ObservedProcessStatus string `json:"observed_process_status"`
	EvidenceDigest        string `json:"evidence_digest"`
	Provenance            string `json:"provenance"`
}

type CIEvidenceAdjudication struct {
	Schema                       string                          `json:"schema"`
	Scope                        string                          `json:"scope"`
	ObservationDigest            string                          `json:"observation_digest"`
	ExpectedCaseIDs              []string                        `json:"expected_case_ids"`
	ObservedCaseIDs              []string                        `json:"observed_case_ids"`
	Rows                         []CIEvidenceRow                 `json:"rows"`
	ArtifactObservations         []CIEvidenceArtifactObservation `json:"artifact_observations"`
	ActualCaseID                 string                          `json:"actual_case_id"`
	ActualRootCauseNumerator     int                             `json:"actual_root_cause_numerator"`
	ActualRootCauseDenominator   int                             `json:"actual_root_cause_denominator"`
	DownstreamMissingNumerator   int                             `json:"downstream_missing_artifact_numerator"`
	DownstreamMissingDenominator int                             `json:"downstream_missing_artifact_denominator"`
	ActualResolution             string                          `json:"actual_resolution"`
	ActualCoordinate             Coordinate                      `json:"actual_coordinate"`
	Digest                       string                          `json:"digest"`
}

// AdjudicateCIEvidence derives transport and causal classifications only
// from raw observations. In particular, HTTP 403 is never a semantic
// contradiction or fixed point, and child artifact absence cannot become a
// second root cause.
func AdjudicateCIEvidence(observations []CIEvidenceObservation, actualCaseID string) (CIEvidenceAdjudication, error) {
	if actualCaseID == "" {
		return CIEvidenceAdjudication{}, fmt.Errorf("actual CI evidence case is required")
	}
	const expectedCount = 4
	expected := []string{"malformed-http-200", "missing-artifact", "normal-http-200", "permission-denied-403"}
	cases := make([]CIEvidenceCase, 0, len(observations))
	var observationRaw []byte
	for _, observation := range observations {
		if observation.Schema != CIEvidenceObservationSchema || observation.Repository == "" || observation.JobName == "" {
			return CIEvidenceAdjudication{}, fmt.Errorf("malformed CI evidence observation")
		}
		if len(observation.Cases) != 1 {
			return CIEvidenceAdjudication{}, fmt.Errorf("CI evidence fixture must contain exactly one case")
		}
		observationDigest, err := digestJSON(observation)
		if err != nil {
			return CIEvidenceAdjudication{}, err
		}
		observationRaw = append(observationRaw, []byte(observationDigest)...)
		cases = append(cases, observation.Cases[0])
	}
	if len(cases) != expectedCount || !sameIDs(expected, caseIDs(cases)) {
		return CIEvidenceAdjudication{}, fmt.Errorf("CI evidence case inventory mismatch")
	}
	seen := map[string]struct{}{}
	rows := make([]CIEvidenceRow, 0, len(cases))
	children := make([]CIEvidenceArtifactObservation, 0)
	for _, value := range cases {
		if value.RunID < 1 || value.JobID < 1 || value.JobName == "" || value.ID == "" || value.Endpoint == "" || value.RequiredPermission == "" || value.HTTPStatus < 100 || value.HTTPStatus > 599 || value.ProcessStatus != CIEvidenceProcessExited || value.ProcessExitCode < 0 || value.Provenance == "" {
			return CIEvidenceAdjudication{}, fmt.Errorf("malformed CI evidence case %q", value.ID)
		}
		if _, exists := seen[value.ID]; exists {
			return CIEvidenceAdjudication{}, fmt.Errorf("duplicate CI evidence case %q", value.ID)
		}
		seen[value.ID] = struct{}{}
		row := CIEvidenceRow{
			RunID: value.RunID, JobID: value.JobID, JobName: value.JobName, CaseID: value.ID, Endpoint: value.Endpoint, RequiredPermission: value.RequiredPermission,
			ObservedHTTPStatus: value.HTTPStatus, ObservedProcessExit: value.ProcessExitCode,
			ObservedProcessStatus: value.ProcessStatus, ResponseDigest: digestBytes([]byte(value.ResponseBody)),
			ResponseBytes: len([]byte(value.ResponseBody)), ObservedStdoutDigest: digestBytes([]byte(value.Stdout)),
			ObservedStdoutBytes: len([]byte(value.Stdout)), ObservedStderrDigest: digestBytes([]byte(value.Stderr)),
			ObservedStderrBytes: len([]byte(value.Stderr)), Provenance: value.Provenance,
		}
		switch {
		case value.HTTPStatus == 403:
			row.Outcome = ExecutionUnknown
			row.Resolution = CIEvidenceResolutionLowered
			row.Coordinate = Coordinate{Stage: "proposal-promotion", Step: "fetch-github-evidence", Reason: CIEvidenceReasonPermission}
		case value.HTTPStatus < 200 || value.HTTPStatus >= 300:
			row.Outcome = ExecutionUnknown
			row.Resolution = CIEvidenceResolutionLowered
			row.Coordinate = Coordinate{Stage: "proposal-promotion", Step: "fetch-github-evidence", Reason: "GITHUB_HTTP_FAILURE"}
		case !looksLikeJSON(value.ResponseBody):
			row.Outcome = ExecutionUnknown
			row.Resolution = CIEvidenceResolutionLowered
			row.Coordinate = Coordinate{Stage: "proposal-promotion", Step: "decode-github-evidence", Reason: CIEvidenceReasonMalformed}
		case hasMissingArtifact(value.Artifacts):
			row.Outcome = ExecutionUnknown
			row.Resolution = CIEvidenceResolutionLowered
			row.Coordinate = Coordinate{Stage: "proposal-promotion", Step: "bind-github-artifact", Reason: CIEvidenceReasonMissing}
		default:
			row.Outcome = CIEvidenceOutcomeObserved
			row.Resolution = CIEvidenceResolutionExact
			row.Coordinate = Coordinate{Stage: "proposal-promotion", Step: "fetch-github-evidence", Reason: CIEvidenceReasonHTTPObserved}
		}
		row.EvidenceDigest, _ = digestJSON(struct {
			CaseID, Endpoint, Permission, ResponseDigest, StdoutDigest, StderrDigest string
			HTTPStatus, ProcessExit, ResponseBytes, StdoutBytes, StderrBytes         int
			ProcessStatus, Outcome, Resolution                                       string
		}{row.CaseID, row.Endpoint, row.RequiredPermission, row.ResponseDigest, row.ObservedStdoutDigest, row.ObservedStderrDigest, row.ObservedHTTPStatus, row.ObservedProcessExit, row.ResponseBytes, row.ObservedStdoutBytes, row.ObservedStderrBytes, row.ObservedProcessStatus, row.Outcome, row.Resolution})
		rows = append(rows, row)
		for _, artifact := range value.Artifacts {
			if artifact.ID == "" || artifact.Path == "" || !artifact.Expected || artifact.Present {
				continue
			}
			relation := CIEvidenceFixtureScenario
			if value.ID == actualCaseID {
				relation = CIEvidenceCausalChild
			}
			child := CIEvidenceArtifactObservation{ArtifactID: artifact.ID, Path: artifact.Path, Expected: artifact.Expected, Present: artifact.Present, ParentCaseID: value.ID, CausalRelation: relation, ObservedProcessExit: value.ProcessExitCode, ObservedProcessStatus: value.ProcessStatus, Provenance: value.Provenance}
			child.EvidenceDigest, _ = digestJSON(child)
			children = append(children, child)
		}
	}
	actual, found := findCIEvidenceRow(rows, actualCaseID)
	if !found || actual.ObservedHTTPStatus != 403 || actual.Outcome != ExecutionUnknown || actual.Resolution != CIEvidenceResolutionLowered || actual.Coordinate.Stage != "proposal-promotion" || actual.Coordinate.Step != "fetch-github-evidence" || actual.Coordinate.Reason != CIEvidenceReasonPermission {
		return CIEvidenceAdjudication{}, fmt.Errorf("actual CI evidence is not a scoped permission-denied observation")
	}
	actualCase := cases[indexOfCase(cases, actualCaseID)]
	if len(actualCase.Artifacts) == 0 {
		return CIEvidenceAdjudication{}, fmt.Errorf("actual permission-denied observation has no downstream artifact inventory")
	}
	missingExpected := 0
	for _, artifact := range actualCase.Artifacts {
		if artifact.Expected {
			if !artifact.Present {
				missingExpected++
			}
		}
	}
	if missingExpected == 0 {
		return CIEvidenceAdjudication{}, fmt.Errorf("actual permission-denied observation has no missing downstream artifact")
	}
	result := CIEvidenceAdjudication{Schema: CIEvidenceAdjudicationSchema, Scope: CIEvidenceAdjudicationScope, ObservationDigest: digestBytes(observationRaw), ExpectedCaseIDs: expected, ObservedCaseIDs: caseIDs(cases), Rows: rows, ArtifactObservations: children, ActualCaseID: actualCaseID, ActualRootCauseNumerator: 1, ActualRootCauseDenominator: 1, DownstreamMissingNumerator: missingExpected, DownstreamMissingDenominator: missingExpected, ActualResolution: actual.Resolution, ActualCoordinate: actual.Coordinate}
	result.Digest, _ = digestJSON(result)
	return result, nil
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
