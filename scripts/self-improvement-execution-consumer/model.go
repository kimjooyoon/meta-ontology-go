package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"time"

	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
	grant "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutiongrant"
	input "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementvaluewitnessinput"
)

const reportSchema = "gooo/self-improvement-execution-consumption/v1"

type metadata struct {
	Repository             string `json:"repository"`
	SubjectSHA             string `json:"subject_sha"`
	GrantRunID             int64  `json:"grant_run_id"`
	GrantAttempt           int    `json:"grant_run_attempt"`
	GrantArtifactID        int64  `json:"grant_artifact_id"`
	GrantArtifactDigest    string `json:"grant_artifact_digest"`
	ContractRunID          int64  `json:"contract_run_id"`
	ContractAttempt        int    `json:"contract_run_attempt"`
	ContractArtifactID     int64  `json:"contract_artifact_id"`
	ContractArtifactDigest string `json:"contract_artifact_digest"`
}

type reservation struct {
	Schema         string `json:"schema"`
	GrantID        string `json:"grant_id"`
	RequestDigest  string `json:"request_digest"`
	SubjectSHA     string `json:"subject_sha"`
	SourceRunID    int64  `json:"source_run_id"`
	SourceAttempt  int    `json:"source_run_attempt"`
	Status         string `json:"status"`
	ExecutionCount int    `json:"execution_count"`
	ConsumedUses   int    `json:"consumed_uses"`
	RemainingUses  int    `json:"remaining_uses"`
	OneUseEnforced bool   `json:"one_use_enforced"`
	ReservedAt     string `json:"reserved_at"`
	Digest         string `json:"digest"`
}

type caseResult struct {
	ID                 string `json:"id"`
	Input              int64  `json:"input"`
	ExpectedOutput     int64  `json:"expected_output"`
	ActualOutput       *int64 `json:"actual_output,omitempty"`
	Passed             bool   `json:"passed"`
	ExecutionDigest    string `json:"execution_digest,omitempty"`
	ResultDigest       string `json:"result_digest,omitempty"`
	ElapsedNanoseconds int64  `json:"elapsed_nanoseconds"`
	Error              string `json:"error,omitempty"`
}

type consumptionReport struct {
	Schema                       string       `json:"schema"`
	GrantID                      string       `json:"grant_id"`
	GrantRequestDigest           string       `json:"grant_request_digest"`
	GrantReceiptDigest           string       `json:"grant_receipt_digest"`
	ReservationDigest            string       `json:"reservation_digest"`
	SubjectSHA                   string       `json:"subject_sha"`
	GrantRunID                   int64        `json:"grant_run_id"`
	GrantRunAttempt              int          `json:"grant_run_attempt"`
	GrantArtifactID              int64        `json:"grant_artifact_id"`
	GrantArtifactDigest          string       `json:"grant_artifact_digest"`
	ContractRunID                int64        `json:"contract_run_id"`
	ContractRunAttempt           int          `json:"contract_run_attempt"`
	ContractArtifactID           int64        `json:"contract_artifact_id"`
	ContractArtifactDigest       string       `json:"contract_artifact_digest"`
	CandidateStableID            string       `json:"candidate_stable_id"`
	CandidateDigest              string       `json:"candidate_digest"`
	ExecutionInputDigest         string       `json:"execution_input_digest"`
	ObservationDigest            string       `json:"observation_digest"`
	OperationID                  string       `json:"operation_id"`
	ExecutionCount               int          `json:"execution_count"`
	EvaluatorInvocations         int          `json:"evaluator_invocations"`
	ConsumedUses                 int          `json:"consumed_uses"`
	RemainingUses                int          `json:"remaining_uses"`
	OneUseEnforced               bool         `json:"one_use_enforced"`
	RepositoryWrites             int          `json:"repository_writes"`
	ExternalEffects              []string     `json:"external_effects"`
	PlanSourceDigest             string       `json:"plan_source_digest"`
	PlanIdentityDigest           string       `json:"plan_identity_digest"`
	CompileElapsedNanoseconds    int64        `json:"compile_elapsed_nanoseconds"`
	EvaluationElapsedNanoseconds int64        `json:"evaluation_elapsed_nanoseconds"`
	ElapsedNanoseconds           int64        `json:"elapsed_nanoseconds"`
	CaseCount                    int          `json:"case_count"`
	PassedCases                  int          `json:"passed_cases"`
	Cases                        []caseResult `json:"cases"`
	PerformanceImprovement       string       `json:"performance_improvement"`
	Digest                       string       `json:"digest"`
}

func readJSON(path string, target any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o600)
}

