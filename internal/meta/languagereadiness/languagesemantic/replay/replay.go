package replay

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var ExpectedStages = []string{
	"READ_SOURCE",
	"PARSE_AST",
	"LOWER_IR",
	"NORMALIZE_IR",
	"REPLAY_IR",
	"SEAL_EFFECTS",
}

type EffectReceipt struct {
	Reads     []string `json:"reads"`
	Writes    int      `json:"writes"`
	Network   int      `json:"network"`
	Processes int      `json:"processes"`
}

type Observation struct {
	Path               string        `json:"path"`
	SourceLines        int           `json:"source_lines"`
	SourceDigest       string        `json:"source_digest"`
	IRVersion          string        `json:"ir_version"`
	Package            string        `json:"package"`
	Namespace          string        `json:"namespace"`
	Nodes              int           `json:"nodes"`
	DeterministicFacts int           `json:"deterministic_facts"`
	CandidateFacts     int           `json:"candidate_facts"`
	Normalized         bool          `json:"normalized"`
	CanonicalReplay    bool          `json:"canonical_replay"`
	SemanticReplay     bool          `json:"semantic_replay"`
	ProvenanceReplay   bool          `json:"provenance_replay"`
	EvidenceReplay     bool          `json:"evidence_replay"`
	SemanticHash       string        `json:"semantic_hash"`
	ProvenanceHash     string        `json:"provenance_hash"`
	EvidenceHash       string        `json:"evidence_hash"`
	Stages             []string      `json:"stages"`
	Effects            EffectReceipt `json:"effects"`
	IR                 semantic.IR   `json:"-"`
}

func Observe(root, relativePath string) (Observation, error) {
	target, cleanPath, err := sourcePath(root, relativePath)
	if err != nil {
		return Observation{}, err
	}
	source, err := os.ReadFile(target)
	if err != nil {
		return Observation{}, fmt.Errorf("read %s: %w", cleanPath, err)
	}
	first, err := lower(cleanPath, source)
	if err != nil {
		return Observation{}, err
	}
	second, err := lower(cleanPath, source)
	if err != nil {
		return Observation{}, fmt.Errorf("replay %s: %w", cleanPath, err)
	}
	comparison := semantic.CompareIR(first, second)
	return Observation{
		Path:               cleanPath,
		SourceLines:        physicalLines(source),
		SourceDigest:       semantic.StableHash(source),
		IRVersion:          first.Version,
		Package:            first.Package,
		Namespace:          first.Namespace.String(),
		Nodes:              len(first.Graph.Nodes()),
		DeterministicFacts: len(first.Graph.DeterministicFacts()),
		CandidateFacts:     len(first.Graph.Candidates()),
		Normalized:         first.Validate() == nil,
		CanonicalReplay:    first.Canonical() == second.Canonical(),
		SemanticReplay:     comparison.SemanticEqual,
		ProvenanceReplay:   comparison.ProvenanceEqual,
		EvidenceReplay:     comparison.ExactEvidenceEqual,
		SemanticHash:       first.StableHash(),
		ProvenanceHash:     first.ProvenanceHash(),
		EvidenceHash:       first.EvidenceHash(),
		Stages:             append([]string(nil), ExpectedStages...),
		Effects: EffectReceipt{
			Reads: []string{cleanPath},
		},
		IR: first,
	}, nil
}

func lower(path string, source []byte) (semantic.IR, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if diagnostics.HasErrors() {
		return semantic.IR{}, fmt.Errorf("parse %s: diagnostics contain errors", path)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return semantic.IR{}, fmt.Errorf("lower %s: %w", path, err)
	}
	normalized, err := ir.Normalized()
	if err != nil {
		return semantic.IR{}, fmt.Errorf("normalize %s: %w", path, err)
	}
	if err := normalized.Validate(); err != nil {
		return semantic.IR{}, fmt.Errorf("validate %s: %w", path, err)
	}
	return normalized, nil
}

func sourcePath(root, relativePath string) (string, string, error) {
	clean := filepath.Clean(filepath.FromSlash(strings.TrimSpace(relativePath)))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("source path %q is outside the repository", relativePath)
	}
	if filepath.Ext(clean) != ".gooo" {
		return "", "", fmt.Errorf("source path %q is not a .gooo file", relativePath)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", "", err
	}
	target := filepath.Join(rootAbs, clean)
	relative, err := filepath.Rel(rootAbs, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("source path %q escapes the repository", relativePath)
	}
	return target, filepath.ToSlash(clean), nil
}

func physicalLines(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	lines := bytes.Count(source, []byte{'\n'})
	if source[len(source)-1] != '\n' {
		lines++
	}
	return lines
}
