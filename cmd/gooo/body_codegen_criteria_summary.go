package main

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func writeBodyCodegenCriteriaSummary(output io.Writer, result bodycodegen.Result) {
	r := result.Report.RecordAssembly
	if r == nil || r.Criteria == nil {
		return
	}
	fmt.Fprintf(output, "gooo: %s: %s; selection conditions %d/%d passed, %d failed, %d unobserved",
		result.Report.Activity, r.Status, r.Criteria.Passed, r.Criteria.Total, r.Criteria.Failed, r.Criteria.Unknown)
	if held := r.CriteriaHoldout; held != nil {
		fmt.Fprintf(output, "; separate conditions %d/%d passed, %d failed, %d unobserved",
			held.Passed, held.Total, held.Failed, held.Unknown)
	}
	fmt.Fprintf(output, "; attempts %d", len(r.Attempts))
	if r.AttemptBudget != nil {
		fmt.Fprintf(output, "/%d", *r.AttemptBudget)
	}
	fmt.Fprintf(output, "; model calls %d\n", r.ModelCalls)
}
