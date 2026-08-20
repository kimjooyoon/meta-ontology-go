package coupling

import (
	"fmt"
)

type wireClaim struct {
	RecordID       string            `json:"record_id"`
	SubjectID      string            `json:"subject_id"`
	ObjectID       string            `json:"object_id"`
	Rule           wireRule          `json:"rule"`
	Phase          string            `json:"phase"`
	Ordinal        uint64            `json:"ordinal"`
	Before         wireSnapshot      `json:"before"`
	After          wireSnapshot      `json:"after"`
	Authority      wireAuthority     `json:"authority"`
	Evidence       []wireEvidenceRef `json:"evidence"`
	Controls       wireControls      `json:"controls"`
	Kind           string            `json:"kind"`
	CanonicalDelta string            `json:"canonical_delta,omitempty"`
	DeltaDigest    string            `json:"delta_digest,omitempty"`
}
type wirePath struct {
	Version  string         `json:"version"`
	Edges    []wireEdge     `json:"edges"`
	Claims   []wireClaim    `json:"claims"`
	Evidence []wireEvidence `json:"evidence"`
}
type wireInput struct {
	Schema                string                    `json:"schema"`
	FixtureID             string                    `json:"fixture_id"`
	RegistryDigest        string                    `json:"registry_digest"`
	Config                wireConfig                `json:"config"`
	Manifest              SourceManifest            `json:"manifest"`
	ResourceRegistry      ResourceBindingConfig     `json:"resource_registry"`
	AuthoritySourceBefore string                    `json:"authority_source_before"`
	AuthoritySourceAfter  string                    `json:"authority_source_after"`
	SemanticBefore        SemanticIR                `json:"semantic_before"`
	SemanticAfter         SemanticIR                `json:"semantic_after"`
	Registry              []CodeBinding             `json:"registry"`
	Changes               []CodeChange              `json:"changes"`
	Receipts              []CouplingReceipt         `json:"receipts"`
	ResourceReceipts      []ExternalResourceReceipt `json:"resource_receipts"`
	Roots                 []string                  `json:"roots"`
	Path                  wirePath                  `json:"path"`
}

func DecodeInput(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, err
	}
	var raw wireInput
	if err := decodeStrictJSON(data, &raw); err != nil {
		return Input{}, fmt.Errorf("decode coupling input: %w", err)
	}
	path, err := pathFromWire(raw.Path)
	if err != nil {
		return Input{}, fmt.Errorf("decode coupling path: %w", err)
	}
	return Input{
		Schema: raw.Schema, FixtureID: raw.FixtureID, RegistryDigest: raw.RegistryDigest,
		Config: EvaluationConfig{ToolchainDigest: raw.Config.ToolchainDigest, Profile: ProfileConfig{
			ID: raw.Config.Profile.ID, Version: raw.Config.Profile.Version, Digest: raw.Config.Profile.Digest,
		}, ResourceBinding: raw.Config.ResourceBinding}, Manifest: raw.Manifest, ResourceRegistry: raw.ResourceRegistry,
		AuthoritySourceBefore: raw.AuthoritySourceBefore, AuthoritySourceAfter: raw.AuthoritySourceAfter,
		SemanticBefore: raw.SemanticBefore, SemanticAfter: raw.SemanticAfter,
		Registry: raw.Registry, Changes: raw.Changes, Receipts: raw.Receipts, ResourceReceipts: raw.ResourceReceipts, Roots: raw.Roots, Path: path,
	}, nil
}
