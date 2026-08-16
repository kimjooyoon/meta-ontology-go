//go:build detector_bridge

package coupling

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const detectorAuthorityHead = "02e35a01946c20c5de67f2ec71eeca5ac6c3aedb"

// This vector includes the producer packet bindings as well as the public
// result. A plausible status cannot hide a changed registry, manifest,
// receipt, path, or external-resource authority.
type productionBindingVector struct {
	AuthoritySchema                  string
	AuthorityRegistryDigest          string
	AuthorityToolchainDigest         string
	AuthorityProfileDigest           string
	AuthoritySnapshotDigest          string
	AuthorityProviderDigest          string
	AuthorityObserverDigest          string
	AuthorityBaselineDigest          string
	AuthorityExternalReceiptRequired bool
	PacketRegistryDigest             string
	ConfigRegistryDigest             string
	ConfigToolchainDigest            string
	ConfigProfileDigest              string
	ConfigSnapshotDigest             string
	ExpectedProviderDigest           string
	ExpectedObserverDigest           string
	BaselineDigest                   string
	ManifestComplete                 bool
	ManifestZeroChange               bool
	ManifestRegistryDigest           string
	ManifestToolchainDigest          string
	ManifestProfileDigest            string
	ManifestBeforeSnapshotDigest     string
	ManifestAfterSnapshotDigest      string
	ManifestDigest                   string
	PathDigest                       string
	ExternalReceiptDigest            string
	Surfaces                         []productionSurfaceBinding
	ManifestEntries                  []productionManifestBinding
	Receipts                         []productionReceiptBinding
	ExternalReceipt                  productionExternalBinding
}

type productionSurfaceBinding struct {
	SurfaceID       string
	CodeSymbolID    string
	SemanticOwnerID string
	SourceMapID     string
	BindingDigest   string
}

type productionManifestBinding struct {
	SurfaceID           string
	CodeSymbolID        string
	SemanticOwnerID     string
	BeforeBindingDigest string
	AfterBindingDigest  string
	BeforeBlobDigest    string
	AfterBlobDigest     string
}

type productionReceiptBinding struct {
	Schema                      string
	ReceiptID                   string
	SurfaceID                   string
	SemanticOwnerID             string
	CodeSymbolID                string
	SourceMapBindingDigest      string
	SnapshotDigest              string
	RegistryDigest              string
	ToolchainDigest             string
	ProfileDigest               string
	BeforeBlobDigest            string
	AfterBlobDigest             string
	BeforeAuthoritySourceDigest string
	AfterAuthoritySourceDigest  string
	BeforeSemanticDigest        string
	AfterSemanticDigest         string
	ChangeClaim                 production.ChangeClaim
	ReceiptKind                 semantic.SemanticChangeKind
	OriginPathIDs               []string
	ClaimID                     string
	EvidenceIDs                 []string
	EvidenceDigests             []string
	CanonicalDelta              string
	DeltaDigest                 string
	AuthoritySourceID           string
	AuthoritySourcePath         string
	State                       string
}

type productionExternalBinding struct {
	Schema                 string
	SnapshotDigest         string
	ProviderDigest         string
	ObserverDigest         string
	CPUWorkUnits           *uint64
	PeakMemoryBytes        *uint64
	DeterministicWorkUnits *uint64
	Digest                 string
}

type productionVector struct {
	Schema       string
	Decision     Decision
	Reasons      []production.Reason
	Accepted     []string
	Observation  production.ObservationVector
	FullSuite    bool
	InputDigest  string
	ResultDigest string
	Bindings     productionBindingVector
}

func TestProductionBridgePreparationHead(t *testing.T) {
	expected := literalProductionCorpusExpectations()
	if len(expected) != len(testCorpus()) {
		t.Fatalf("literal producer corpus has %d rows, want %d", len(expected), len(testCorpus()))
	}
	for index, row := range testCorpus() {
		want, ok := expected[row.Name]
		if !ok {
			t.Fatalf("missing literal producer expectation for %s", row.Name)
		}
		input := detectorInputFromCanonical(row.Input)
		authority := detectorAuthorityForCorpus(index)
		got := productionVectorFromResult(production.Evaluate(input, authority), input, authority)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s producer vector=%+v want=%+v detector_head=%s", row.Name, got, want, detectorAuthorityHead)
		}
	}
}

func TestProductionBridgeRejectsProducerOnlyResultMutations(t *testing.T) {
	row := testCorpus()[0]
	input := detectorInputFromCanonical(row.Input)
	authority := detectorAuthorityForCorpus(0)
	base := production.Evaluate(input, authority)
	for _, mutation := range literalResultMutations(base, input, authority) {
		gotResult := base
		mutation.mutate(&gotResult)
		got := productionVectorFromResult(gotResult, input, authority)
		if reflect.DeepEqual(got, mutation.want) {
			t.Errorf("producer-only %s mutation was not rejected: vector=%+v", mutation.name, got)
		}
	}
}

func TestProductionBridgeRejectsProducerInputMutations(t *testing.T) {
	row := testCorpus()[0]
	authority := detectorAuthorityForCorpus(0)
	for _, mutation := range literalInputMutations(detectorInputFromCanonical(row.Input), authority) {
		input := cloneProductionInput(mutation.input)
		got := production.Evaluate(input, authority)
		vector := productionVectorFromResult(got, input, authority)
		if !reflect.DeepEqual(vector, mutation.want) {
			t.Errorf("producer input %s vector=%+v want=%+v", mutation.name, vector, mutation.want)
		}
		if !reflect.DeepEqual(authority, mutation.authorityBefore) {
			t.Errorf("producer input %s changed evaluator authority", mutation.name)
		}
	}
}

