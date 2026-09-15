package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/causalci/consumer"
)

func main() {
	inputPath := flag.String("input", "", "raw observation JSON")
	sourcePath := flag.String("source-file", "", "source bytes to replay")
	logicalPath := flag.String("logical-source", "", "logical source path in the observation")
	receiptPath := flag.String("receipt", "", "producer plan receipt")
	outputPath := flag.String("adjudication", "", "consumer adjudication receipt")
	variantID := flag.String("variant-id", "base", "plan variant identity")
	intervention := flag.Bool("intervention", false, "replay an explicitly supplied intervention source")
	flag.Parse()
	if *inputPath == "" || *sourcePath == "" || *logicalPath == "" || *receiptPath == "" || *outputPath == "" {
		fail(fmt.Errorf("consumer requires input, source-file, logical-source, receipt, and adjudication"))
	}
	input, err := os.ReadFile(*inputPath)
	if err != nil {
		fail(err)
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err)
	}
	receiptRaw, err := os.ReadFile(*receiptPath)
	if err != nil {
		fail(err)
	}
	var receipt struct {
		Digest string `json:"digest"`
		Source struct {
			BindingMode string `json:"binding_mode"`
		} `json:"source"`
	}
	if err := json.Unmarshal(receiptRaw, &receipt); err != nil {
		fail(err)
	}

	result := "FAIL"
	exitCode := 1
	coordinate := consumer.Coordinate{Stage: "ADJUDICATION", Step: "consumer-replay", Reason: "INDEPENDENT_RECONSTRUCTION_REJECTED"}
	verifyErr := error(nil)
	if *intervention {
		verifyErr = consumer.VerifyIntervention(input, *logicalPath, source, receiptRaw)
	} else {
		verifyErr = consumer.Verify(input, *logicalPath, source, receiptRaw)
	}
	if verifyErr == nil {
		result = "PASS"
		exitCode = 0
		coordinate.Reason = "INDEPENDENT_RECONSTRUCTION_OBSERVED"
	}
	outputMaterial := []byte(result)
	if verifyErr != nil {
		outputMaterial = []byte(result + ":" + verifyErr.Error())
	}
	value := consumer.ConsumerAdjudication{Schema: "gooo/causal-ci-selection-process-observation/v1", VariantID: *variantID, LogicalSourcePath: *logicalPath, BindingMode: receipt.Source.BindingMode, PlanReceiptDigest: receipt.Digest, InputDigest: bytesDigest(input), SourceBytesDigest: bytesDigest(source), ResultDigest: bytesDigest(outputMaterial), ConsumerIdentity: "gooo://consumer/causal-ci-selection/process", ExitCode: exitCode, Result: result, Coordinate: coordinate}
	value.Digest = digestWithoutDigest(value)
	if err := writeJSON(*outputPath, value); err != nil {
		fail(err)
	}
	if verifyErr != nil {
		fmt.Fprintln(os.Stderr, verifyErr)
	}
	os.Exit(exitCode)
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepathDir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func filepathDir(path string) string {
	for index := len(path) - 1; index >= 0; index-- {
		if path[index] == '/' {
			if index == 0 {
				return "/"
			}
			return path[:index]
		}
	}
	return "."
}

func digestWithoutDigest(value consumer.ConsumerAdjudication) string {
	value.Digest = ""
	raw, _ := json.Marshal(value)
	return bytesDigest(raw)
}
func bytesDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
