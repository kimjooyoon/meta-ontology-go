package valueexecution

import (
	"bytes"
	"fmt"
)

func valueFixture(activity string) []byte {
	return []byte("package valuewitness\nnamespace valuewitness\n\n" +
		"entity Integer id \"gooo://value-witness/entity/integer\"\n\n" + activity + "\n")
}

func compileReason(filename string, source []byte, activity string) string {
	_, err := Compile(filename, source, activity)
	return Reason(err)
}

func countLines(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	return bytes.Count(source, []byte{'\n'}) + boolInt(source[len(source)-1] != '\n')
}

func coordinate(satisfied, total int) Coordinate {
	basisPoints := 0
	if total > 0 {
		basisPoints = satisfied * 10_000 / total
	}
	return Coordinate{Satisfied: satisfied, Total: total, BasisPoints: basisPoints}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func finalize(report Report) Report {
	report.Digest = reportDigest(report)
	return report
}

func requireExactCount(label string, got, want int) error {
	if got != want {
		return fmt.Errorf("%s=%d want=%d", label, got, want)
	}
	return nil
}
