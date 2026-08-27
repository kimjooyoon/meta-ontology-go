package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/operationprovenance"
	independent "github.com/kimjooyoon/meta-ontology-go/internal/meta/operationprovenance/verify"
)

func main() {
	mode := flag.String("mode", "build", "build or verify")
	sourcePath := flag.String("source", "", "Gooo source")
	receiptPath := flag.String("receipt", "", "receipt JSON for verify")
	outPath := flag.String("out", "", "output JSON")
	flag.Parse()
	if *sourcePath == "" || *outPath == "" {
		fail("source and out are required")
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail("read source: %v", err)
	}
	var value any
	switch *mode {
	case "build":
		value, err = operationprovenance.Build(source)
	case "verify":
		if *receiptPath == "" {
			fail("receipt is required for verify")
		}
		payload, readErr := os.ReadFile(*receiptPath)
		if readErr != nil {
			fail("read receipt: %v", readErr)
		}
		value, err = independent.Verify(payload, source)
	default:
		fail("unknown mode %q", *mode)
	}
	if err != nil {
		fail("%s: %v", *mode, err)
	}
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fail("encode output: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		fail("create output directory: %v", err)
	}
	if err := os.WriteFile(*outPath, append(payload, '\n'), 0o644); err != nil {
		fail("write output: %v", err)
	}
	fmt.Printf("%s: %s\n", *mode, *outPath)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "meta-operation-provenance: "+format+"\n", args...)
	os.Exit(1)
}
