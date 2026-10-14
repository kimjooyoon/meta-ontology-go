package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func writeBodyFillOriginCandidates(file string, source []byte, activity string, stdout, stderr io.Writer) int {
	r, err := bodycodegen.InspectSourceFillOriginCandidates(context.Background(), file, source, activity)
	if err == nil {
		err = json.NewEncoder(stdout).Encode(r)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	return exitOK
}

func writeBodyDerivedCandidateOrigins(file string, source []byte, flags map[string]string, stdout, stderr io.Writer) int {
	r, err := bodycodegen.InspectSourceFillDerivedOrigins(context.Background(), file, source, flags["--activity"], flags["--fill-candidate"])
	if err == nil && flags["--inputs-for"] != "" {
		var raw []byte
		raw, err = bodycodegen.SourceFillDerivedOriginInputs(r, flags["--inputs-for"])
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
