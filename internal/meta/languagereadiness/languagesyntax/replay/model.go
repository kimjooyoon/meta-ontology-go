package replay

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
)

const (
	DecisionPass    = "PASS"
	DecisionClosed  = "FAIL_CLOSED"
	DecisionUnknown = "UNKNOWN"
)

type FileObservation struct {
	Path         string `json:"path"`
	GoooLines    int    `json:"gooo_lines"`
	SourceDigest string `json:"source_digest"`
}

type Result struct {
	ObservedDecision  string   `json:"observed_decision"`
	SourceLines       int      `json:"source_lines"`
	SourceDigest      string   `json:"source_digest,omitempty"`
	ASTDigest         string   `json:"ast_digest,omitempty"`
	CanonicalDigest   string   `json:"canonical_digest,omitempty"`
	SemanticDigest    string   `json:"semantic_digest,omitempty"`
	ASTReplayed       bool     `json:"ast_replayed"`
	ByteReplayed      bool     `json:"byte_replayed"`
	SemanticReplayed  bool     `json:"semantic_replayed"`
	GetPut            bool     `json:"get_put"`
	PutGet            bool     `json:"put_get"`
	DiagnosticRejected bool    `json:"diagnostic_rejected"`
	Diagnostics       []string `json:"diagnostics"`
}

func sourceLines(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	lines := bytes.Count(raw, []byte{'\n'})
	if raw[len(raw)-1] != '\n' {
		lines++
	}
	return lines
}

func digestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}
