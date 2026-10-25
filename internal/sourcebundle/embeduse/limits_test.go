package embeduse

import (
	"io"
	"strings"
	"testing"
)

func TestToolOutputBudgetAppliesToReaderCopy(t *testing.T) {
	output := &boundedOutput{limit: 4}
	reader := io.LimitReader(strings.NewReader("too much output"), 15)
	if _, err := io.Copy(output, reader); err == nil || len(output.Bytes()) > 4 {
		t.Fatal("io.Copy bypassed the output budget through a promoted ReaderFrom")
	}
}
