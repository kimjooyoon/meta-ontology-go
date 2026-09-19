package selfimprovementcontinuation

import "fmt"

func BuildRequest(program PolicyProgram, input ContinuationInput) ContinuationRequest {
	request := ContinuationRequest{
		Schema: RequestSchema, Lifecycle: "REQUESTED", ContractID: ContractID, Input: input,
		Decision: Decision("REQUESTED"), Resolution: Resolution("UNRESOLVED"), Reason: "CI_CONTINUATION_REQUESTED",
		ExecutionAuthorized: false, RepositoryWrites: 0, LocalTestExecutions: 0,
	}
	request.Digest = requestDigest(request)
	return request
}

func BuildReport(program PolicyProgram, input ContinuationInput) Report {
	request := BuildRequest(program, input)
	resolution := Evaluate(program, input)
	resolution.RequestDigest = request.Digest
	resolution.Digest = resolutionDigest(resolution)
	report := Report{Schema: Schema, Policy: program.Evidence, Request: request, Resolution: resolution, Metrics: resolution.Metrics,
		Verification: Verify(program, request, resolution)}
	report.Digest = reportDigest(report)
	return report
}

func ValidateReport(report Report) error {
	if report.Schema != Schema || report.Policy.Schema != PolicySchema || report.Request.Schema != RequestSchema || report.Resolution.Schema != ResolutionSchema || report.Verification.Schema != VerificationSchema {
		return fmt.Errorf("continuation report schema mismatch")
	}
	if report.Request.Digest != requestDigest(report.Request) || report.Resolution.Digest != resolutionDigest(report.Resolution) || report.Verification.Digest != verificationDigest(report.Verification) || report.Digest != reportDigest(report) {
		return fmt.Errorf("continuation report digest mismatch")
	}
	if report.Resolution.RequestDigest != report.Request.Digest || report.Verification.RequestDigest != report.Request.Digest || report.Verification.ResolutionDigest != report.Resolution.Digest {
		return fmt.Errorf("continuation report references mismatch")
	}
	if err := ValidateResolution(report.Resolution); err != nil {
		return err
	}
	if !report.Verification.Verified || report.Verification.IndependentReplayComparisons != 1 {
		return fmt.Errorf("continuation report independent verification failed")
	}
	return nil
}
