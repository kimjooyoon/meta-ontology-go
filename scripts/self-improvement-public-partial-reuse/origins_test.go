package main

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicpartialreuse"
)

func TestReuseReportKeepsPreviousSource(t *testing.T) {
	for _, change := range []string{"none", "same-source", "current-origin", "missing-origin", "duplicate-origin", "wrong-partition", "whole-program-equality"} {
		t.Run(change, func(t *testing.T) {
			prior, current := strings.Repeat("a", 64), strings.Repeat("b", 64)
			value := report{PriorSourceDigest: prior, CurrentSourceDigest: current,
				ReusedOrigins:     []receiptOrigin{{"inventory", strings.Repeat("c", 64), prior}},
				SingleChangeAfter: publicpartialreuse.Metrics{TestUnitsReused: 1},
				Cases:             make([]publicpartialreuse.CaseReport, 2)}
			switch change {
			case "same-source":
				value.CurrentSourceDigest = prior
			case "current-origin":
				value.ReusedOrigins[0].SourceDigest = current
			case "missing-origin":
				value.ReusedOrigins = nil
			case "duplicate-origin":
				value.ReusedOrigins = append(value.ReusedOrigins, value.ReusedOrigins[0])
			case "wrong-partition":
				value.ReusedOrigins[0].Partition = "orders"
			case "whole-program-equality":
				value.GeneratedBytesEqual = true
			}
			if err := verifyReusedOrigins(value); (err == nil) != (change == "none") {
				t.Fatal("incorrect prior-source report acceptance", change, err)
			}
		})
	}
}
