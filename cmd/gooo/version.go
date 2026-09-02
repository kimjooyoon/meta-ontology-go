package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const (
	goooVersion   = "0.2.0-dev"
	versionSchema = "gooo-version/v1"
	versionStatus = "development"
	versionUsage  = "usage: gooo version [--json]"
)

type versionInfo struct {
	SchemaVersion string `json:"schema_version"`
	Language      string `json:"language"`
	Version       string `json:"version"`
	Status        string `json:"status"`
	SemanticIR    string `json:"semantic_ir"`
	SemanticCheck string `json:"semantic_check"`
	Graph         string `json:"graph"`
	FixPlan       string `json:"fix_plan"`
}

func runVersion(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, err := fmt.Fprintf(stdout, "gooo %s (%s)\n", goooVersion, versionStatus)
		if err != nil {
			return exitFailure
		}
		return exitOK
	}
	if len(args) != 1 || args[0] != "--json" {
		fmt.Fprintln(stderr, versionUsage)
		return exitUsage
	}
	payload, err := json.Marshal(versionInfo{
		SchemaVersion: versionSchema, Language: "gooo", Version: goooVersion,
		Status: versionStatus, SemanticIR: semantic.CurrentIRVersion, SemanticCheck: semanticCheckSchemaVersion,
		Graph: graphDumpSchemaVersion, FixPlan: fixPlanSchemaVersion,
	})
	if err != nil {
		return exitFailure
	}
	if _, err := stdout.Write(append(payload, '\n')); err != nil {
		return exitFailure
	}
	return exitOK
}