func TestProductionBridgeExpectedOnlyMutationIsolation(t *testing.T) {
	row := testCorpus()[0]
	input := detectorInputFromCanonical(row.Input)
	authority := detectorAuthorityForCorpus(0)
	before := productionVectorFromResult(production.Evaluate(input, authority), input, authority)
	oracleBefore := Evaluate(row.Input)
	expected := literalProductionCorpusExpectations()
	mutated := expected[row.Name]
	mutated.Decision = DecisionFailClosed
	mutated.Reasons = []production.Reason{{Code: production.ReasonCode("EXPECTED_ONLY"), Detail: "fixture mutation"}}
	mutated.Observation.CPU.Value = 999
	expected[row.Name] = mutated
	after := productionVectorFromResult(production.Evaluate(input, authority), input, authority)
	oracleAfter := Evaluate(row.Input)
	if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(oracleBefore, oracleAfter) {
		t.Fatalf("expected-only fixture mutation affected subject output: before=%+v after=%+v", before, after)
	}
	if reflect.DeepEqual(expected[row.Name], before) {
		t.Fatal("expected-only mutation did not mutate fixture expectation")
	}
	presentation := cloneInput(row.Input)
	presentation.FixtureID = "fixture-label/bridge-mutated"
	presentation.Registry[0].PackageLabel = "renamed-package"
	presentation.Registry[0].FileLabel = "renamed.go"
	presentation.Registry[0].SourceSpan = "99:1-99:2"
	presentationVector := productionVectorFromResult(production.Evaluate(detectorInputFromCanonical(presentation), authority), detectorInputFromCanonical(presentation), authority)
	if !reflect.DeepEqual(presentationVector, before) {
		t.Fatalf("presentation-only producer packet mutation changed authoritative vector: got=%+v want=%+v", presentationVector, before)
	}
}

func TestProductionBridgeSubjectsAreReadOnlyAndCrossChecked(t *testing.T) {
	for index, row := range testCorpus() {
		canonical := cloneInput(row.Input)
		before := canonical
		oracle := Evaluate(canonical)
		if !reflect.DeepEqual(canonical, before) {
			t.Fatalf("%s independent oracle mutated canonical input", row.Name)
		}
		input := detectorInputFromCanonical(row.Input)
		authority := detectorAuthorityForCorpus(index)
		producerInput := cloneProductionInput(input)
		_ = production.Evaluate(producerInput, authority)
		if oracle.Decision != row.Expected.Decision || oracle.Reason != row.Expected.Reason {
			t.Fatalf("%s independent oracle fixture changed: got=%+v want=%+v", row.Name, oracle, row.Expected)
		}
	}
}

// The projection below is run once, before any producer-only mutation. It
// never receives a mutated producer packet and therefore cannot repair one.
func detectorInputFromCanonical(input Input) production.Input {
	registry := productionRegistryFromCanonical(input)
	config := productionConfigFromCanonical(input, registry)
	manifest := productionManifestFromCanonical(input, registry, config)
	path := productionPathFromCanonical(input.Path, input.Receipts, input.Roots)
	return production.Input{
		Schema: production.InputSchemaV1, Config: config, Registry: registry, Manifest: manifest,
		Receipts: productionReceiptsFromCanonical(input, registry, config, path), InferencePath: path,
		ExternalReceipt: productionExternalReceiptFromCanonical(input, config), WorkspaceRoot: input.Config.Profile.ID,
	}
}

func productionRegistryFromCanonical(input Input) production.Registry {
	registry := production.Registry{Schema: production.RegistrySchemaV1}
	for _, raw := range input.Registry {
		surface := production.Surface{
			SurfaceID: bridgeID(raw.RegisteredSurfaceID), CodeSymbolID: bridgeID(raw.CodeSymbolID), SemanticOwnerID: bridgeID(raw.SemanticOwnerID),
			Binding:           production.SourceMapBinding{SourceMapID: bridgeSourceMapID(raw.SourceMapID), BindingDigest: bridgeBindingDigestValues(raw.RegisteredSurfaceID, raw.CodeSymbolID, raw.SemanticOwnerID, bridgeSourceMapID(raw.SourceMapID).String()), PackageLabel: raw.PackageLabel, FileLabel: raw.FileLabel, SourceSpan: raw.SourceSpan},
			PresentationLabel: raw.RegisteredSurfaceID,
		}
		registry.Surfaces = append(registry.Surfaces, surface)
	}
	registry.Digest = bridgeRegistryDigest(registry)
	return registry
}

func productionConfigFromCanonical(input Input, registry production.Registry) production.Config {
	baseline := production.BaselineConfig{Schema: production.BaselineSchemaV1, FullSuiteRequired: true}
	baseline.Digest = bridgeBaselineDigest(baseline)
	return production.Config{Schema: production.ConfigSchemaV1, RegistryDigest: registry.Digest,
		ToolchainDigest: bridgeRawDigest(input.Config.ToolchainDigest), ProfileDigest: bridgeRawDigest(input.Config.Profile.Digest),
		SnapshotDigest: bridgeRawDigest(input.Config.ResourceBinding.SnapshotDigest), ExpectedProviderDigest: bridgeRawDigest(input.Config.ResourceBinding.ProviderDigest),
		ExpectedObserverDigest: bridgeRawDigest(input.Config.ResourceBinding.ObserverDigest), Baseline: baseline, ExternalReceiptRequired: true}
}

func productionManifestFromCanonical(input Input, registry production.Registry, config production.Config) production.ChangeManifest {
	manifest := production.ChangeManifest{Schema: production.ManifestSchemaV1, Complete: input.Manifest.Complete, ZeroChange: input.Manifest.ZeroChange,
		RegistryDigest: registry.Digest, ToolchainDigest: config.ToolchainDigest, ProfileDigest: config.ProfileDigest,
		BeforeSnapshotDigest: bridgeRawDigest(input.Manifest.BeforeSnapshotDigest), AfterSnapshotDigest: config.SnapshotDigest}
	changes := make(map[string]CodeChange, len(input.Changes))
	for _, change := range input.Changes {
		changes[change.CodeSymbolID] = change
	}
	for _, surface := range registry.Surfaces {
		change := changes[surface.CodeSymbolID.String()]
		before, after := bridgeRawDigest(change.BeforeDigest), bridgeRawDigest(change.AfterDigest)
		if input.Manifest.ZeroChange && before == "" && after == "" {
			before, after = bridgeHash("unchanged:"+surface.CodeSymbolID.String()), bridgeHash("unchanged:"+surface.CodeSymbolID.String())
		}
		manifest.Entries = append(manifest.Entries, production.ManifestEntry{SurfaceID: surface.SurfaceID, CodeSymbolID: surface.CodeSymbolID, SemanticOwnerID: surface.SemanticOwnerID,
			BeforeBindingDigest: surface.Binding.BindingDigest, AfterBindingDigest: surface.Binding.BindingDigest,
			BeforeBlobDigest: before, AfterBlobDigest: after,
			BeforeSourcePath: surface.Binding.FileLabel, AfterSourcePath: surface.Binding.FileLabel})
	}
	manifest.Digest = bridgeManifestDigest(manifest)
	return manifest
}

