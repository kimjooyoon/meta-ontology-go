package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"
)

func assessOriginalReport(policy publictestreuse.Policy, input publictestreuse.OriginalObservation, report *publictestreuse.Report) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	value, assessment, err := policy.AssessOriginal(ctx, input)
	if err != nil {
		return err
	}
	if assessment.Reuse || assessment.CaseID != publictestreuse.CaseBaselineExecution {
		return errors.New("original Gooo assessment has a different operation")
	}
	return applyAssessment(report, input, value, assessment)
}

func assessReplayReport(policy publictestreuse.Policy, input publictestreuse.ReuseObservation,
	report *publictestreuse.Report) (publictestreuse.Assessment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	value, assessment, err := policy.AssessReplay(ctx, input)
	if err != nil {
		return assessment, err
	}
	if (assessment.Decision == publictestreuse.DecisionClosed) != assessment.Reuse {
		return assessment, errors.New("replay Gooo assessment contradicts its proposed operation")
	}
	return assessment, applyAssessment(report, input, value, assessment)
}

func applyAssessment(report *publictestreuse.Report, input any, value bodycodegen.FixedValue, assessment publictestreuse.Assessment) error {
	raw, err := json.Marshal(input)
	if err != nil {
		return err
	}
	report.CaseID, report.Decision, report.Reason = assessment.CaseID, assessment.Decision, assessment.Reason
	report.PolicyInputs, report.PolicyEvaluation = raw, &value
	report.PolicyDecisionSourceDigest = value.SourceSHA256
	report.PolicyEvaluations = 1
	report.Unknown = nil
	if assessment.Decision == publictestreuse.DecisionUnknown {
		report.Unknown = publictestreuse.Unknown(assessment.Reason, assessment.Step, assessment.NextOperation, assessment.BlockedBy)
	}
	return nil
}

func observeReuse(input executionInput, expected publictestreuse.Binding) (publictestreuse.ReuseObservation, *publictestreuse.Receipt, error) {
	observed := publictestreuse.ReuseObservation{Requested: input.Authorize, InputProfile: expected.PackageInputs.Profile}
	if !input.Authorize {
		return observed, nil, nil
	}
	path, state, err := resolveReceiptPath(input.Receipt)
	if err != nil || state != "" {
		observed.ReceiptState = state
		return observed, nil, err
	}
	receipt, err := publictestreuse.ReadReceipt(path)
	if err != nil {
		observed.ReceiptState = "INVALID"
		if errors.Is(err, publictestreuse.ErrUnobservedPackageInputs) {
			observed.ReceiptState = "LEGACY_INPUTS_UNOBSERVED"
		}
		if errors.Is(err, publictestreuse.ErrUnobservedToolInputs) {
			observed.ReceiptState = "LEGACY_TOOL_UNOBSERVED"
		}
		return observed, nil, nil
	}
	observed.ReceiptState = "VALID"
	observed.PolicyState = "MISMATCH"
	if receipt.Binding.PolicySourceDigest == expected.PolicySourceDigest &&
		receipt.Binding.PolicySemanticDigest == expected.PolicySemanticDigest &&
		receipt.Binding.PolicyEvaluatorDigest == expected.PolicyEvaluatorDigest {
		observed.PolicyState = "MATCH"
	}
	observed.BindingState = "MISMATCH"
	if publictestreuse.VerifyReceipt(receipt, expected) == nil {
		observed.BindingState = "MATCH"
	}
	return observed, &receipt, nil
}

// The source result can restrict reuse; it cannot provide a missing receipt or
// replace the host's final comparison of the currently observed package bytes.
func authorizeObservedReuse(input executionInput, identity identity, receipt *publictestreuse.Receipt) error {
	if !input.Authorize || receipt == nil {
		return errors.New("Gooo policy cannot supply a reuse request or successful receipt")
	}
	observed, err := readExecutionPackageInputs(input.PackageDir, identity.ProgramBytes, identity.TestBytes)
	if err != nil {
		return err
	}
	if observed != identity.Binding.PackageInputs {
		return errors.New("execution package changed after its policy observation")
	}
	command, err := captureGoCommand()
	if err != nil {
		return err
	}
	if command.Inputs != identity.Binding.ToolInputs {
		return errors.New("selected Go command changed after its policy observation")
	}
	return publictestreuse.VerifyReceipt(*receipt, identity.Binding)
}
