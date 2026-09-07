package main

import (
	"errors"
	"strings"
	"testing"
)

func TestMarkdownReportsIntervalsAndLimits(t *testing.T) {
	report, err := readCostReport(strings.NewReader(startEvent + "\n" + returnEvent))
	if err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	if err := writeCostMarkdown(&output, report); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"| Measured intervals | 1 |", "| Unknown returns | 0 |",
		"**UNKNOWN**", "**UNVERIFIED**", "123", "Do not add these rows", "not zero execution cost"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("missing report statement %q", want)
		}
	}
}

func TestMarkdownEscapesInputAndPropagatesWriteFailure(t *testing.T) {
	got := costMarkdownCell("[link](x)|<script>\n*bold*")
	if strings.ContainsAny(got, "[]|<>\n*") {
		t.Fatalf("unescaped input: %s", got)
	}
	if err := writeCostMarkdown(costFailWriter{}, costReport{}); err == nil {
		t.Fatal("write failure was hidden")
	}
}

type costFailWriter struct{}

func (costFailWriter) Write([]byte) (int, error) {
	return 0, errors.New("output unavailable")
}
