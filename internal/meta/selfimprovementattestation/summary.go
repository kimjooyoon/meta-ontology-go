package selfimprovementattestation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Summary(receipt ResolutionReceipt) string {
	var output strings.Builder
	fmt.Fprintln(&output, "## Gooo EHT-8 producer attestation")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "| Reader | Resolution | Verified | Total | Coverage (bps) |")
	fmt.Fprintln(&output, "| --- | --- | ---: | ---: | ---: |")
	for _, view := range receipt.Views {
		fmt.Fprintf(&output, "| %s | %s | %d | %d | %d |\n", view.Audience,
			view.Resolution, view.VerifiedTotal, view.FixedTotal, view.CoverageBasisPoints)
	}
	fmt.Fprintln(&output)
	fmt.Fprintf(&output, "- Decision: `%s / %s / %s`\n", receipt.Decision, receipt.Resolution, receipt.Reason)
	fmt.Fprintf(&output, "- D logical subject: `%s`\n", visible(receipt.SubjectSHA))
	fmt.Fprintf(&output, "- R workflow source: `%s`\n", visible(receipt.ProducerIdentity.WorkflowSHA))
	fmt.Fprintf(&output, "- T archive digest: `%s`\n", visible(receipt.SourceArchiveDigest))
	fmt.Fprintf(&output, "- Open / unknown / false: `%d / %d / %d`\n",
		receipt.Metrics.OpenTotal, receipt.Metrics.UnknownTotal, receipt.Metrics.FalseTotal)
	fmt.Fprintf(&output, "- False promotion: `%d`\n", receipt.Metrics.FalsePromotionCount)
	fmt.Fprintf(&output, "- Claim: `%s`\n", claimSummary(receipt))
	fmt.Fprintf(&output, "- Checker: `%s`, `%s`, exit `%d`, matches `%d`\n", receipt.Checker.Name,
		receipt.Checker.Version, receipt.Checker.ExitCode, receipt.Checker.VerifiedResultTotal)
	fmt.Fprintf(&output, "- Repository writes: `%d`\n", receipt.Authority.RepositoryWrites)
	fmt.Fprintf(&output, "- Mutation / execution / promotion / adoption: `%t / %t / %t / %t`\n",
		receipt.Authority.MutationAuthorized, receipt.Authority.ExecutionAuthorized,
		receipt.Authority.PromotionAuthorized, receipt.Authority.AutomaticAdoptionAuthorized)
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "### Explicit non-claims")
	for _, claim := range receipt.NotClaimed {
		fmt.Fprintf(&output, "- `%s`\n", claim)
	}
	fmt.Fprintf(&output, "\nReceipt: `%s`\n", receipt.Digest)
	return output.String()
}

func visible(value string) string {
	if value == "" {
		return "UNKNOWN"
	}
	return value
}

func claimSummary(receipt ResolutionReceipt) string {
	if len(receipt.ClaimTransitions) == 0 {
		return attestationID + " OPEN"
	}
	transition := receipt.ClaimTransitions[0]
	return transition.ClaimID + " " + transition.Before + " -> " + transition.After
}

func WriteSummary(filename string, receipt ResolutionReceipt) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(Summary(receipt)), 0o644)
}