// The detector and independent oracle use different typed path projections.
// This function is the explicit fixture-to-producer projection; after it
// returns, all producer mutations operate on the raw production packet.
func productionPathFromCanonical(path semantic.InferencePathV1, rawReceipts []CouplingReceipt, roots []string) semantic.InferencePathV1 {
	if len(path.Edges) == 0 {
		return semantic.InferencePathV1{}
	}
	var raw CouplingReceipt
	if len(rawReceipts) != 0 {
		raw = rawReceipts[0]
	}
	out := semantic.InferencePathV1{Version: path.Version}
	for _, rawEdge := range path.Edges {
		edge := rawEdge
		edge.InferenceRecord = productionRecordFromCanonical(rawEdge.InferenceRecord)
		if rawEdge.Kind == semantic.InferenceAuthoritativeDeclaration && len(roots) != 0 {
			edge.SourceRoots = []semantic.ID{bridgeID(firstString(roots))}
		}
		switch rawEdge.Kind {
		case semantic.InferenceAuthoritativeDeclaration:
			edge.SubjectID, edge.ObjectID = bridgeID(raw.SemanticOwnerID), bridgeID(raw.CodeSymbolID)
		case semantic.InferenceDerivedProjection:
			edge.SubjectID, edge.ObjectID = bridgeID(raw.CodeSymbolID), bridgeID(raw.SurfaceID)
		case semantic.InferenceIndependentVerification:
			edge.SubjectID = bridgeID(raw.SurfaceID)
			if len(rawEdge.Evidence) != 0 {
				edge.ObjectID = rawEdge.Evidence[0].ID
			}
		}
		edge.InferenceRecord.SubjectID, edge.InferenceRecord.ObjectID = edge.SubjectID, edge.ObjectID
		out.Edges = append(out.Edges, edge)
	}
	for _, claim := range path.Claims {
		mapped := claim
		mapped.InferenceRecord = productionRecordFromCanonical(claim.InferenceRecord)
		mapped.CanonicalDelta = strings.TrimSpace(claim.CanonicalDelta)
		mapped.DeltaDigest = ""
		if mapped.CanonicalDelta != "" {
			mapped.DeltaDigest = bridgeHash(mapped.CanonicalDelta)
		}
		out.Claims = append(out.Claims, mapped)
	}
	for _, evidence := range path.Evidence {
		mapped := evidence
		mapped.Digest = bridgeRawDigest(evidence.Digest)
		mapped.Before = productionSnapshot(evidence.Before)
		mapped.After = productionSnapshot(evidence.After)
		mapped.Controls = productionControls(evidence.Controls)
		out.Evidence = append(out.Evidence, mapped)
	}
	return out
}

func productionRecordFromCanonical(record semantic.InferenceRecord) semantic.InferenceRecord {
	result := record
	result.Rule.Digest = bridgeRawDigest(record.Rule.Digest)
	result.Before = productionSnapshot(record.Before)
	result.After = productionSnapshot(record.After)
	result.Controls = productionControls(record.Controls)
	result.Evidence = make([]semantic.EvidenceReference, 0, len(record.Evidence))
	for _, ref := range record.Evidence {
		result.Evidence = append(result.Evidence, semantic.EvidenceReference{ID: ref.ID, Digest: bridgeRawDigest(ref.Digest)})
	}
	return result
}

func productionSnapshot(snapshot semantic.SnapshotDigests) semantic.SnapshotDigests {
	return semantic.SnapshotDigests{Source: bridgeRawDigest(snapshot.Source), Semantic: bridgeRawDigest(snapshot.Semantic)}
}

func productionControls(controls semantic.InferenceControls) semantic.InferenceControls {
	result := controls
	result.CatalogDigest = bridgeRawDigest(controls.CatalogDigest)
	result.PolicyDigest = bridgeRawDigest(controls.PolicyDigest)
	result.Profile.Digest = bridgeRawDigest(controls.Profile.Digest)
	return result
}

func productionReceiptsFromCanonical(input Input, registry production.Registry, config production.Config, path semantic.InferencePathV1) []production.CouplingReceipt {
	result := make([]production.CouplingReceipt, 0, len(input.Receipts))
	for _, raw := range input.Receipts {
		surface := registrySurface(registry, raw.SurfaceID)
		pathIDs := productionSelectedPathIDs(path)
		evidenceIDs := productionSelectedEvidence(path, pathIDs)
		evidenceRefs := make([]semantic.EvidenceReference, 0, len(evidenceIDs))
		for _, id := range evidenceIDs {
			evidenceRefs = append(evidenceRefs, semantic.EvidenceReference{ID: id, Digest: productionEvidenceDigest(path, id)})
		}
		var source *production.AuthoritySource
		if raw.AuthoritativeSourceRef != "" {
			source = &production.AuthoritySource{SourceID: bridgeID(firstString(input.Roots)), Path: raw.AuthoritativeSourceRef}
		}
		canonicalDelta := strings.TrimSpace(raw.SemanticDelta)
		deltaDigest := ""
		if canonicalDelta != "" {
			deltaDigest = bridgeHash(canonicalDelta)
		}
		result = append(result, production.CouplingReceipt{Schema: production.ReceiptSchemaV1, ReceiptID: bridgeID(raw.ReceiptID), SurfaceID: bridgeID(raw.SurfaceID), SemanticOwnerID: bridgeID(raw.SemanticOwnerID), CodeSymbolID: bridgeID(raw.CodeSymbolID),
			SourceMapBindingDigest: surface.Binding.BindingDigest, SnapshotDigest: bridgeRawDigest(raw.SnapshotDigest), RegistryDigest: config.RegistryDigest, ToolchainDigest: config.ToolchainDigest, ProfileDigest: config.ProfileDigest,
			BeforeBlobDigest: bridgeRawDigest(rawBeforeBlobDigest(input, raw)), AfterBlobDigest: bridgeRawDigest(rawAfterBlobDigest(input, raw)), BeforeAuthoritySourceDigest: bridgeRawDigest(raw.AuthoritySourceBeforeDigest), AfterAuthoritySourceDigest: bridgeRawDigest(raw.AuthoritySourceAfterDigest),
			BeforeCanonicalSemanticDigest: bridgeRawDigest(raw.BeforeIRDigest), AfterCanonicalSemanticDigest: bridgeRawDigest(raw.AfterIRDigest), ChangeClaim: production.ChangeClaim(raw.ChangeClaim), ReceiptKind: semantic.SemanticChangeKind(raw.ReceiptKind),
			CanonicalDelta: canonicalDelta, DeltaDigest: deltaDigest, AuthoritativeSource: source, OriginPathIDs: pathIDs, InferenceClaimID: bridgeID(raw.ClaimRecordID), EvidenceRefs: evidenceRefs, State: raw.State})
	}
	return result
}

