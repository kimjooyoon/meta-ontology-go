package main

import (
	"fmt"
	"io"
	"time"
)

const (
	provenanceCLISchema       = "gooo/provenance/v1"
	provenancePublishUsage    = "usage: gooo provenance publish [--json] <file.gooo> --store <ledger.jsonl> --evidence <evidence.json>"
	provenanceStatusCommitted = "committed"
	provenanceStatusRejected  = "rejected"
)

type provenancePublishOptions struct {
	source   string
	store    string
	evidence string
}

func runProvenance(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	options, err := parseProvenancePublishArguments(args)
	if err != nil {
		if jsonMode {
			return writeProvenanceFailure(stdout, provenancePublishResponse{
				Schema: provenanceCLISchema, Status: provenanceStatusRejected, Records: []provenanceCLIRecord{},
			}, "cli.usage", err.Error())
		}
		fmt.Fprintln(stderr, err)
		return exitUsage
	}

	deadline := time.Now().Add(2 * commandDeadline)
	response, err := publishProvenance(options, reader, parser, deadline)
	if err != nil {
		code := provenanceErrorCode(err)
		if jsonMode {
			return writeProvenanceFailure(stdout, response, code, err.Error())
		}
		fmt.Fprintf(stderr, "gooo: provenance: %s: %v\n", code, err)
		return exitFailure
	}
	if jsonMode {
		if err := writeProvenanceResponse(stdout, response, deadline); err != nil {
			return exitFailure
		}
		return exitOK
	}
	fmt.Fprintf(stdout, "%s: records=%d store_digest=%s\n", response.Status, len(response.Records), response.StoreDigest)
	return exitOK
}
