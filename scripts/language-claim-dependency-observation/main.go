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
	expected := flag.String("expected-bytes-digest", "", "digest expected by the target operation")
	profile := flag.String("profile", "", "accepted, contradiction, contradiction-single, contradiction-no-failure, or unrelated")
	output := flag.String("output", "", "observation receipt output")
	flag.Parse()
	if *source == "" || *artifact == "" || *expected == "" || *profile == "" || *output == "" {
		fail("-source, -artifact, -expected-bytes-digest, -profile, and -output are required")
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
	comparison := "MISMATCH"
	if actual == *expected {
		comparison = "MATCH"
	}
	procedureOutput := fmt.Sprintf("read target path=%s bytes=%d sha256=%s expected=%s comparison_result=%s", *artifact, len(artifactBytes), actual, *expected, comparison)
	bundle, err := claimdependency.BuildObservationBundle(*source, sourceBytes, *artifact, *expected, procedureOutput, *profile)
	if err != nil {
		fail(err.Error())
	}
	writeJSON(*output, bundle)
	fmt.Printf("target_observation profile=%s observations=%d comparison_result=%s target_bytes_digest=%s bundle_digest=%s\n", bundle.Profile, len(bundle.Observations), comparison, bundle.ArtifactBytesDigest, bundle.Digest)
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