func productionExternalReceiptFromCanonical(input Input, config production.Config) *production.ExternalResourceReceipt {
	if len(input.ResourceReceipts) == 0 {
		return nil
	}
	external := &production.ExternalResourceReceipt{Schema: production.ResourceSchemaV1, SnapshotDigest: config.SnapshotDigest, ProviderDigest: bridgeRawDigest(input.ResourceReceipts[0].ProviderDigest), ObserverDigest: bridgeRawDigest(input.ResourceReceipts[0].ObserverDigest)}
	for _, raw := range input.ResourceReceipts {
		value := raw.Value
		switch raw.Metric {
		case "cpu-core-ns":
			external.CPUWorkUnits = &value
		case "peak-memory-bytes":
			external.PeakMemoryBytes = &value
		case "work-units":
			external.DeterministicWorkUnits = &value
		}
	}
	external.Digest = bridgeExternalDigest(*external)
	return external
}

func productionSelectedPathIDs(path semantic.InferencePathV1) []semantic.ID {
	ids := make([]semantic.ID, 0, 3)
	for _, edge := range path.Edges {
		switch edge.Kind {
		case semantic.InferenceAuthoritativeDeclaration, semantic.InferenceDerivedProjection, semantic.InferenceIndependentVerification:
			ids = append(ids, edge.RecordID)
		}
	}
	return ids
}

