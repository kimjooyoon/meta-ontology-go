package coupling

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func cloneSemanticIR(input SemanticIR) SemanticIR {
	input.Nodes = append([]SemanticNode(nil), input.Nodes...)
	input.Relations = append([]SemanticRelation(nil), input.Relations...)
	for i := range input.Nodes {
		input.Nodes[i].Aliases = append([]string(nil), input.Nodes[i].Aliases...)
	}
	return input
}

func normalizeSemanticIR(input *SemanticIR) {
	sort.Slice(input.Nodes, func(i, j int) bool { return input.Nodes[i].ID < input.Nodes[j].ID })
	sort.Slice(input.Relations, func(i, j int) bool {
		left := input.Relations[i].Subject + "\x00" + input.Relations[i].Predicate + "\x00" + input.Relations[i].Object
		right := input.Relations[j].Subject + "\x00" + input.Relations[j].Predicate + "\x00" + input.Relations[j].Object
		return left < right
	})
	for i := range input.Nodes {
		sort.Strings(input.Nodes[i].Aliases)
	}
}

func pathToWire(path semantic.InferencePathV1) wirePath {
	out := wirePath{Version: path.Version, Edges: make([]wireEdge, 0, len(path.Edges)), Claims: make([]wireClaim, 0, len(path.Claims)), Evidence: make([]wireEvidence, 0, len(path.Evidence))}
	for _, edge := range path.Edges {
		out.Edges = append(out.Edges, wireEdgeFromSemantic(edge))
	}
	for _, claim := range path.Claims {
		out.Claims = append(out.Claims, wireClaimFromSemantic(claim))
	}
	for _, evidence := range path.Evidence {
		out.Evidence = append(out.Evidence, wireEvidenceFromSemantic(evidence))
	}
	return out
}

func pathFromWire(raw wirePath) (semantic.InferencePathV1, error) {
	out := semantic.InferencePathV1{Version: raw.Version, Edges: make([]semantic.InferenceEdge, 0, len(raw.Edges)), Claims: make([]semantic.SemanticChangeClaim, 0, len(raw.Claims)), Evidence: make([]semantic.InferenceEvidence, 0, len(raw.Evidence))}
	for _, edge := range raw.Edges {
		value, err := semanticEdgeFromWire(edge)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		out.Edges = append(out.Edges, value)
	}
	for _, claim := range raw.Claims {
		value, err := semanticClaimFromWire(claim)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		out.Claims = append(out.Claims, value)
	}
	for _, evidence := range raw.Evidence {
		value, err := semanticEvidenceFromWire(evidence)
		if err != nil {
			return semantic.InferencePathV1{}, err
		}
		out.Evidence = append(out.Evidence, value)
	}
	return out, nil
}

func wireEdgeFromSemantic(edge semantic.InferenceEdge) wireEdge {
	record := wireRecordFromSemantic(edge.InferenceRecord)
	return wireEdge{RecordID: record.RecordID, SubjectID: record.SubjectID, ObjectID: record.ObjectID, Rule: record.Rule, Phase: record.Phase, Ordinal: record.Ordinal, Before: record.Before, After: record.After, Authority: record.Authority, Evidence: record.Evidence, Controls: record.Controls, Kind: edge.Kind.String(), SourceRoots: idsToStrings(edge.SourceRoots), AcceptanceReceipt: edge.AcceptanceReceipt.String()}
}

func wireClaimFromSemantic(claim semantic.SemanticChangeClaim) wireClaim {
	record := wireRecordFromSemantic(claim.InferenceRecord)
	return wireClaim{RecordID: record.RecordID, SubjectID: record.SubjectID, ObjectID: record.ObjectID, Rule: record.Rule, Phase: record.Phase, Ordinal: record.Ordinal, Before: record.Before, After: record.After, Authority: record.Authority, Evidence: record.Evidence, Controls: record.Controls, Kind: claim.Kind.String(), CanonicalDelta: claim.CanonicalDelta, DeltaDigest: claim.DeltaDigest}
}

func wireRecordFromSemantic(record semantic.InferenceRecord) wireRecord {
	return wireRecord{RecordID: record.RecordID.String(), SubjectID: record.SubjectID.String(), ObjectID: record.ObjectID.String(), Rule: wireRule{ID: record.Rule.ID.String(), Version: record.Rule.Version, Digest: record.Rule.Digest}, Phase: record.Phase.Phase.String(), Ordinal: record.Phase.Ordinal, Before: wireSnapshot{Source: record.Before.Source, Semantic: record.Before.Semantic}, After: wireSnapshot{Source: record.After.Source, Semantic: record.After.Semantic}, Authority: wireAuthority{Layer: record.Authority.Layer.String(), Effect: record.Authority.Effect.String()}, Evidence: evidenceRefsToWire(record.Evidence), Controls: wireControlsFromSemantic(record.Controls)}
}

func wireEvidenceFromSemantic(record semantic.InferenceEvidence) wireEvidence {
	return wireEvidence{ID: record.ID.String(), Digest: record.Digest, Before: wireSnapshot{Source: record.Before.Source, Semantic: record.Before.Semantic}, After: wireSnapshot{Source: record.After.Source, Semantic: record.After.Semantic}, SourceBacked: record.SourceBacked, Independent: record.Independent, Controls: wireControlsFromSemantic(record.Controls)}
}

func wireControlsFromSemantic(value semantic.InferenceControls) wireControls {
	return wireControls{CatalogDigest: value.CatalogDigest, PolicyDigest: value.PolicyDigest, Profile: wireProfile{ID: value.Profile.ID, Version: value.Profile.Version, Digest: value.Profile.Digest}}
}

