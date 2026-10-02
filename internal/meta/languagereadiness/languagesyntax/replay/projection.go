package replay

import (
	"bytes"
	"fmt"
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/receiptprojection"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// ProjectionResult describes the explicit structure projection profile. It
// makes no claim about the general semantic IR or its Get-Put/Put-Get laws.
type ProjectionResult struct {
	ObservedDecision   string   `json:"observed_decision"`
	Profile            string   `json:"profile"`
	SourceDigest       string   `json:"source_digest,omitempty"`
	CanonicalDigest    string   `json:"canonical_digest,omitempty"`
	StructureDigest    string   `json:"structure_digest,omitempty"`
	GoArtifactDigest   string   `json:"go_artifact_digest,omitempty"`
	JSONArtifactDigest string   `json:"json_artifact_digest,omitempty"`
	ASTReplayed        bool     `json:"ast_replayed"`
	ByteReplayed       bool     `json:"byte_replayed"`
	StructureReplayed  bool     `json:"structure_replayed"`
	ArtifactsMatched   bool     `json:"artifacts_matched"`
	Diagnostics        []string `json:"diagnostics"`
}

func ExecuteProjection(repository fs.FS, path, profile, root, goPath, jsonPath string) ProjectionResult {
	r := ProjectionResult{ObservedDecision: DecisionClosed, Profile: profile}
	reject := func(err error) ProjectionResult {
		r.Diagnostics = append(r.Diagnostics, err.Error())
		return r
	}
	if profile != receiptprojection.Profile {
		return reject(fmt.Errorf("projection.unsupported-profile"))
	}
	raw, err := fs.ReadFile(repository, path)
	if err != nil {
		r.ObservedDecision = DecisionUnknown
		return reject(fmt.Errorf("projection.read-source: %w", err))
	}
	r.SourceDigest = digestBytes(raw)
	p, err := receiptprojection.Compile(path, raw, root)
	if err != nil {
		return reject(err)
	}
	canonical, astOK, bytesOK, err := replayProjectionSyntax(path, raw)
	if err != nil {
		return reject(err)
	}
	r.ASTReplayed, r.ByteReplayed = astOK, bytesOK
	r.CanonicalDigest = digestBytes(canonical)
	replayed, err := receiptprojection.Compile(path, canonical, root)
	if err != nil {
		return reject(err)
	}
	r.StructureDigest, err = p.StructureDigest()
	if err != nil {
		return reject(err)
	}
	replayedDigest, err := replayed.StructureDigest()
	if err != nil {
		return reject(err)
	}
	r.StructureReplayed = r.StructureDigest == replayedDigest
	goBytes, err := p.Go()
	if err != nil {
		return reject(err)
	}
	jsonBytes, err := p.JSONSchema()
	if err != nil {
		return reject(err)
	}
	actualGo, goErr := fs.ReadFile(repository, goPath)
	actualJSON, jsonErr := fs.ReadFile(repository, jsonPath)
	if goErr != nil || jsonErr != nil {
		r.ObservedDecision = DecisionUnknown
		return reject(fmt.Errorf("projection.read-artifacts: go=%v json=%v", goErr, jsonErr))
	}
	r.GoArtifactDigest, r.JSONArtifactDigest = digestBytes(actualGo), digestBytes(actualJSON)
	r.ArtifactsMatched = bytes.Equal(goBytes, actualGo) && bytes.Equal(jsonBytes, actualJSON)
	if r.ASTReplayed && r.ByteReplayed && r.StructureReplayed && r.ArtifactsMatched {
		r.ObservedDecision = DecisionPass
		return r
	}
	return reject(fmt.Errorf("projection.replay-or-generated-artifact-mismatch"))
}

func replayProjectionSyntax(path string, raw []byte) ([]byte, bool, bool, error) {
	support := syntax.CurrentEntityFieldsSupport()
	support.State = syntax.EntityFieldsSupported
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(path, string(raw), support)
	if len(diagnostics) != 0 {
		return nil, false, false, fmt.Errorf("projection.parse: %v", diagnostics)
	}
	canonical, err := syntax.FormatWithEntityFieldsSupport(file, support)
	if err != nil {
		return nil, false, false, err
	}
	replayed, diagnostics := syntax.ParseFileWithEntityFieldsSupport(path, canonical, support)
	if len(diagnostics) != 0 {
		return nil, false, false, fmt.Errorf("projection.reparse: %v", diagnostics)
	}
	second, err := syntax.FormatWithEntityFieldsSupport(replayed, support)
	if err != nil {
		return nil, false, false, err
	}
	left, err := astShape(file)
	if err != nil {
		return nil, false, false, err
	}
	right, err := astShape(replayed)
	return []byte(canonical), left == right, canonical == second, err
}