func productionSelectedEvidence(path semantic.InferencePathV1, ids []semantic.ID) []semantic.ID {
	byID := make(map[semantic.ID]semantic.InferenceEdge, len(path.Edges))
	for _, edge := range path.Edges {
		byID[edge.RecordID] = edge
	}
	selected := make(map[semantic.ID]struct{})
	for _, id := range ids {
		for _, ref := range byID[id].Evidence {
			selected[ref.ID] = struct{}{}
		}
	}
	result := make([]semantic.ID, 0, len(selected))
	for id := range selected {
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func productionEvidenceDigest(path semantic.InferencePathV1, id semantic.ID) string {
	for _, evidence := range path.Evidence {
		if evidence.ID == id {
			return bridgeRawDigest(evidence.Digest)
		}
	}
	return ""
}

func productionVectorFromResult(result production.Result, input production.Input, authority production.AuthorityContext) productionVector {
	accepted := make([]string, 0, len(result.AcceptedSurfaceIDs))
	for _, id := range result.AcceptedSurfaceIDs {
		accepted = append(accepted, id.String())
	}
	sort.Strings(accepted)
	return productionVector{Schema: result.Schema, Decision: bridgeDecision(result.Status), Reasons: append([]production.Reason(nil), result.Reasons...), Accepted: accepted, Observation: result.Observation, FullSuite: result.FullSuiteRequired, InputDigest: result.InputDigest, ResultDigest: result.Digest, Bindings: productionBindings(input, authority)}
}

func productionBindings(input production.Input, authority production.AuthorityContext) productionBindingVector {
	b := productionBindingVector{AuthoritySchema: authority.Schema, AuthorityRegistryDigest: authority.Registry.Digest, AuthorityToolchainDigest: authority.ToolchainDigest, AuthorityProfileDigest: authority.ProfileDigest, AuthoritySnapshotDigest: authority.SnapshotDigest, AuthorityProviderDigest: authority.ExpectedProviderDigest, AuthorityObserverDigest: authority.ExpectedObserverDigest, AuthorityBaselineDigest: authority.Baseline.Digest, AuthorityExternalReceiptRequired: authority.ExternalReceiptRequired, PacketRegistryDigest: input.Registry.Digest, ConfigRegistryDigest: input.Config.RegistryDigest, ConfigToolchainDigest: input.Config.ToolchainDigest, ConfigProfileDigest: input.Config.ProfileDigest, ConfigSnapshotDigest: input.Config.SnapshotDigest, ExpectedProviderDigest: input.Config.ExpectedProviderDigest, ExpectedObserverDigest: input.Config.ExpectedObserverDigest, BaselineDigest: input.Config.Baseline.Digest, ManifestComplete: input.Manifest.Complete, ManifestZeroChange: input.Manifest.ZeroChange, ManifestRegistryDigest: input.Manifest.RegistryDigest, ManifestToolchainDigest: input.Manifest.ToolchainDigest, ManifestProfileDigest: input.Manifest.ProfileDigest, ManifestBeforeSnapshotDigest: input.Manifest.BeforeSnapshotDigest, ManifestAfterSnapshotDigest: input.Manifest.AfterSnapshotDigest, ManifestDigest: input.Manifest.Digest, PathDigest: bridgeHash(input.InferencePath.Canonical())}
	for _, surface := range input.Registry.Surfaces {
		b.Surfaces = append(b.Surfaces, productionSurfaceBinding{SurfaceID: surface.SurfaceID.String(), CodeSymbolID: surface.CodeSymbolID.String(), SemanticOwnerID: surface.SemanticOwnerID.String(), SourceMapID: surface.Binding.SourceMapID.String(), BindingDigest: surface.Binding.BindingDigest})
	}
	for _, entry := range input.Manifest.Entries {
		b.ManifestEntries = append(b.ManifestEntries, productionManifestBinding{SurfaceID: entry.SurfaceID.String(), CodeSymbolID: entry.CodeSymbolID.String(), SemanticOwnerID: entry.SemanticOwnerID.String(), BeforeBindingDigest: entry.BeforeBindingDigest, AfterBindingDigest: entry.AfterBindingDigest, BeforeBlobDigest: entry.BeforeBlobDigest, AfterBlobDigest: entry.AfterBlobDigest})
	}
	for _, receipt := range input.Receipts {
		r := productionReceiptBinding{Schema: receipt.Schema, ReceiptID: receipt.ReceiptID.String(), SurfaceID: receipt.SurfaceID.String(), SemanticOwnerID: receipt.SemanticOwnerID.String(), CodeSymbolID: receipt.CodeSymbolID.String(), SourceMapBindingDigest: receipt.SourceMapBindingDigest, SnapshotDigest: receipt.SnapshotDigest, RegistryDigest: receipt.RegistryDigest, ToolchainDigest: receipt.ToolchainDigest, ProfileDigest: receipt.ProfileDigest, BeforeBlobDigest: receipt.BeforeBlobDigest, AfterBlobDigest: receipt.AfterBlobDigest, BeforeAuthoritySourceDigest: receipt.BeforeAuthoritySourceDigest, AfterAuthoritySourceDigest: receipt.AfterAuthoritySourceDigest, BeforeSemanticDigest: receipt.BeforeCanonicalSemanticDigest, AfterSemanticDigest: receipt.AfterCanonicalSemanticDigest, ChangeClaim: receipt.ChangeClaim, ReceiptKind: receipt.ReceiptKind, CanonicalDelta: receipt.CanonicalDelta, DeltaDigest: receipt.DeltaDigest, ClaimID: receipt.InferenceClaimID.String(), State: receipt.State}
		if receipt.AuthoritativeSource != nil {
			r.AuthoritySourceID, r.AuthoritySourcePath = receipt.AuthoritativeSource.SourceID.String(), receipt.AuthoritativeSource.Path
		}
		for _, id := range receipt.OriginPathIDs {
			r.OriginPathIDs = append(r.OriginPathIDs, id.String())
		}
		for _, ref := range receipt.EvidenceRefs {
			r.EvidenceIDs = append(r.EvidenceIDs, ref.ID.String())
			r.EvidenceDigests = append(r.EvidenceDigests, ref.Digest)
		}
		b.Receipts = append(b.Receipts, r)
	}
	if input.ExternalReceipt != nil {
		b.ExternalReceiptDigest = input.ExternalReceipt.Digest
		r := input.ExternalReceipt
		b.ExternalReceipt = productionExternalBinding{Schema: r.Schema, SnapshotDigest: r.SnapshotDigest, ProviderDigest: r.ProviderDigest, ObserverDigest: r.ObserverDigest, CPUWorkUnits: r.CPUWorkUnits, PeakMemoryBytes: r.PeakMemoryBytes, DeterministicWorkUnits: r.DeterministicWorkUnits, Digest: r.Digest}
	}
	return b
}

func bridgeDecision(status production.Status) Decision {
	switch status {
	case production.StatusPass:
		return DecisionPass
	case production.StatusFailClosed:
		return DecisionFailClosed
	default:
		return DecisionUnknown
	}
}

func bridgeBindingDigestValues(surface, code, owner, sourceMap string) string {
	var b strings.Builder
	bridgeField(&b, surface)
	bridgeField(&b, code)
	bridgeField(&b, owner)
	bridgeField(&b, sourceMap)
	return bridgeHash(b.String())
}
func bridgeRegistryDigest(registry production.Registry) string {
	surfaces := append([]production.Surface(nil), registry.Surfaces...)
	sort.Slice(surfaces, func(i, j int) bool { return surfaces[i].SurfaceID < surfaces[j].SurfaceID })
	var b strings.Builder
	bridgeField(&b, production.RegistrySchemaV1)
	for _, s := range surfaces {
		bridgeField(&b, s.SurfaceID.String())
		bridgeField(&b, s.CodeSymbolID.String())
		bridgeField(&b, s.SemanticOwnerID.String())
		bridgeField(&b, s.Binding.SourceMapID.String())
		bridgeField(&b, s.Binding.BindingDigest)
	}
	return bridgeHash(b.String())
}
func bridgeManifestDigest(manifest production.ChangeManifest) string {
	entries := append([]production.ManifestEntry(nil), manifest.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].SurfaceID < entries[j].SurfaceID })
	var b strings.Builder
	bridgeField(&b, production.ManifestSchemaV1)
	bridgeField(&b, strconv.FormatBool(manifest.Complete))
	bridgeField(&b, strconv.FormatBool(manifest.ZeroChange))
	bridgeField(&b, manifest.RegistryDigest)
	bridgeField(&b, manifest.ToolchainDigest)
	bridgeField(&b, manifest.ProfileDigest)
	bridgeField(&b, manifest.BeforeSnapshotDigest)
	bridgeField(&b, manifest.AfterSnapshotDigest)
	for _, e := range entries {
		bridgeField(&b, e.SurfaceID.String())
		bridgeField(&b, e.CodeSymbolID.String())
		bridgeField(&b, e.SemanticOwnerID.String())
		bridgeField(&b, e.BeforeBindingDigest)
		bridgeField(&b, e.AfterBindingDigest)
		bridgeField(&b, e.BeforeBlobDigest)
		bridgeField(&b, e.AfterBlobDigest)
	}
	return bridgeHash(b.String())
}
func bridgeBaselineDigest(b production.BaselineConfig) string {
	var s strings.Builder
	bridgeField(&s, production.BaselineSchemaV1)
	bridgeField(&s, strconv.FormatBool(b.FullSuiteRequired))
	return bridgeHash(s.String())
}
func bridgeExternalDigest(r production.ExternalResourceReceipt) string {
	var b strings.Builder
	bridgeField(&b, production.ResourceSchemaV1)
	bridgeField(&b, r.SnapshotDigest)
	bridgeField(&b, r.ProviderDigest)
	bridgeField(&b, r.ObserverDigest)
	if r.CPUWorkUnits != nil {
		bridgeField(&b, "cpu_work_units")
		bridgeField(&b, strconv.FormatUint(*r.CPUWorkUnits, 10))
	}
	if r.PeakMemoryBytes != nil {
		bridgeField(&b, "peak_memory_bytes")
		bridgeField(&b, strconv.FormatUint(*r.PeakMemoryBytes, 10))
	}
	if r.DeterministicWorkUnits != nil {
		bridgeField(&b, "deterministic_work_units")
		bridgeField(&b, strconv.FormatUint(*r.DeterministicWorkUnits, 10))
	}
	return bridgeHash(b.String())
}
func bridgeField(b *strings.Builder, value string) {
	b.WriteString(strconv.Itoa(len(value)))
	b.WriteByte(':')
	b.WriteString(value)
	b.WriteByte('|')
}
func bridgeHash(value string) string {
	return strings.TrimPrefix(semantic.StableHashString(value), "sha256:")
}
func bridgeRawDigest(value string) string { return strings.TrimPrefix(value, "sha256:") }
func bridgeID(value string) semantic.ID   { return semantic.MustIdentity(value) }
func bridgeSourceMapID(value string) semantic.ID {
	return semantic.MustIdentity("urn:gooo:source-map:" + value)
}
func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
func registrySurface(registry production.Registry, id string) production.Surface {
	for _, s := range registry.Surfaces {
		if s.SurfaceID.String() == id {
			return s
		}
	}
	return production.Surface{}
}

