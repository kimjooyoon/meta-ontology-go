package main

import (
	"bytes"
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
	if err := validateBranchProtection(context.BranchProtection, evidence, context); err != nil {
		return proofInputs{}, err
	}
	if err := validateDomainEvidence(context.DomainEvidence, evidence, context); err != nil {
		return proofInputs{}, err
	}
	return proofInputs{Governance: governanceInput{
		Schema:           matrix.Schema,
		RequiredContexts: governanceContexts{Dev: matrix.RequiredContexts.Dev, Main: matrix.RequiredContexts.Main},
		GuardianContexts: guardianContexts{DevShadow: matrix.GuardianContexts.DevShadow, MainRequired: matrix.GuardianContexts.MainRequired},
		ProofJobs:        matrix.ProofJobs,
		Promotion:        promotionInput{Source: matrix.Promotion.Source, Target: matrix.Promotion.Target, RequiredChecks: matrix.Promotion.RequiredChecks, BranchProtectionRequired: matrix.Promotion.BranchProtectionRequired},
	}, Evidence: evidence, Jobs: jobs, Context: context}, nil
}

func readJobs(filename string) ([]jobInput, error) {
	jobs, err := readJSON[[]jobInput](filename)
	if err != nil {
		return nil, fmt.Errorf("read workflow jobs: %w", err)
	}
	byName := make(map[string]jobInput, len(proofJobs))
	seenIDs := make(map[int64]bool, len(proofJobs))
	for _, job := range jobs {
		if !isProofJob(job.Name) {
			continue
		}
		if _, exists := byName[job.Name]; exists {
			return nil, fmt.Errorf("duplicate canonical proof job %q", job.Name)
		}
		if job.ID <= 0 || seenIDs[job.ID] {
			return nil, fmt.Errorf("duplicate or invalid canonical proof job id %d", job.ID)
		}
		seenIDs[job.ID] = true
		byName[job.Name] = job
	}
	result := make([]jobInput, 0, len(proofJobs))
	for _, name := range proofJobs {
		job, ok := byName[name]
		if !ok || job.ID <= 0 || job.Status != "completed" || job.Conclusion != "success" || !validSHA(job.HeadSHA) || job.RunID <= 0 || job.RunAttempt <= 0 {
			return nil, fmt.Errorf("canonical proof job %q is missing or unsuccessful", name)
		}
		result = append(result, job)
	}
	return result, nil
}

