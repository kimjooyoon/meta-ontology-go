package externaloraclehumility

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
)

const (
	ContractSchema         = "gooo/external-oracle-humility-contract/v1"
	EvidenceSchema         = "gooo/external-oracle-reference-set/v1"
	ReceiptSchema          = "gooo/external-oracle-humility-receipt/v1"
	ReportSchema           = "gooo/external-oracle-humility-report/v1"
	SuiteSchema            = "gooo/external-oracle-humility-suite/v1"
	DenominatorVersion     = "gooo/external-oracle-humility-denominator/v1"
	CaseDenominatorVersion = "gooo/external-oracle-humility-cases/v1"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

var denominator = []Criterion{
	{"source-receipt-bound", "DRIVER", "FOUNDATION", "source-receipt-producer", "independent-judge", "bind-gooo-source-receipt", "observe", "source-receipt", "claims", "EQUALS", 1},
	{"source-declarations-reread", "DRIVER", "FOUNDATION", "independent-judge", "source-receipt-consumer", "reread-gooo-declarations", "observe", "source-claims", "declarations", "EQUALS", 1},
	{"receipt-roles-bound", "DRIVER", "FOUNDATION", "independent-judge", "receipt-audit", "bind-receipt-roles", "bind", "producer-consumer", "roles", "EQUALS", 1},
	{"gomacro-reference-available", "DRIVER", "FOUNDATION", "reference-capsule", "independent-judge", "bind-gomacro-comparison", "compare", "gomacro", "reference", "EQUALS", 1},
	{"racket-reference-available", "DRIVER", "FOUNDATION", "reference-capsule", "independent-judge", "bind-racket-comparison", "compare", "racket", "reference", "EQUALS", 1},
	{"reproducible-builds-reference-available", "DRIVER", "FOUNDATION", "reference-capsule", "independent-judge", "bind-reproducibility-comparison", "compare", "reproducible-builds", "reference", "EQUALS", 1},
	{"comparative-relations-exact", "OUTCOME", "COHERENCE", "independent-judge", "reference-agreement-consumer", "compare-reference-relations", "compare", "relations", "comparison", "EQUALS", 1},
	{"agreement-state-classified", "OUTCOME", "COHERENCE", "independent-judge", "reference-agreement-consumer", "classify-reference-agreement", "classify", "agreement", "state", "EQUALS", 1},
	{"external-authority-refused", "GUARDRAIL", "REGRESSION", "independent-judge", "semantic-authority-governor", "refuse-external-semantic-authority", "govern", "authority", "grant", "EQUALS", 1},
	{"claim-transition-persisted", "OUTCOME", "COHERENCE", "independent-judge", "claim-ledger", "persist-claim-transition", "persist", "claim-transition", "transitions", "EQUALS", 1},
	{"read-only-boundary", "GUARDRAIL", "REGRESSION", "independent-judge", "read-only-governor", "enforce-read-only-boundary", "guard", "effects", "writes", "EQUALS", 0},
	{"pass-promotion-refused", "GUARDRAIL", "REGRESSION", "independent-judge", "semantic-authority-governor", "refuse-pass-promotion", "govern", "promotion", "decision", "EQUALS", 1},
}

func Criteria() []Criterion { return append([]Criterion(nil), denominator...) }

func Digest(value any) string {
	raw, _ := json.Marshal(value)
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DigestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DenominatorDigest() string { return Digest(denominator) }

func validDigest(value string) bool { return digestPattern.MatchString(value) }

func validateContract(c Contract) error {
	if c.Schema != ContractSchema || c.Version != 1 || c.Source.Authority != "GOOO_SOURCE_INTENT" {
		return fmt.Errorf("contract identity or source authority mismatch")
	}
	if !validDigest(c.Source.SHA256) || c.FixedDenominator.Version != DenominatorVersion ||
		c.FixedDenominator.Total != len(denominator) || c.FixedDenominator.BasisPointsGoal != 10000 {
		return fmt.Errorf("fixed denominator contract mismatch")
	}
	if len(c.Source.Claims) != 3 || len(c.References) != 3 || len(c.Cases) != 3 {
		return fmt.Errorf("contract cardinality mismatch")
	}
	claimIDs := make(map[string]bool, len(c.Source.Claims))
	for _, claim := range c.Source.Claims {
		if claim.ID == "" || claim.Text == "" || claim.State == "" || claimIDs[claim.ID] {
			return fmt.Errorf("source claim identity mismatch")
		}
		claimIDs[claim.ID] = true
	}
	referenceIDs := make(map[string]bool, len(c.References))
	for _, reference := range c.References {
		if reference.ID == "" || reference.ClaimID == "" || !reference.Primary ||
			reference.Relation != "COMPARATIVE_EVIDENCE" || reference.Authority != "NOT_AUTHORITY" ||
			!validDigest(reference.DocumentSHA256) || !claimIDs[reference.ClaimID] || referenceIDs[reference.ID] {
			return fmt.Errorf("reference %q is not comparison-only", reference.ID)
		}
		referenceIDs[reference.ID] = true
	}
	return nil
}
