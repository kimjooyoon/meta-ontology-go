package coupling

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

type wireSnapshot struct {
	Source   string `json:"source"`
	Semantic string `json:"semantic"`
}

type wireProfile struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type wireConfig struct {
	ToolchainDigest string                `json:"toolchain_digest"`
	Profile         wireProfile           `json:"profile"`
	ResourceBinding ResourceBindingConfig `json:"resource_binding"`
}

type wireRule struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type wireControls struct {
	CatalogDigest string      `json:"catalog_digest,omitempty"`
	PolicyDigest  string      `json:"policy_digest,omitempty"`
	Profile       wireProfile `json:"profile"`
}

type wireEvidenceRef struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}

type wireEvidence struct {
	ID           string       `json:"id"`
	Digest       string       `json:"digest"`
	Before       wireSnapshot `json:"before"`
	After        wireSnapshot `json:"after"`
	SourceBacked bool         `json:"source_backed"`
	Independent  bool         `json:"independent"`
	Controls     wireControls `json:"controls"`
}

type wireRecord struct {
	RecordID  string            `json:"record_id"`
	SubjectID string            `json:"subject_id"`
	ObjectID  string            `json:"object_id"`
	Rule      wireRule          `json:"rule"`
	Phase     string            `json:"phase"`
	Ordinal   uint64            `json:"ordinal"`
	Before    wireSnapshot      `json:"before"`
	After     wireSnapshot      `json:"after"`
	Authority wireAuthority     `json:"authority"`
	Evidence  []wireEvidenceRef `json:"evidence"`
	Controls  wireControls      `json:"controls"`
}

type wireAuthority struct {
	Layer  string `json:"layer"`
	Effect string `json:"effect"`
}

type wireEdge struct {
	RecordID          string            `json:"record_id"`
	SubjectID         string            `json:"subject_id"`
	ObjectID          string            `json:"object_id"`
	Rule              wireRule          `json:"rule"`
	Phase             string            `json:"phase"`
	Ordinal           uint64            `json:"ordinal"`
	Before            wireSnapshot      `json:"before"`
	After             wireSnapshot      `json:"after"`
	Authority         wireAuthority     `json:"authority"`
	Evidence          []wireEvidenceRef `json:"evidence"`
	Controls          wireControls      `json:"controls"`
	Kind              string            `json:"kind"`
	SourceRoots       []string          `json:"source_roots"`
	AcceptanceReceipt string            `json:"acceptance_receipt,omitempty"`
}

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

