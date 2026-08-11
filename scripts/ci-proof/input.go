package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

func readInputs(root, governancePath, evidencePath, jobsPath, contextPath string) (proofInputs, error) {
	matrix, err := verify.ReadGovernanceMatrix(governancePath)
	if err != nil {
		return proofInputs{}, err
	}
	evidence, err := readJSON[evidenceInput](evidencePath)
	if err != nil {
		return proofInputs{}, fmt.Errorf("read CI evidence: %w", err)
	}
	jobs, err := readJobs(jobsPath)
	if err != nil {
		return proofInputs{}, err
	}
	context, err := readJSON[contextInput](contextPath)
	if err != nil {
		return proofInputs{}, fmt.Errorf("read proof context: %w", err)
	}
	if err := validateInputIdentity(evidence, context, jobs); err != nil {
		return proofInputs{}, err
	}
	if err := validateEvidenceDigests(root, evidence); err != nil {
		return proofInputs{}, err
	}
	return proofInputs{Governance: governanceInput{Schema: matrix.Schema, Promotion: promotionInput{Source: matrix.Promotion.Source, Target: matrix.Promotion.Target, RequiredChecks: matrix.Promotion.RequiredChecks, BranchProtectionRequired: matrix.Promotion.BranchProtectionRequired}}, Evidence: evidence, Jobs: jobs, Context: context}, nil
}

func readJobs(filename string) ([]jobInput, error) {
	jobs, err := readJSON[[]jobInput](filename)
	if err != nil {
		return nil, fmt.Errorf("read workflow jobs: %w", err)
	}
	byName := make(map[string]jobInput, len(proofJobs))
	for _, job := range jobs {
		if !isProofJob(job.Name) {
			continue
		}
		if _, exists := byName[job.Name]; exists {
			return nil, fmt.Errorf("duplicate canonical proof job %q", job.Name)
		}
		if job.Name == "CI policy" && job.Conclusion == "" {
			job.Conclusion = "success"
		}
		byName[job.Name] = job
	}
	result := make([]jobInput, 0, len(proofJobs))
	for _, name := range proofJobs {
		job, ok := byName[name]
		if !ok || job.ID <= 0 || job.Conclusion != "success" || !validSHA(job.HeadSHA) {
			return nil, fmt.Errorf("canonical proof job %q is missing or unsuccessful", name)
		}
		result = append(result, job)
	}
	return result, nil
}

func validateInputIdentity(evidence evidenceInput, context contextInput, jobs []jobInput) error {
	if evidence.Schema != "gooo/ci-evidence/v1" || evidence.Repository == "" || evidence.Event == "" || evidence.BaseRef == "" || evidence.RunID <= 0 || evidence.Attempt <= 0 {
		return fmt.Errorf("CI evidence identity is incomplete")
	}
	if context.Repository != evidence.Repository || context.Event != evidence.Event || context.BaseRef != evidence.BaseRef || context.BaseSHA != evidence.BaseSHA || context.HeadSHA != evidence.HeadSHA || context.WorkflowSHA != evidence.WorkflowSHA || context.RunID != evidence.RunID || context.RunAttempt != evidence.Attempt {
		return fmt.Errorf("proof context does not match CI evidence identity")
	}
	if context.Ref == "" || context.Actor == "" || context.Builder == "" || context.Gate == "" || !validSHA(evidence.BaseSHA) || !validSHA(evidence.HeadSHA) || !validSHA(evidence.WorkflowSHA) || evidence.BaseSHA == evidence.HeadSHA {
		return fmt.Errorf("proof identity is missing, invalid, or identical")
	}
	if evidence.Event != "pull_request" && evidence.Event != "push" {
		return fmt.Errorf("unsupported proof event %q", evidence.Event)
	}
	if evidence.Event == "pull_request" && context.PRNumber <= 0 {
		return fmt.Errorf("pull-request proof number is required")
	}
	if evidence.Event == "push" && context.PRNumber != 0 {
		return fmt.Errorf("push proof cannot carry a pull request number")
	}
	return compareJobs(evidence.Jobs, jobs, evidence.HeadSHA)
}

func compareJobs(expected, actual []jobInput, head string) error {
	if len(expected) != len(proofJobs) || len(actual) != len(proofJobs) {
		return fmt.Errorf("proof must contain exactly six canonical jobs")
	}
	for index, name := range proofJobs {
		left, right := expected[index], actual[index]
		if left.Name != name || right.Name != name || left.ID != right.ID || right.Conclusion != "success" || right.HeadSHA != head || left.HeadSHA != head {
			return fmt.Errorf("proof job %q is missing or mismatched", name)
		}
	}
	return nil
}

func validateEvidenceDigests(root string, evidence evidenceInput) error {
	if !validDigest(evidence.Digests.Source) || !validDigest(evidence.Digests.IR) || !validDigest(evidence.Digests.Generated) || !validDigest(evidence.Digests.Policy) || !validDigest(evidence.Digests.Toolchain) || !validDigest(evidence.Digests.Bundle) {
		return fmt.Errorf("CI evidence has missing or malformed digests")
	}
	if root == "" {
		return fmt.Errorf("proof repository root is required")
	}
	return nil
}

func readJSON[T any](filename string) (T, error) {
	var value T
	data, err := os.ReadFile(filename)
	if err != nil {
		return value, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return value, fmt.Errorf("empty JSON input %s", filename)
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, err
	}
	return value, nil
}

func isProofJob(name string) bool {
	for _, canonical := range proofJobs {
		if name == canonical {
			return true
		}
	}
	return false
}

func validSHA(value string) bool {
	if len(value) != 40 || value == strings.Repeat("0", 40) {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}

func validDigest(value string) bool {
	if len(value) != 64 || value == strings.Repeat("0", 64) {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