func rawBeforeBlobDigest(input Input, receipt CouplingReceipt) string {
	for _, change := range input.Changes {
		for _, binding := range input.Registry {
			if binding.RegisteredSurfaceID == receipt.SurfaceID && binding.CodeSymbolID == change.CodeSymbolID {
				return change.BeforeDigest
			}
		}
	}
	return digestText("bridge-before-blob:" + receipt.SurfaceID)
}

func rawAfterBlobDigest(input Input, receipt CouplingReceipt) string {
	for _, change := range input.Changes {
		for _, binding := range input.Registry {
			if binding.RegisteredSurfaceID == receipt.SurfaceID && binding.CodeSymbolID == change.CodeSymbolID {
				return change.AfterDigest
			}
		}
	}
	return digestText("bridge-after-blob:" + receipt.SurfaceID)
}

// These tables are literal contract data. They must never call either
// evaluator to derive expected decisions, reasons, counts, resources, or
// digests. They are filled with immutable vectors for every corpus row and
// mutation before this bridge is publishable.

func literalProductionCorpusExpectations() map[string]productionVector {
	return decodeLiteralProductionVectors(literalProducerCorpusGZIP)
}

func decodeLiteralProductionVectors(encoded string) map[string]productionVector {
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic("invalid literal producer corpus")
	}
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		panic("invalid literal producer corpus gzip")
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		panic("invalid literal producer corpus data")
	}
	if err := reader.Close(); err != nil {
		panic("invalid literal producer corpus gzip close")
	}
	var expected map[string]productionVector
	if err := json.Unmarshal(data, &expected); err != nil {
		panic("invalid literal producer corpus JSON")
	}
	return expected
}

const literalProducerCorpusGZIP = `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`

const literalProducerMutationGZIP = `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`

func literalProductionMutationExpectations() map[string]productionVector {
	return decodeLiteralProductionVectors(literalProducerMutationGZIP)
}

