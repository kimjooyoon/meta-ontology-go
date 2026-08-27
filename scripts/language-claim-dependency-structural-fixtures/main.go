package main

// This fixture tool reseals deliberately malformed structural inventories so
// producer and independent judge must report the typed inventory failure, not
// merely reject a stale top-level digest.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependency"
)

func main() {
	evidencePath := flag.String("evidence", "", "valid evidence receipt containing a structural inventory")
	caseName := flag.String("case", "", "missing, duplicate, additional, replacement, contract-substitution, or edge-predicate")
	replacementContractPath := flag.String("replacement-contract", "", "replacement external validator contract for contract-substitution")
	replacementManifestPath := flag.String("replacement-manifest", "", "replacement external structural manifest for contract-substitution")
	outputPath := flag.String("output", "", "resealed malformed evidence receipt")
	flag.Parse()
	if *evidencePath == "" || *caseName == "" || *outputPath == "" {
		fail("-evidence, -case, and -output are required")
	}
	data, err := os.ReadFile(*evidencePath)
	if err != nil {
		fail(err.Error())
	}
	var evidence claimdependency.EvidenceReceipt
	if err := strictJSON(data, &evidence); err != nil {
		fail(err.Error())
	}
	if len(evidence.StructuralContradictions) != 2 || len(evidence.ObservationBundleRaw) == 0 {
		fail("fixture requires the refuted two-row structural inventory")
	}
	var bundle claimdependency.ObservationBundle
	if err := strictJSON(evidence.ObservationBundleRaw, &bundle); err != nil {
		fail(err.Error())
	}
	if *caseName == "edge-predicate" {
		mutated := false
		for i := range bundle.Observations {
			if bundle.Observations[i].Binding == "EDGE" {
				bundle.Observations[i].ObservedPredicate = claimdependency.ObservationUnknown
				bundle.Observations[i].ObservedValue = "edge-predicate-mutation"
				bundle.Observations[i].ComparisonResult = "MISMATCH"
				bundle.Observations[i].Digest = observationDigest(bundle.Observations[i])
				mutated = true
				break
			}
		}
		if !mutated {
			fail("fixture requires an edge observation")
		}
		evidence.Observations = append([]claimdependency.ObservationReceipt(nil), bundle.Observations...)
	} else if *caseName == "contract-substitution" {
		if *replacementContractPath == "" || *replacementManifestPath == "" {
			fail("contract-substitution requires -replacement-contract and -replacement-manifest")
		}
		contractBytes, err := os.ReadFile(*replacementContractPath)
		if err != nil {
			fail(err.Error())
		}
		manifestBytes, err := os.ReadFile(*replacementManifestPath)
		if err != nil {
			fail(err.Error())
		}
		bundle.ContractPath = *replacementContractPath
		bundle.ContractDigest = bytesDigest(contractBytes)
		bundle.ContractRaw = contractBytes
		bundle.StructuralManifestPath = *replacementManifestPath
		bundle.StructuralManifestDigest = bytesDigest(manifestBytes)
		bundle.StructuralManifestRaw = manifestBytes
		evidence.ValidatorContractPath = *replacementContractPath
		evidence.ValidatorContractDigest = bytesDigest(contractBytes)
		evidence.ValidatorContractRaw = contractBytes
		evidence.StructuralManifestPath = *replacementManifestPath
		evidence.StructuralManifestDigest = bytesDigest(manifestBytes)
		evidence.StructuralManifestRaw = manifestBytes
	} else {
		bundle.StructuralContradictions = mutate(bundle.StructuralContradictions, *caseName)
	}
	for i := range bundle.StructuralContradictions {
		bundle.StructuralContradictions[i].Digest = structuralDigest(bundle.StructuralContradictions[i])
	}
	bundle.Digest = bundleDigest(bundle)
	bundleBytes, err := json.Marshal(bundle)
	if err != nil {
		fail(err.Error())
	}
	evidence.StructuralContradictions = append([]claimdependency.StructuralContradiction(nil), bundle.StructuralContradictions...)
	evidence.ObservationBundleRaw = bundleBytes
	evidence.ObservationBundleDigest = bundle.Digest
	evidence.ObservationBundleRawDigest = bytesDigest(bundleBytes)
	if *caseName != "edge-predicate" {
		evidence.SemanticEvidenceDigest = claimdependency.SemanticEvidenceDigestForObservation(evidence)
	}
	evidence.Digest = ""
	evidence.Digest = evidenceDigest(evidence)
	writeJSON(*outputPath, evidence)
}

func structuralDigest(value claimdependency.StructuralContradiction) string {
	value.Digest = ""
	data, err := json.Marshal(value)
	if err != nil {
		fail(err.Error())
	}
	return bytesDigest(data)
}

func observationDigest(value claimdependency.ObservationReceipt) string {
	value.Digest = ""
	data, err := json.Marshal(value)
	if err != nil {
		fail(err.Error())
	}
	return bytesDigest(data)
}

func mutate(rows []claimdependency.StructuralContradiction, name string) []claimdependency.StructuralContradiction {
	switch name {
	case "missing":
		return append([]claimdependency.StructuralContradiction(nil), rows[:1]...)
	case "duplicate":
		return append(append([]claimdependency.StructuralContradiction(nil), rows...), rows[0])
	case "additional":
		extra := rows[0]
		extra.ClaimID = "claimdependency://activity/extra"
		return append(append([]claimdependency.StructuralContradiction(nil), rows...), extra)
	case "replacement":
		replaced := append([]claimdependency.StructuralContradiction(nil), rows...)
		replaced[0].ClaimID = "claimdependency://activity/replacement"
		return replaced
	default:
		fail("unknown structural fixture case: " + name)
		return nil
	}
}

func bundleDigest(bundle claimdependency.ObservationBundle) string {
	bundle.Digest = ""
	// Profile is a fixture label, and is intentionally excluded from the
	// semantic bundle digest just as it is in the producer and judge.
	bundle.Profile = ""
	data, err := json.Marshal(bundle)
	if err != nil {
		fail(err.Error())
	}
	return bytesDigest(data)
}

func evidenceDigest(evidence claimdependency.EvidenceReceipt) string {
	evidence.Digest = ""
	data, err := json.Marshal(evidence)
	if err != nil {
		fail(err.Error())
	}
	return bytesDigest(data)
}

func bytesDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func strictJSON(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return fmt.Errorf("trailing JSON data: %w", err)
	}
	return nil
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
