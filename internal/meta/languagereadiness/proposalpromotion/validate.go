package proposalpromotion

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func Validate(receipt Receipt, expectedCurrentHead string) error {
	switch {
	case receipt.Schema != Schema:
		return fmt.Errorf("FAIL_CLOSED: proposal promotion schema %q", receipt.Schema)
	case receipt.CurrentHeadSHA != expectedCurrentHead:
		return fmt.Errorf("FAIL_CLOSED: proposal promotion current head mismatch")
	case !validSHA(receipt.CurrentHeadSHA) || !validSHA(receipt.EvidenceHeadSHA):
		return fmt.Errorf("FAIL_CLOSED: proposal promotion sha is invalid")
	case receipt.CurrentHeadSHA == receipt.EvidenceHeadSHA:
		return fmt.Errorf("FAIL_CLOSED: proposal promotion self-authorization")
	case receipt.Repository == "":
		return fmt.Errorf("FAIL_CLOSED: proposal promotion repository is empty")
	case receipt.Source.Selection.RunID <= 0 || receipt.Source.Selection.ArtifactID <= 0:
		return fmt.Errorf("FAIL_CLOSED: proposal promotion source identity is invalid")
	}
	expected := evaluate(receipt.CurrentHeadSHA, receipt.EvidenceHeadSHA, receipt.Source)
	if !reflect.DeepEqual(receipt, expected) {
		return fmt.Errorf("FAIL_CLOSED: proposal promotion receipt mismatch")
	}
	if receipt.Decision != DecisionPass || receipt.Summary.Satisfied != totalCoordinates ||
		receipt.Summary.Unresolved != 0 || receipt.RepositoryWrites != 0 ||
		receipt.RepositoryMutationAuthorized {
		return fmt.Errorf("FAIL_CLOSED: proposal promotion guardrail failed")
	}
	return nil
}

func seal(receipt Receipt) Receipt {
	receipt.ReportDigest = ""
	receipt.ReportDigest = digestJSON(receipt)
	return receipt
}

func digestJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
