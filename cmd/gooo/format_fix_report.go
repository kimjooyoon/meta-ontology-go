package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/formatfix"
)

type formatCommandReport struct {
	Schema          string   `json:"schema"`
	Command         string   `json:"command"`
	Status          string   `json:"status"`
	File            string   `json:"file"`
	Changed         bool     `json:"changed"`
	Source          string   `json:"source,omitempty"`
	SourceDigest    string   `json:"source_digest"`
	FormattedDigest string   `json:"formatted_digest"`
	Diagnostics     []string `json:"diagnostics"`
	DirectWrites    int      `json:"direct_writes"`
}

func writeFormatJSON(writer io.Writer, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "%s\n", raw)
	return err
}

func reportFormatFailure(plan formatfix.Plan, jsonMode bool, stdout, stderr io.Writer) int {
	if jsonMode {
		if writeFormatJSON(stdout, plan) != nil {
			return exitFailure
		}
		return exitFailure
	}
	fmt.Fprintf(stderr, "gooo: %s: %s\n", plan.File, plan.ReasonCode)
	return exitFailure
}
