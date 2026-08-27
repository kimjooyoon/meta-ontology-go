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
	source := flag.String("source", "", "raw .gooo source that supplies the claim graph")
	artifact := flag.String("artifact", "", "raw .gooo artifact to observe")
	operation := flag.String("operation", "", "availability, acceptance, or contradiction")
	capability := flag.String("capability", "", "current CI capability evidence")
	repository := flag.String("repo-root", "", "repository root for tracked/untracked snapshot")
	output := flag.String("output", "", "actual output path recorded in the receipt")
	observation := flag.String("observation", "", "raw target observation receipt; omit when no external observation exists")
	contract := flag.String("contract", "", "external validator contract bytes read by the producer")
	manifest := flag.String("structural-manifest", "", "external structural inventory oracle bytes read by the producer")
	flag.Parse()
	if *source == "" || *artifact == "" || *operation == "" || *capability == "" || *repository == "" || *output == "" || *contract == "" || *manifest == "" {
		fail("-source, -artifact, -operation, -capability, -repo-root, -contract, -structural-manifest, and -output are required")
	}
	receipt, err := claimdependency.BuildCurrentEvidenceForSourceWithExternal(*source, *artifact, *operation, *capability, *repository, *output, *observation, *contract, *manifest)
	if err != nil {
		fail(err.Error())
	}
	writeJSON(*output, receipt)
	fmt.Printf("current evidence operation=%s status=%s claims=%d unknown=%d artifact=%s snapshot=%s authority=%s\n", receipt.Operation, receipt.Status, len(receipt.Claims), countUnknown(receipt), receipt.ArtifactBytesDigest, receipt.Snapshot.AfterDigest, receipt.Capability.Permission)
}

func countUnknown(receipt claimdependency.EvidenceReceipt) int {
	total := 0
	for _, claim := range receipt.Claims {
		if claim.ObservedPredicate == claimdependency.ObservationUnknown {
			total++
		}
	}
	return total
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
func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(2) }
