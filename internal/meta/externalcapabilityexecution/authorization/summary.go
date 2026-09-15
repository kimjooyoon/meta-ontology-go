package authorization

import (
	"fmt"
	"strings"
)

func Markdown(receipt Receipt, suite Suite) string {
	var output strings.Builder
	output.WriteString("### External capability authorization\n\n")
	fmt.Fprintf(&output, "- decision: `%s / %s` (`%s`)\n", receipt.Decision,
		receipt.Resolution, receipt.Reason)
	fmt.Fprintf(&output, "- CAB-10: `%d/%d` (`%d` bps); unknown/open/rejected: `%d/%d/%d`\n",
		receipt.Completed, receipt.Total, receipt.BasisPoints, receipt.UnknownIndicators,
		receipt.OpenClaims, receipt.RejectedClaims)
	fmt.Fprintf(&output, "- conformance: `%d/%d`; authorized/unknown/denied cases: `%d/%d/%d`\n\n",
		suite.Passed, suite.Total, suite.AuthorizedCases, suite.UnknownCases, suite.DeniedCases)
	output.WriteString("| Reader | Resolution | Coverage |\n| --- | --- | --- |\n")
	for _, view := range receipt.ReaderViews {
		fmt.Fprintf(&output, "| %s | %s | %d/%d (%d bps) |\n", view.Reader,
			view.Resolution, view.Completed, view.Total, view.BasisPoints)
	}
	output.WriteString("\n| Indicator | Stage | Status |\n| --- | --- | --- |\n")
	for _, indicator := range receipt.Indicators {
		fmt.Fprintf(&output, "| %s | %s | %s |\n", indicator.MetaOperation,
			indicator.Stage, indicator.Status)
	}
	if len(receipt.Unknowns) > 0 {
		output.WriteString("\n#### Unknown provenance\n")
		for _, unknown := range receipt.Unknowns {
			fmt.Fprintf(&output, "- `%s`: `%s`\n", unknown.Stage, unknown.Reason)
		}
	}
	output.WriteString("\n#### Non-claims\n")
	for _, claim := range receipt.NonClaims {
		fmt.Fprintf(&output, "- %s\n", claim)
	}
	return output.String()
}
