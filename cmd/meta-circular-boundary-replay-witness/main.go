package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
)

func main() {
	aPath := flag.String("a", "", "first producer report")
	bPath := flag.String("b", "", "second producer report")
	output := flag.String("output", "", "raw replay evidence output path")
	flag.Parse()
	if *aPath == "" || *bPath == "" || *output == "" {
		fatal(fmt.Errorf("--a, --b, and --output are required"))
	}
	a, err := readReport(*aPath)
	if err != nil {
		fatal(err)
	}
	b, err := readReport(*bPath)
	if err != nil {
		fatal(err)
	}
	if len(a.Receipts) != len(b.Receipts) || len(a.Receipts) != 4 || len(a.Cases) != len(b.Cases) {
		fatal(fmt.Errorf("replay reports have different shapes"))
	}
	evidence := contract.ReplayEvidence{Schema: "gooo/meta-circular-boundary-replay/v1", Producer: "ci-replay-observer", Equal: true}
	for index := range a.Receipts {
		evidence.ReceiptDigestsA = append(evidence.ReceiptDigestsA, a.Receipts[index].ReceiptDigest)
		evidence.ReceiptDigestsB = append(evidence.ReceiptDigestsB, b.Receipts[index].ReceiptDigest)
		evidence.ExecutionDigestsA = append(evidence.ExecutionDigestsA, executionDigest(a, index))
		evidence.ExecutionDigestsB = append(evidence.ExecutionDigestsB, executionDigest(b, index))
		if evidence.ReceiptDigestsA[index] == "" || evidence.ReceiptDigestsA[index] != evidence.ReceiptDigestsB[index] || evidence.ExecutionDigestsA[index] != evidence.ExecutionDigestsB[index] {
			evidence.Equal = false
		}
	}
	encoded, err := json.Marshal(evidence)
	if err != nil {
		fatal(err)
	}
	evidence.EvidenceDigest = digestBytes(encoded)
	encoded, err = json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		fatal(err)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("replay evidence: equal=%t cases=%d\n", evidence.Equal, len(evidence.ReceiptDigestsA))
}

func readReport(path string) (contract.Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return contract.Report{}, err
	}
	var report contract.Report
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return contract.Report{}, err
	}
	return report, nil
}

func executionDigest(report contract.Report, index int) string {
	if index >= len(report.Cases) || report.Cases[index].Receipt.ExecutionArtifact == nil {
		return ""
	}
	return report.Cases[index].Receipt.ExecutionArtifact.OutputDigest
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
