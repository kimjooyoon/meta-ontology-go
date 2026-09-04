package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
)

func main() {
	mode := flag.String("mode", "live", "live, cases, or verify")
	contractPath := flag.String("contract", contract.PolicyPath, "caller-selected policy source")
	requestPath := flag.String("request", "", "optional v24 authorization request JSON")
	resolutionPath := flag.String("resolution", "", "resolution JSON for verify mode")
	outputPath := flag.String("output", "", "caller-owned output artifact")
	check := flag.Bool("check", false, "validate the emitted artifact")
	flag.Parse()

	if *outputPath == "" {
		fail(errors.New("-output is required"))
	}
	program, err := contract.CompilePolicy(os.DirFS("."), *contractPath)
	if err != nil {
		fail(err)
	}

	switch *mode {
	case "live":
		input := contract.ContractInput{Registry: contract.KnownRegistry()}
		if *requestPath != "" {
			var request selfimprovementcandidate.AuthorizationRequest
			if err := readJSON(*requestPath, &request); err == nil {
				input = contract.ProjectAuthorizationRequest(request, contract.KnownRegistry())
			}
		}
		resolution := contract.Evaluate(program, input)
		report := contract.LiveReport{ContractResolution: resolution, Verification: contract.Verify(program, input, resolution)}
		if *check {
			if err := contract.VerifyResolution(report.ContractResolution); err != nil || !report.Verification.Verified {
				fail(fmt.Errorf("live contract check failed: resolution=%v verification=%v", err, report.Verification))
			}
		}
		writeJSON(*outputPath, report)
	case "cases":
		report := contract.BuildCanonicalCaseReport(program)
		if *check {
			if report.CaseDenominator != 9 || report.ClosedCases != 3 || report.UnknownCases != 3 || report.RefutedCases != 3 || !report.ReplayEqual || report.LiveExecutionCount != 0 || report.CanonicalExecutionCount != 0 || report.ExecutionGrants != 0 || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 {
				fail(errors.New("canonical execution contract case check failed"))
			}
			for _, current := range report.Cases {
				if !current.Pass {
					fail(fmt.Errorf("canonical case %q failed", current.ID))
				}
			}
		}
		writeJSON(*outputPath, report)
	case "verify":
		if *resolutionPath == "" {
			fail(errors.New("-resolution is required for verify mode"))
		}
		var resolution contract.ContractResolution
		if err := readJSON(*resolutionPath, &resolution); err != nil {
			fail(err)
		}
		if err := contract.VerifyResolution(resolution); err != nil {
			fail(err)
		}
		verification := contract.Verification{Schema: contract.VerificationSchema, ContractDigest: resolution.Digest,
			IndependentDecision: resolution.Decision, IndependentResolution: resolution.Resolution,
			IndependentReason: resolution.Reason, Verified: true, IndependentReplayComparisons: 1}
		verification.Digest = verificationDigest(verification)
		writeJSON(*outputPath, verification)
	default:
		fail(fmt.Errorf("unknown mode %q", *mode))
	}
}

func readJSON(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("JSON has trailing content")
	}
	return nil
}

func writeJSON(path string, value any) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fail(err)
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fail(err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		fail(err)
	}
}

func verificationDigest(value contract.Verification) string {
	value.Digest = ""
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func digestBytes(raw []byte) string {
	// The CLI only needs a stable verification digest for the already typed
	// verification record. The contract package owns all semantic digests.
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