func validateInputIdentity(evidence evidenceInput, context contextInput, jobs []jobInput) error {
	if evidence.Schema != evidenceSchema || evidence.Repository == "" || evidence.Event == "" || !validEventRef(evidence.Event, evidence.EventRef) || evidence.CheckoutRef != evidence.HeadSHA || evidence.BaseRef == "" || evidence.RunID <= 0 || evidence.Attempt <= 0 {
		return fmt.Errorf("CI evidence identity is incomplete")
	}
	if context.Repository != evidence.Repository || context.Event != evidence.Event || context.Ref != evidence.EventRef || context.EventRef != evidence.EventRef || context.CheckoutRef != evidence.CheckoutRef || context.BaseRef != evidence.BaseRef || context.BaseSHA != evidence.BaseSHA || context.HeadSHA != evidence.HeadSHA || context.WorkflowSHA != evidence.WorkflowSHA || context.RunID != evidence.RunID || context.RunAttempt != evidence.Attempt {
		return fmt.Errorf("proof context does not match CI evidence identity")
	}
	if context.Ref == "" || context.EventRef == "" || context.CheckoutRef == "" || context.EventRef != context.Ref || context.CheckoutRef != evidence.HeadSHA || context.Actor == "" || context.Builder == "" || context.Gate == "" || !validSHA(evidence.BaseSHA) || !validSHA(evidence.HeadSHA) || !validSHA(evidence.WorkflowSHA) || !validSHA(context.CheckoutRef) || evidence.BaseSHA == evidence.HeadSHA {
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
	if err := compareJobs(evidence.Jobs, jobs, evidence.HeadSHA, evidence.RunID, evidence.Attempt); err != nil {
		return err
	}
	if err := validateArtifacts(context.Artifacts, context.RunID, context.RunAttempt); err != nil {
		return err
	}
	if err := validateMissingReasons(context.MissingReasons, context.DomainEvidence.ProtectionStatus, context.DomainEvidence.ProvenanceStatus); err != nil {
		return err
	}
	return nil
}

func compareJobs(expected, actual []jobInput, head string, runID, runAttempt int64) error {
	if len(expected) != len(proofJobs) || len(actual) != len(proofJobs) {
		return fmt.Errorf("proof must contain exactly six canonical jobs")
	}
	seenIDs := make(map[int64]bool, len(expected))
	for index, name := range proofJobs {
		left, right := expected[index], actual[index]
		if left.Name != name || right.Name != name || left.ID != right.ID || left.ID <= 0 || seenIDs[left.ID] || left.Status != "completed" || right.Status != "completed" || left.Conclusion != "success" || right.Conclusion != "success" || right.HeadSHA != head || left.HeadSHA != head || left.RunID != runID || right.RunID != runID || left.RunAttempt != runAttempt || right.RunAttempt != runAttempt {
			return fmt.Errorf("proof job %q is missing or mismatched", name)
		}
		seenIDs[left.ID] = true
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

func validateBranchProtection(protection branchProtection, evidence evidenceInput, context contextInput) error {
	if protection.Repository != evidence.Repository || protection.Branch != context.BaseRef || protection.PolicySHA != evidence.Digests.Policy || protection.EventRef != context.EventRef || protection.CheckoutRef != context.CheckoutRef || protection.BaseSHA != evidence.BaseSHA || protection.HeadSHA != evidence.HeadSHA || protection.RunID != evidence.RunID || protection.RunAttempt != evidence.Attempt || protection.WorkflowSHA != evidence.WorkflowSHA {
		return fmt.Errorf("branch protection snapshot is missing or unbound")
	}
	if protection.TokenSource != "github.token" && protection.TokenSource != "BRANCH_PROTECTION_TOKEN" || protection.ReadStatus != "verified" && protection.ReadStatus != "unavailable" {
		return fmt.Errorf("branch protection snapshot source or status is invalid")
	}
	if protection.ReadStatus == "unavailable" && protection.MissingReason == "" || protection.ReadStatus == "verified" && protection.MissingReason != "" {
		return fmt.Errorf("branch protection missing reason is inconsistent with read status")
	}
	if protection.Digest != digestBranchProtection(protection) {
		return fmt.Errorf("branch protection snapshot digest mismatch")
	}
	if protection.ReadStatus == "verified" && protection.Exists && !sameStringSet(protection.RequiredChecks, requiredContextsForBase(context.BaseRef)) {
		return fmt.Errorf("branch protection required contexts do not match route %q", context.BaseRef)
	}
	if protection.ReadStatus == "verified" && protection.Exists && !validRequiredCheckBindings(protection.RequiredCheckBindings, requiredContextsForBase(context.BaseRef)) {
		return fmt.Errorf("branch protection required check app bindings do not match route %q", context.BaseRef)
	}
	return nil
}

func branchProtectionReady(protection branchProtection) bool {
	return branchProtectionReadyFor(protection, "dev")
}

func requiredContextsForBase(base string) []string {
	if base == "main" {
		return append(append([]string(nil), proofJobs...), "CI guardian")
	}
	return append([]string(nil), proofJobs...)
}

func branchProtectionReadyFor(protection branchProtection, base string) bool {
	return protection.ReadStatus == "verified" && protection.Exists && protection.Strict && protection.EnforceAdmins && protection.RequiredReviews == 0 && !protection.DismissStaleReviews && !protection.RequireLastPushApproval && protection.LinearHistory && !protection.AllowForcePushes && !protection.AllowDeletions && sameStringSet(protection.RequiredChecks, requiredContextsForBase(base)) && validRequiredCheckBindings(protection.RequiredCheckBindings, requiredContextsForBase(base))
}

func promotionReady(promotion promotionInput, protection branchProtection) bool {
	return promotion.BranchProtectionRequired && branchProtectionReady(protection)
}

func digestBranchProtection(protection branchProtection) string {
	protection.Digest = ""
	data, _ := json.Marshal(protection)
	return digestBytes(data)
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[string]bool, len(left))
	for _, value := range left {
		seen[value] = true
	}
	for _, value := range right {
		if !seen[value] {
			return false
		}
	}
	return true
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

func readStrictJSON[T any](filename string) (T, error) {
	var value T
	data, err := os.ReadFile(filename)
	if err != nil {
		return value, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
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

func validEventRef(event, ref string) bool {
	if ref == "" || strings.ContainsAny(ref, "\r\n") {
		return false
	}
	if event == "pull_request" {
		return strings.HasPrefix(ref, "refs/pull/") && strings.HasSuffix(ref, "/merge")
	}
	if event == "push" {
		return strings.HasPrefix(ref, "refs/heads/")
	}
	return false
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

func validArtifactDigest(value string) bool {
	return strings.HasPrefix(value, "sha256:") && validDigest(strings.TrimPrefix(value, "sha256:"))
}
