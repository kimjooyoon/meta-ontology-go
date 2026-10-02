package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/completenessdelta"
)

const completenessDeltaUsage = "usage: gooo completeness-delta --before <receipt-or-producer.json> --after <receipt-or-producer.json>"

func runCompletenessDelta(args []string, stdout, stderr io.Writer) int {
	flags := map[string]string{"--before": "", "--after": ""}
	for i := 0; i < len(args); i += 2 {
		v, ok := flags[args[i]]
		if !ok || v != "" || i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "--") {
			fmt.Fprintln(stderr, completenessDeltaUsage)
			return exitUsage
		}
		flags[args[i]] = args[i+1]
	}
	if flags["--before"] == "" || flags["--after"] == "" {
		fmt.Fprintln(stderr, completenessDeltaUsage)
		return exitUsage
	}
	before, err := readBodyExecutionFile(flags["--before"], completenessdelta.MaxInputBytes)
	if err != nil {
		fmt.Fprintln(stderr, "gooo completeness-delta: before:", err)
		return exitFailure
	}
	after, err := readBodyExecutionFile(flags["--after"], completenessdelta.MaxInputBytes)
	if err != nil {
		fmt.Fprintln(stderr, "gooo completeness-delta: after:", err)
		return exitFailure
	}
	r, err := completenessdelta.Compare(before, after)
	if err != nil {
		fmt.Fprintln(stderr, "gooo completeness-delta:", err)
		return exitFailure
	}
	if err := json.NewEncoder(stdout).Encode(r); err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	return exitOK
}