func literalResultMutations(base production.Result, input production.Input, authority production.AuthorityContext) []struct {
	name   string
	mutate func(*production.Result)
	want   productionVector
} {
	wants := literalProductionMutationExpectations()
	return []struct {
		name   string
		mutate func(*production.Result)
		want   productionVector
	}{
		{name: "result-wrong-decision", mutate: func(result *production.Result) {
			result.Status = production.StatusFailClosed
			result.Reasons = []production.Reason{{Code: production.ReasonDigestMismatch, Detail: "producer-only mutation"}}
		}, want: wants["result-wrong-decision"]},
		{name: "result-wrong-reason", mutate: func(result *production.Result) {
			result.Reasons = []production.Reason{{Code: production.ReasonDigestMismatch, Detail: "producer-only mutation"}}
		}, want: wants["result-wrong-reason"]},
		{name: "result-missing-accepted-surface", mutate: func(result *production.Result) {
			result.AcceptedSurfaceIDs = nil
		}, want: wants["result-missing-accepted-surface"]},
		{name: "result-extra-accepted-surface", mutate: func(result *production.Result) {
			result.AcceptedSurfaceIDs = []semantic.ID{bridgeID("urn:gooo:surface:billing/pay-order"), bridgeID("urn:gooo:surface:unexpected")}
		}, want: wants["result-extra-accepted-surface"]},
		{name: "result-count-drift", mutate: func(result *production.Result) {
			result.Observation.InferenceRecords.Value++
		}, want: wants["result-count-drift"]},
		{name: "result-resource-drift", mutate: func(result *production.Result) {
			result.Observation.CPU.Value++
		}, want: wants["result-resource-drift"]},
		{name: "result-input-digest-drift", mutate: func(result *production.Result) {
			result.InputDigest = bridgeHash("producer-only-input-digest")
		}, want: wants["result-input-digest-drift"]},
		{name: "result-result-digest-drift", mutate: func(result *production.Result) {
			result.Digest = bridgeHash("producer-only-result-digest")
		}, want: wants["result-result-digest-drift"]},
	}
}
func literalInputMutations(base production.Input, authority production.AuthorityContext) []struct {
	name            string
	input           production.Input
	authorityBefore production.AuthorityContext
	want            productionVector
} {
	wants := literalProductionMutationExpectations()
	mutations := []struct {
		name            string
		input           production.Input
		authorityBefore production.AuthorityContext
		want            productionVector
	}{}
	add := func(name string, mutate func(*production.Input)) {
		input := cloneProductionInput(base)
		mutate(&input)
		mutations = append(mutations, struct {
			name            string
			input           production.Input
			authorityBefore production.AuthorityContext
			want            productionVector
		}{name: name, input: input, authorityBefore: authority, want: wants[name]})
	}
	add("input-stale-receipt", func(input *production.Input) {
		input.Receipts[0].SnapshotDigest = bridgeHash("stale-receipt")
	})
	add("input-missing-receipt", func(input *production.Input) {
		input.Receipts = nil
	})
	add("input-arbitrary-provider", func(input *production.Input) {
		input.ExternalReceipt.ProviderDigest = bridgeHash("arbitrary-provider")
	})
	add("input-arbitrary-observer", func(input *production.Input) {
		input.ExternalReceipt.ObserverDigest = bridgeHash("arbitrary-observer")
	})
	add("input-wrong-path-endpoint", func(input *production.Input) {
		last := len(input.InferencePath.Edges) - 1
		input.InferencePath.Edges[last].ObjectID = bridgeID("urn:gooo:evidence:wrong-endpoint")
		input.InferencePath.Edges[last].InferenceRecord.ObjectID = input.InferencePath.Edges[last].ObjectID
	})
	add("input-omitted-evidence", func(input *production.Input) {
		input.Receipts[0].EvidenceRefs = nil
	})
	add("input-extra-unrelated-evidence", func(input *production.Input) {
		input.Receipts[0].EvidenceRefs = append(input.Receipts[0].EvidenceRefs, semantic.EvidenceReference{ID: bridgeID("urn:gooo:evidence:unrelated"), Digest: bridgeHash("unrelated-evidence")})
	})
	add("input-duplicate-evidence", func(input *production.Input) {
		input.Receipts[0].EvidenceRefs = append(input.Receipts[0].EvidenceRefs, input.Receipts[0].EvidenceRefs[0])
	})
	add("input-reordered-path-id-presentation", func(input *production.Input) {
		reverseProductionIDs(input.Receipts[0].OriginPathIDs)
	})
	add("input-disconnected-selected-edge", func(input *production.Input) {
		edge := input.InferencePath.Edges[1]
		edge.RecordID = bridgeID("urn:gooo:path:disconnected")
		edge.SubjectID, edge.ObjectID = bridgeID("urn:gooo:term:disconnected"), bridgeID("urn:gooo:code:disconnected")
		edge.InferenceRecord.RecordID, edge.InferenceRecord.SubjectID, edge.InferenceRecord.ObjectID = edge.RecordID, edge.SubjectID, edge.ObjectID
		input.InferencePath.Edges = append(input.InferencePath.Edges, edge)
		input.Receipts[0].OriginPathIDs = append(input.Receipts[0].OriginPathIDs, edge.RecordID)
	})
	add("input-forked-selected-path", func(input *production.Input) {
		edge := input.InferencePath.Edges[1]
		edge.RecordID, edge.ObjectID = bridgeID("urn:gooo:path:fork"), bridgeID("urn:gooo:surface:fork")
		edge.InferenceRecord.RecordID, edge.InferenceRecord.ObjectID = edge.RecordID, edge.ObjectID
		input.InferencePath.Edges = append(input.InferencePath.Edges, edge)
		input.Receipts[0].OriginPathIDs = append(input.Receipts[0].OriginPathIDs, edge.RecordID)
	})
	add("input-cyclic-selected-path", func(input *production.Input) {
		edge := input.InferencePath.Edges[1]
		edge.RecordID, edge.SubjectID, edge.ObjectID = bridgeID("urn:gooo:path:cycle"), bridgeID("urn:gooo:evidence:verification"), bridgeID("urn:gooo:code:billing/pay-order")
		edge.InferenceRecord.RecordID, edge.InferenceRecord.SubjectID, edge.InferenceRecord.ObjectID = edge.RecordID, edge.SubjectID, edge.ObjectID
		input.InferencePath.Edges = append(input.InferencePath.Edges, edge)
		input.Receipts[0].OriginPathIDs = append(input.Receipts[0].OriginPathIDs, edge.RecordID)
	})
	add("input-wrong-start-end", func(input *production.Input) {
		input.InferencePath.Edges[0].SubjectID = bridgeID("urn:gooo:source:wrong")
		input.InferencePath.Edges[0].InferenceRecord.SubjectID = input.InferencePath.Edges[0].SubjectID
		last := len(input.InferencePath.Edges) - 1
		input.InferencePath.Edges[last].SubjectID = bridgeID("urn:gooo:surface:wrong")
		input.InferencePath.Edges[last].InferenceRecord.SubjectID = input.InferencePath.Edges[last].SubjectID
	})
	add("input-rehash-all-packet-authority", mutateRehashAllPacketAuthority)
	add("input-rehash-all-external-authority", mutateRehashAllExternalAuthority)
	return mutations
}
func detectorAuthorityForCorpus(index int) production.AuthorityContext {
	snapshot := "a96c0e268b59e4cf96b821bf38290a6adc7a67298c1bb56f9bb3cf72b0c4b665"
	if index == 1 || index == 9 {
		snapshot = "c033a3f54b319d78cdced9eed61ce7e99859b08987ea6d4b62208d34ad73c644"
	}
	if index == 10 {
		snapshot = "c033a3f54b319d78cdced9eed61ce7e99859b08987ea6d4b62208d34ad73c644"
	}
	registry := production.Registry{Schema: production.RegistrySchemaV1, Digest: "8338cd4a0c97ab010ccb82150d50c7418cdb1c96c543ae0fad097493a0a773c9", Surfaces: []production.Surface{
		{SurfaceID: bridgeID("urn:gooo:surface:billing/pay-order"), CodeSymbolID: bridgeID("urn:gooo:code:billing/pay-order"), SemanticOwnerID: bridgeID("urn:gooo:owner:billing/pay-order"), Binding: production.SourceMapBinding{SourceMapID: bridgeSourceMapID("sm.billing.pay-order"), BindingDigest: "8215244803bce19bff427e5843a64e920330ab4b4a377c731338cf81df77eac8"}},
		{SurfaceID: bridgeID("urn:gooo:surface:billing/pay-order-helper"), CodeSymbolID: bridgeID("urn:gooo:code:billing/pay-order-helper"), SemanticOwnerID: bridgeID("urn:gooo:owner:billing/pay-order-helper"), Binding: production.SourceMapBinding{SourceMapID: bridgeSourceMapID("sm.billing.pay-order-helper"), BindingDigest: "d0f939d45a374c9a06164b10b4eaba1d66ee91cbda58d80c24dd82779f244a10"}},
	}}
	baseline := production.BaselineConfig{Schema: production.BaselineSchemaV1, FullSuiteRequired: true, Digest: "06fc34747e4cadf50f2b013d08ba817817cff1a0be52b09a6f6cedb44f012f02"}
	return production.AuthorityContext{Schema: production.AuthorityContextSchemaV1, Registry: registry, ToolchainDigest: "d214b2b2087b7acc5e3f8ebb3eea7419adbdfbeb0e8e14833feaaa487e475da7", ProfileDigest: "ddd660840b0bae00220bf57cb12f542373f6eddcde3bc0851a269322fb39fb3b", SnapshotDigest: snapshot, ExpectedProviderDigest: "0397c5f0be6ec940d353c05ccac9039256bb1589be30de89924df412ce5db183", ExpectedObserverDigest: "d3e90cc0c61f4342a180c18abf2403dca3bdc6c2284458f3c3df68c7ffb40eba", Baseline: baseline, ExternalReceiptRequired: true}
}
func cloneProductionInput(input production.Input) production.Input {
	output := input
	output.Registry.Surfaces = append([]production.Surface(nil), input.Registry.Surfaces...)
	output.Manifest.Entries = append([]production.ManifestEntry(nil), input.Manifest.Entries...)
	output.Receipts = append([]production.CouplingReceipt(nil), input.Receipts...)
	for i := range output.Receipts {
		output.Receipts[i].OriginPathIDs = append([]semantic.ID(nil), input.Receipts[i].OriginPathIDs...)
		output.Receipts[i].EvidenceRefs = append([]semantic.EvidenceReference(nil), input.Receipts[i].EvidenceRefs...)
		if input.Receipts[i].AuthoritativeSource != nil {
			source := *input.Receipts[i].AuthoritativeSource
			output.Receipts[i].AuthoritativeSource = &source
		}
	}
	if input.ExternalReceipt != nil {
		external := *input.ExternalReceipt
		if input.ExternalReceipt.CPUWorkUnits != nil {
			value := *input.ExternalReceipt.CPUWorkUnits
			external.CPUWorkUnits = &value
		}
		if input.ExternalReceipt.PeakMemoryBytes != nil {
			value := *input.ExternalReceipt.PeakMemoryBytes
			external.PeakMemoryBytes = &value
		}
		if input.ExternalReceipt.DeterministicWorkUnits != nil {
			value := *input.ExternalReceipt.DeterministicWorkUnits
			external.DeterministicWorkUnits = &value
		}
		output.ExternalReceipt = &external
	}
	output.InferencePath.Edges = append([]semantic.InferenceEdge(nil), input.InferencePath.Edges...)
	output.InferencePath.Claims = append([]semantic.SemanticChangeClaim(nil), input.InferencePath.Claims...)
	output.InferencePath.Evidence = append([]semantic.InferenceEvidence(nil), input.InferencePath.Evidence...)
	for i := range output.InferencePath.Edges {
		output.InferencePath.Edges[i].SourceRoots = append([]semantic.ID(nil), input.InferencePath.Edges[i].SourceRoots...)
		output.InferencePath.Edges[i].Evidence = append([]semantic.EvidenceReference(nil), input.InferencePath.Edges[i].Evidence...)
	}
	for i := range output.InferencePath.Claims {
		output.InferencePath.Claims[i].Evidence = append([]semantic.EvidenceReference(nil), input.InferencePath.Claims[i].Evidence...)
	}
	for i := range output.InferencePath.Evidence {
		output.InferencePath.Evidence[i].Controls = input.InferencePath.Evidence[i].Controls
	}
	return output
}

