package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func writeBodyFillCandidateOrigins(file string, source []byte, flags map[string]string, stdout, stderr io.Writer) int {
	derived, err := bodycodegen.IsSourceFillDerivation(context.Background(), file, source, flags["--activity"])
	if err == nil && derived {
		return writeBodyDerivedCandidateOrigins(file, source, flags, stdout, stderr)
	}
	r, err := bodycodegen.InspectSourceFillCandidateOrigins(file, source, flags["--activity"], flags["--fill-candidate"])
	if err == nil && flags["--inputs-for"] != "" {
		var raw []byte
		raw, err = bodycodegen.SourceFillCandidateOriginInputs(r, flags["--inputs-for"])
		if err == nil {
			_, err = fmt.Fprintln(stdout, string(raw))
		}
	} else if err == nil {
		err = json.NewEncoder(stdout).Encode(r)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	if r.Status != "RESOLVED" {
		return exitFailure
	}
	return exitOK
}