func EncodeInputJSON(input Input) ([]byte, error) {
	raw := inputToWire(input, false)
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// CanonicalInputBytes excludes fixture labels. The result is the exact byte
// sequence used for InputDigest, so fixture expectations cannot affect it.
func CanonicalInputBytes(input Input) ([]byte, error) {
	raw := inputToWire(input, true)
	return json.Marshal(raw)
}

func CanonicalInputDigest(input Input) string {
	data, err := CanonicalInputBytes(input)
	if err != nil {
		return digestBytes([]byte("canonical-input-error:" + err.Error()))
	}
	return digestBytes(data)
}

type outputDigestView struct {
	Schema               string              `json:"schema"`
	InputDigest          string              `json:"input_digest"`
	Decision             Decision            `json:"decision"`
	Reason               Reason              `json:"reason"`
	AcceptedSurfaces     []string            `json:"accepted_surfaces"`
	ChangedSurfaces      []string            `json:"changed_surfaces"`
	ReceiptSurfaces      []string            `json:"receipt_surfaces"`
	SemanticBeforeDigest string              `json:"semantic_before_digest"`
	SemanticAfterDigest  string              `json:"semantic_after_digest"`
	SemanticDeltaDigest  string              `json:"semantic_delta_digest,omitempty"`
	PathClosureDigest    string              `json:"path_closure_digest,omitempty"`
	ObservationCounts    ObservationCounts   `json:"observation_counts"`
	Resources            ResourceObservation `json:"resources"`
}

func CanonicalOutputDigest(output Output) string {
	view := outputDigestView{
		Schema: output.Schema, InputDigest: output.InputDigest, Decision: output.Decision, Reason: output.Reason,
		AcceptedSurfaces: sortedUnique(output.AcceptedSurfaces),
		ChangedSurfaces:  sortedUnique(output.ChangedSurfaces), ReceiptSurfaces: sortedUnique(output.ReceiptSurfaces),
		SemanticBeforeDigest: output.SemanticBeforeDigest, SemanticAfterDigest: output.SemanticAfterDigest,
		SemanticDeltaDigest: output.SemanticDeltaDigest, PathClosureDigest: output.PathClosureDigest,
		ObservationCounts: output.ObservationCounts, Resources: output.Resources,
	}
	data, _ := json.Marshal(view)
	return digestBytes(data)
}

func ReplayDigest(inputDigest, outputDigest string) string {
	return digestBytes([]byte(inputDigest + "\x00" + outputDigest))
}

func EncodeOutputJSON(output Output) ([]byte, error) {
	output.ChangedSurfaces = sortedUnique(output.ChangedSurfaces)
	output.ReceiptSurfaces = sortedUnique(output.ReceiptSurfaces)
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func inputToWire(input Input, canonical bool) wireInput {
	raw := wireInput{
		Schema: input.Schema, FixtureID: input.FixtureID, RegistryDigest: input.RegistryDigest,
		Config: wireConfig{ToolchainDigest: input.Config.ToolchainDigest, Profile: wireProfile{
			ID: input.Config.Profile.ID, Version: input.Config.Profile.Version, Digest: input.Config.Profile.Digest,
		}, ResourceBinding: input.Config.ResourceBinding}, Manifest: input.Manifest, ResourceRegistry: input.ResourceRegistry,
		AuthoritySourceBefore: input.AuthoritySourceBefore, AuthoritySourceAfter: input.AuthoritySourceAfter,
		SemanticBefore: cloneSemanticIR(input.SemanticBefore), SemanticAfter: cloneSemanticIR(input.SemanticAfter),
		Registry: append([]CodeBinding(nil), input.Registry...), Changes: append([]CodeChange(nil), input.Changes...),
		Receipts: append([]CouplingReceipt(nil), input.Receipts...), Roots: append([]string(nil), input.Roots...),
		ResourceReceipts: append([]ExternalResourceReceipt(nil), input.ResourceReceipts...),
		Path:             pathToWire(input.Path),
	}
	if canonical {
		raw.FixtureID = ""
		for i := range raw.SemanticBefore.Nodes {
			raw.SemanticBefore.Nodes[i].Name = ""
			raw.SemanticBefore.Nodes[i].Aliases = nil
		}
		for i := range raw.SemanticAfter.Nodes {
			raw.SemanticAfter.Nodes[i].Name = ""
			raw.SemanticAfter.Nodes[i].Aliases = nil
		}
		for i := range raw.Registry {
			raw.Registry[i].PackageLabel = ""
			raw.Registry[i].FileLabel = ""
			raw.Registry[i].SourceSpan = ""
		}
		normalizeWireInput(&raw)
	}
	return raw
}

func normalizeWireInput(raw *wireInput) {
	sort.Slice(raw.Registry, func(i, j int) bool {
		return raw.Registry[i].RegisteredSurfaceID+"\x00"+raw.Registry[i].CodeSymbolID <
			raw.Registry[j].RegisteredSurfaceID+"\x00"+raw.Registry[j].CodeSymbolID
	})
	sort.Slice(raw.Changes, func(i, j int) bool { return raw.Changes[i].CodeSymbolID < raw.Changes[j].CodeSymbolID })
	sort.Slice(raw.Receipts, func(i, j int) bool {
		return raw.Receipts[i].SurfaceID+"\x00"+raw.Receipts[i].ReceiptID < raw.Receipts[j].SurfaceID+"\x00"+raw.Receipts[j].ReceiptID
	})
	sort.Slice(raw.ResourceReceipts, func(i, j int) bool {
		return raw.ResourceReceipts[i].Metric+"\x00"+raw.ResourceReceipts[i].ReceiptID < raw.ResourceReceipts[j].Metric+"\x00"+raw.ResourceReceipts[j].ReceiptID
	})
	sort.Strings(raw.Roots)
	normalizeSemanticIR(&raw.SemanticBefore)
	normalizeSemanticIR(&raw.SemanticAfter)
	sort.Slice(raw.Path.Edges, func(i, j int) bool { return raw.Path.Edges[i].RecordID < raw.Path.Edges[j].RecordID })
	sort.Slice(raw.Path.Claims, func(i, j int) bool { return raw.Path.Claims[i].RecordID < raw.Path.Claims[j].RecordID })
	sort.Slice(raw.Path.Evidence, func(i, j int) bool { return raw.Path.Evidence[i].ID < raw.Path.Evidence[j].ID })
	for i := range raw.Path.Edges {
		sort.Strings(raw.Path.Edges[i].SourceRoots)
		sort.Slice(raw.Path.Edges[i].Evidence, func(a, b int) bool { return raw.Path.Edges[i].Evidence[a].ID < raw.Path.Edges[i].Evidence[b].ID })
	}
	for i := range raw.Path.Claims {
		sort.Slice(raw.Path.Claims[i].Evidence, func(a, b int) bool { return raw.Path.Claims[i].Evidence[a].ID < raw.Path.Claims[i].Evidence[b].ID })
	}
}
