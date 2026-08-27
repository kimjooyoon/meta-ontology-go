package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependency"
)

func main() {
	source := flag.String("source", "", "raw .gooo source whose graph supplies claim and edge bindings")
	artifact := flag.String("artifact", "", "target artifact to observe")
	contract := flag.String("contract", "", "fixed external validator material; it declares no outcome")
	failureReceipt := flag.String("failure-receipt", "", "optional receipt from an actually non-zero CI process")
	profile := flag.String("profile", "", "fixture label only; it cannot select a predicate or state")
	output := flag.String("output", "", "observation receipt output")
	flag.Parse()
	if *source == "" || *artifact == "" || *contract == "" || *profile == "" || *output == "" {
		fail("-source, -artifact, -contract, -profile, and -output are required")
	}
	sourceBytes, err := os.ReadFile(*source)
	if err != nil {
		fail(err.Error())
	}
	artifactBytes, err := os.ReadFile(*artifact)
	if err != nil {
		fail(err.Error())
	}
	actual := digestBytes(artifactBytes)
	procedureOutput := fmt.Sprintf("read target path=%s bytes=%d actual_sha256=%s procedure=raw-bytes-read", *artifact, len(artifactBytes), actual)
	bundle, err := claimdependency.BuildObservationBundle(*source, sourceBytes, *artifact, procedureOutput, *profile, *contract, *failureReceipt)
	if err != nil {
		fail(err.Error())
	}
	writeJSON(*output, bundle)
	fmt.Printf("target_observation profile=%s observations=%d target_bytes_digest=%s bundle_digest=%s\n", bundle.Profile, len(bundle.Observations), bundle.ArtifactBytesDigest, bundle.Digest)
}

func digestBytes(data []byte) string {
	return claimdependency.DigestBytesForObservation(data)
}

func writeJSON(path string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fail(err.Error())
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