func digestJSON(value any) string {
	raw, _ := json.Marshal(value)
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func validDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func validSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func loadAndValidate(grantPath, contractPath, metadataPath string) (grant.LiveReport, contract.LiveReport, metadata, error) {
	var g grant.LiveReport
	var c contract.LiveReport
	var m metadata
	if err := readJSON(grantPath, &g); err != nil {
		return g, c, m, err
	}
	if err := readJSON(contractPath, &c); err != nil {
		return g, c, m, err
	}
	if err := readJSON(metadataPath, &m); err != nil {
		return g, c, m, err
	}
	if err := validate(g, c, m); err != nil {
		return g, c, m, err
	}
	return g, c, m, nil
}

func validate(g grant.LiveReport, c contract.LiveReport, m metadata) error {
	if g.Schema != grant.Schema || g.Digest == "" {
		return errors.New("v26 grant report identity is missing")
	}
	copyReport := g
	copyReport.Digest = ""
	if digestJSON(copyReport) != g.Digest {
		return errors.New("v26 grant report digest mismatch")
	}
	if c.Schema != contract.Schema || contract.VerifyResolution(c.ContractResolution) != nil ||
		!c.Verification.Verified || c.ContractResolution.Decision != contract.DecisionClosed ||
		c.ContractResolution.ExecutionInput == nil {
		return errors.New("v25 contract report is not verified and executable")
	}
	if err := input.Validate(*c.ContractResolution.ExecutionInput); err != nil {
		return err
	}
	if !g.Request.Source.ArtifactRetrieved || g.Request.Source.ArtifactExpired ||
		!g.Request.Source.ArtifactExpiryKnown || g.Request.Source.ArtifactRetrievalError != "" ||
		g.Request.Source.ArtifactDigest != g.Request.Source.ObservedArtifactDigest {
		return errors.New("v26 grant does not bind a retrieved exact v25 artifact")
	}
	if g.Request.Source.Repository != m.Repository || g.Request.Source.WorkflowRunID != m.ContractRunID ||
		g.Request.Source.WorkflowRunAttempt != m.ContractAttempt || g.Request.Source.ArtifactID != m.ContractArtifactID ||
		g.Request.Source.ArtifactDigest != m.ContractArtifactDigest ||
		g.Request.V25 != grant.ProjectV25(c.ContractResolution) {
		return errors.New("v26 grant and v25 artifact provenance do not match")
	}
	if m.Repository == "" || !validSHA(m.SubjectSHA) || m.GrantRunID <= 0 || m.GrantAttempt != 1 ||
		m.GrantArtifactID <= 0 || !validDigest(m.GrantArtifactDigest) || m.ContractRunID <= 0 ||
		m.ContractAttempt != 1 || m.ContractArtifactID <= 0 || !validDigest(m.ContractArtifactDigest) {
		return errors.New("workflow source metadata is incomplete")
	}
	if c.ContractResolution.SubjectSHA != m.SubjectSHA || g.Request.V25.SubjectSHA != m.SubjectSHA ||
		c.ContractResolution.ExecutionInput.SubjectSHA != m.SubjectSHA {
		return errors.New("v25 contract subject does not match the v26 workflow subject")
	}
	policy, err := grant.CompilePolicy(os.DirFS("."), grant.PolicyPath)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(policy.Evidence, g.Policy) {
		return errors.New("v26 grant policy evidence does not match the checked-out subject")
	}
	requestInput := grant.GrantInput{Request: g.Request, Live: true}
	if err := grant.VerifyGrantResolution(g.Resolution); err != nil {
		return err
	}
	verification := grant.Verify(policy, requestInput, g.Resolution)
	if !verification.Verified || verification != g.Verification ||
		!reflect.DeepEqual(g.Metrics, g.Resolution.Metrics) ||
		!g.Resolution.GrantAllowsExecution ||
		g.Resolution.Receipt == nil || g.Resolution.Receipt.RemainingUses != 1 ||
		g.Resolution.Receipt.ConsumedUses != 0 || g.Resolution.ExecutionCount != 0 ||
		g.Resolution.OneUseEnforced || g.Metrics.LiveGrants != 1 ||
		g.Request.V25.MaxExecutions != input.MaxExecutions || g.Request.V25.RepositoryWritesAllowed {
		return errors.New("v26 grant is invalid, already consumed, or outside its declared boundary")
	}
	return nil
}

func newReservation(g grant.LiveReport, _ contract.LiveReport, m metadata) reservation {
	value := reservation{Schema: reportSchema + "/reservation", GrantID: g.Resolution.Receipt.GrantID,
		RequestDigest: g.Request.Digest, SubjectSHA: m.SubjectSHA, SourceRunID: m.GrantRunID,
		SourceAttempt: m.GrantAttempt, Status: "CONSUMPTION_RESERVED", ConsumedUses: 1,
		OneUseEnforced: true,
		ReservedAt:     time.Now().UTC().Format(time.RFC3339Nano)}
	value.Digest = digestJSON(value)
	return value
}