func reverseProductionIDs(values []semantic.ID) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func mutateRehashAllPacketAuthority(input *production.Input) {
	surface := &input.Registry.Surfaces[0]
	surface.CodeSymbolID = bridgeID("urn:gooo:code:billing/pay-order-rehashed")
	surface.Binding.BindingDigest = bridgeBindingDigestValues(surface.SurfaceID.String(), surface.CodeSymbolID.String(), surface.SemanticOwnerID.String(), surface.Binding.SourceMapID.String())
	input.Registry.Digest = bridgeRegistryDigest(input.Registry)
	input.Config.RegistryDigest = input.Registry.Digest
	input.Manifest.RegistryDigest = input.Registry.Digest
	for i := range input.Manifest.Entries {
		if input.Manifest.Entries[i].SurfaceID == surface.SurfaceID {
			input.Manifest.Entries[i].CodeSymbolID = surface.CodeSymbolID
			input.Manifest.Entries[i].AfterBindingDigest = surface.Binding.BindingDigest
			input.Manifest.Entries[i].BeforeBindingDigest = surface.Binding.BindingDigest
		}
	}
	input.Manifest.Digest = bridgeManifestDigest(input.Manifest)
	for i := range input.Receipts {
		if input.Receipts[i].SurfaceID == surface.SurfaceID {
			input.Receipts[i].CodeSymbolID = surface.CodeSymbolID
			input.Receipts[i].SourceMapBindingDigest = surface.Binding.BindingDigest
			input.Receipts[i].RegistryDigest = input.Registry.Digest
		}
	}
	for i := range input.InferencePath.Edges {
		edge := &input.InferencePath.Edges[i]
		if edge.Kind == semantic.InferenceDerivedProjection {
			edge.SubjectID = surface.CodeSymbolID
			edge.InferenceRecord.SubjectID = surface.CodeSymbolID
		}
	}
}

func mutateRehashAllExternalAuthority(input *production.Input) {
	input.Config.ExpectedProviderDigest = bridgeHash("rehashed-provider-authority")
	if input.ExternalReceipt != nil {
		input.ExternalReceipt.ProviderDigest = input.Config.ExpectedProviderDigest
		input.ExternalReceipt.Digest = bridgeExternalDigest(*input.ExternalReceipt)
	}
}