func evidenceRefsToWire(refs []semantic.EvidenceReference) []wireEvidenceRef {
	out := make([]wireEvidenceRef, 0, len(refs))
	for _, ref := range refs {
		out = append(out, wireEvidenceRef{ID: ref.ID.String(), Digest: ref.Digest})
	}
	return out
}

func idsToStrings(ids []semantic.ID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

func semanticEdgeFromWire(raw wireEdge) (semantic.InferenceEdge, error) {
	record, err := semanticRecordFromWire(wireRecord{RecordID: raw.RecordID, SubjectID: raw.SubjectID, ObjectID: raw.ObjectID, Rule: raw.Rule, Phase: raw.Phase, Ordinal: raw.Ordinal, Before: raw.Before, After: raw.After, Authority: raw.Authority, Evidence: raw.Evidence, Controls: raw.Controls})
	if err != nil {
		return semantic.InferenceEdge{}, err
	}
	kind := semantic.InferenceKind(raw.Kind)
	if !kind.Valid() {
		return semantic.InferenceEdge{}, fmt.Errorf("unknown inference kind %q", raw.Kind)
	}
	roots := make([]semantic.ID, 0, len(raw.SourceRoots))
	for _, root := range raw.SourceRoots {
		id, err := semantic.ParseIdentity(root)
		if err != nil {
			return semantic.InferenceEdge{}, err
		}
		roots = append(roots, id)
	}
	var receipt semantic.ID
	if raw.AcceptanceReceipt != "" {
		receipt, err = semantic.ParseIdentity(raw.AcceptanceReceipt)
		if err != nil {
			return semantic.InferenceEdge{}, err
		}
	}
	return semantic.InferenceEdge{InferenceRecord: record, Kind: kind, SourceRoots: roots, AcceptanceReceipt: receipt}, nil
}

func semanticClaimFromWire(raw wireClaim) (semantic.SemanticChangeClaim, error) {
	record, err := semanticRecordFromWire(wireRecord{RecordID: raw.RecordID, SubjectID: raw.SubjectID, ObjectID: raw.ObjectID, Rule: raw.Rule, Phase: raw.Phase, Ordinal: raw.Ordinal, Before: raw.Before, After: raw.After, Authority: raw.Authority, Evidence: raw.Evidence, Controls: raw.Controls})
	if err != nil {
		return semantic.SemanticChangeClaim{}, err
	}
	return semantic.SemanticChangeClaim{InferenceRecord: record, Kind: semantic.SemanticChangeKind(raw.Kind), CanonicalDelta: raw.CanonicalDelta, DeltaDigest: raw.DeltaDigest}, nil
}

func semanticEvidenceFromWire(raw wireEvidence) (semantic.InferenceEvidence, error) {
	id, err := semantic.ParseIdentity(raw.ID)
	if err != nil {
		return semantic.InferenceEvidence{}, err
	}
	return semantic.InferenceEvidence{ID: id, Digest: raw.Digest, Before: semantic.SnapshotDigests{Source: raw.Before.Source, Semantic: raw.Before.Semantic}, After: semantic.SnapshotDigests{Source: raw.After.Source, Semantic: raw.After.Semantic}, SourceBacked: raw.SourceBacked, Independent: raw.Independent, Controls: semanticControlsFromWire(raw.Controls)}, nil
}

func semanticRecordFromWire(raw wireRecord) (semantic.InferenceRecord, error) {
	recordID, err := semantic.ParseIdentity(raw.RecordID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	subjectID, err := semantic.ParseIdentity(raw.SubjectID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	objectID, err := semantic.ParseIdentity(raw.ObjectID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	ruleID, err := semantic.ParseIdentity(raw.Rule.ID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	evidence := make([]semantic.EvidenceReference, 0, len(raw.Evidence))
	for _, ref := range raw.Evidence {
		id, parseErr := semantic.ParseIdentity(ref.ID)
		if parseErr != nil {
			return semantic.InferenceRecord{}, parseErr
		}
		evidence = append(evidence, semantic.EvidenceReference{ID: id, Digest: ref.Digest})
	}
	return semantic.InferenceRecord{RecordID: recordID, SubjectID: subjectID, ObjectID: objectID, Rule: semantic.RuleBinding{ID: ruleID, Version: raw.Rule.Version, Digest: raw.Rule.Digest}, Phase: semantic.PhasePlacement{Phase: semantic.InferencePhase(raw.Phase), Ordinal: raw.Ordinal}, Before: semantic.SnapshotDigests{Source: raw.Before.Source, Semantic: raw.Before.Semantic}, After: semantic.SnapshotDigests{Source: raw.After.Source, Semantic: raw.After.Semantic}, Authority: semantic.AuthorityBinding{Layer: semantic.AuthorityLayer(raw.Authority.Layer), Effect: semantic.AuthorityEffect(raw.Authority.Effect)}, Evidence: evidence, Controls: semanticControlsFromWire(raw.Controls)}, nil
}

func semanticControlsFromWire(raw wireControls) semantic.InferenceControls {
	return semantic.InferenceControls{CatalogDigest: raw.CatalogDigest, PolicyDigest: raw.PolicyDigest, Profile: semantic.ProfileBinding{ID: raw.Profile.ID, Version: raw.Profile.Version, Digest: raw.Profile.Digest}}
}

func decodeStrictJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delim == '[' {
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	}
	if delim != '{' {
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	seen := map[string]struct{}{}
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := keyToken.(string)
		if !ok {
			return fmt.Errorf("object key is not a string")
		}
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate JSON key %q", key)
		}
		seen[key] = struct{}{}
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}

func sortedUnique(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	copyValues := append([]string(nil), values...)
	sort.Strings(copyValues)
	out := copyValues[:1]
	for _, value := range copyValues[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}
