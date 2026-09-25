package main

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"
)

func writeDebugReceipt(receipt languagedebug.Receipt, jsonOutput bool, stdout, stderr io.Writer) int {
	if jsonOutput {
		data, err := languagedebug.Marshal(receipt)
		if err != nil {
			fmt.Fprintf(stderr, "gooo debug: %v\n", err)
			return exitUsage
		}
		_, _ = stdout.Write(data)
	} else if receipt.CurrentEvent != nil {
		fmt.Fprintf(stdout, "debug: %s %s event=%s trace=%d remaining=%d\n",
			receipt.Decision, receipt.State, receipt.CurrentEvent.Kind,
			len(receipt.Trace), receipt.RemainingEvents)
	} else {
		fmt.Fprintf(stdout, "debug: %s %s reason=%s\n", receipt.Decision, receipt.State, receipt.Reason)
	}
	if receipt.Decision != languagedebug.DecisionPass {
		return exitFailure
	}
	return exitOK
}
