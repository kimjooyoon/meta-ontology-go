package main

import (
	"fmt"
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
