package semanticdelta

import (
	"strconv"
	"strings"
)

func writeTextRecord(builder *strings.Builder, fields ...string) {
	builder.WriteString(strings.Join(fields, "\t"))
	builder.WriteByte('\n')
}

// EncodeReportText returns a stable tab-separated report intended for logs.
func EncodeReportText(report Report) []byte {
	report.Normalize()
	var builder strings.Builder
	writeTextRecord(&builder, "allowed", strconv.FormatBool(report.Passes()))
	writeTextRecord(&builder, "violations", strconv.Itoa(len(report.Violations)))
	for _, violation := range report.Violations {
		writeTextRecord(&builder, "violation", string(violation.Operation), string(violation.Change),
			violation.ID, violation.Kind, violation.Subject, violation.Predicate,
			violation.Object, violation.Endpoint, strconv.Quote(violation.Reason))
	}
	return []byte(builder.String())
}
