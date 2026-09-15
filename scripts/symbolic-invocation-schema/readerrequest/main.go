package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintln(os.Stderr, "usage: readerrequest REQUEST PROJECTION SUBJECT_SHA OUTPUT")
		os.Exit(2)
	}
	request, err := os.ReadFile(os.Args[1])
	if err != nil {
		fail(err)
	}
	projection, err := os.ReadFile(os.Args[2])
	if err != nil {
		fail(err)
	}
	result, err := artifactemit.CompileSymbolicReaderRequest(request, projection, os.Args[3])
	if err != nil {
		fail(err)
	}
	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(os.Args[4], append(payload, '\n'), 0o644); err != nil {
		fail(err)
	}
	fmt.Printf(
		"compiler Gooo reader request: %s %d/%d audience=%s indicators=%d\n",
		result.Decision, result.Coordinates.Satisfied, result.Coordinates.Total,
		result.View.Audience, len(result.View.IndicatorIDs),
	)
	if result.Decision != "PASS" {
		os.Exit(1)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
