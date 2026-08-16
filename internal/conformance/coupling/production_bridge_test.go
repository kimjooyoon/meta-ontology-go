//go:build detector_bridge

package coupling

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const detectorPreparationHead = "1d80ecee63fecea5a99cb3575cdbcc0f45b682d0"

type productionVector struct {
	Decision    Decision
	Reason      Reason
	Accepted    []string
	Observation production.ObservationVector
	FullSuite   bool
}

func TestProductionBridgePreparationHead(t *testing.T) {
	for _, row := range testCorpus()[:4] {
		want := expectedProductionVector(row.Expected)
		oracle := Evaluate(row.Input)
		if got := oracleProductionVector(oracle); !reflect.DeepEqual(got, want) {
			t.Fatalf("%s oracle vector=%+v want=%+v", row.Name, got, want)
		}
		producer := production.Evaluate(detectorInputFromCanonical(row.Input))
		if got := productionVectorFromResult(producer); !reflect.DeepEqual(got, want) {
			t.Fatalf("%s producer vector=%+v want=%+v detector_head=%s", row.Name, got, want, detectorPreparationHead)
		}
	}
}

func TestProductionBridgeRejectsProducerOnlyResultMutations(t *testing.T) {
	row := testCorpus()[0]
	want := expectedProductionVector(row.Expected)
	base := production.Evaluate(detectorInputFromCanonical(row.Input))
	mutations := []struct {
		name   string
		mutate func(*production.Result)
	}{
		{name: "wrong-decision", mutate: func(result *production.Result) {
			result.Status = production.StatusFailClosed
			result.Reasons = []production.Reason{{Code: production.ReasonDigestMismatch, Detail: "producer-only mutation"}}
		}},
		{name: "wrong-reason", mutate: func(result *production.Result) {
			result.Reasons = []production.Reason{{Code: production.ReasonDigestMismatch, Detail: "producer-only mutation"}}
		}},
		{name: "missing-accepted-surface", mutate: func(result *production.Result) { result.AcceptedSurfaceIDs = nil }},
		{name: "extra-accepted-surface", mutate: func(result *production.Result) {
			result.AcceptedSurfaceIDs = append(result.AcceptedSurfaceIDs, semantic.MustIdentity("urn:gooo:surface:unexpected"))
		}},
		{name: "count-drift", mutate: func(result *production.Result) { result.Observation.InferenceRecords.Value++ }},
		{name: "resource-drift", mutate: func(result *production.Result) { result.Observation.CPU.Value++ }},
	}
	for _, mutation := range mutations {
		got := base
		mutation.mutate(&got)
		if reflect.DeepEqual(productionVectorFromResult(got), want) {
			t.Errorf("producer-only %s mutation was not detected", mutation.name)
		}
	}
}

func TestProductionBridgeRejectsProducerInputMutations(t *testing.T) {
	row := testCorpus()[0]
	want := expectedProductionVector(row.Expected)
	mutations := []struct {
		name   string
		mutate func(*production.Input)
	}{
		{name: "stale-receipt", mutate: func(input *production.Input) {
			input.Receipts[0].SnapshotDigest = bridgeHash("stale-receipt")
		}},
		{name: "missing-receipt", mutate: func(input *production.Input) { input.Receipts = nil }},
		{name: "arbitrary-provider", mutate: func(input *production.Input) {
			input.ExternalReceipt.ProviderDigest = bridgeHash("arbitrary-provider")
			input.ExternalReceipt.Digest = bridgeExternalDigest(*input.ExternalReceipt)
		}},
		{name: "arbitrary-observer", mutate: func(input *production.Input) {
			input.ExternalReceipt.ObserverDigest = bridgeHash("arbitrary-observer")
			input.ExternalReceipt.Digest = bridgeExternalDigest(*input.ExternalReceipt)
		}},
		{name: "path-endpoint", mutate: func(input *production.Input) {
			input.InferencePath.Edges[len(input.InferencePath.Edges)-1].ObjectID = semantic.MustIdentity("urn:gooo:receipt:wrong")
		}},
		{name: "evidence-omission", mutate: func(input *production.Input) { input.Receipts[0].EvidenceRefs = nil }},
	}
	for _, mutation := range mutations {
		input := cloneProductionInput(detectorInputFromCanonical(row.Input))
		mutation.mutate(&input)
		got := productionVectorFromResult(production.Evaluate(input))
		if reflect.DeepEqual(got, want) {
			t.Errorf("producer input mutation %s falsely matched the independent vector", mutation.name)
		}
	}
}

func expectedProductionVector(expected FixtureExpectation) productionVector {
	counts := expected.ObservationCounts
	inferenceRecords := counts.PathEdges + counts.PathClaims + counts.PathEvidence
	work := counts.ChangedRegistered + counts.ReceiptRecords + inferenceRecords
	return productionVector{
		Decision: expected.Decision, Reason: expected.Reason,
		Accepted: append([]string(nil), expected.AcceptedSurfaces...),
		Observation: production.ObservationVector{
			ChangedSurfaces:   productionDimension(counts.ChangedRegistered),
			Receipts:          productionDimension(counts.ReceiptRecords),
			InferenceRecords:  productionDimension(inferenceRecords),
			InferencePaths:    productionDimension(counts.ReceiptRecords),
			DeterministicWork: productionDimension(work),
			ResourceWork:      productionDimension(expected.Resources.WorkUnits),
			CPU:               productionDimension(expected.Resources.CPUCoreNS),
			Memory:            productionDimension(expected.Resources.PeakMemoryBytes),
		},
		FullSuite: expected.Decision != DecisionPass,
	}
}

func oracleProductionVector(output Output) productionVector {
	counts := output.ObservationCounts
	inferenceRecords := counts.PathEdges + counts.PathClaims + counts.PathEvidence
	work := counts.ChangedRegistered + counts.ReceiptRecords + inferenceRecords
	return productionVector{
		Decision: output.Decision, Reason: output.Reason,
		Accepted: append([]string(nil), output.AcceptedSurfaces...),
		Observation: production.ObservationVector{
			ChangedSurfaces:   productionDimension(counts.ChangedRegistered),
			Receipts:          productionDimension(counts.ReceiptRecords),
			InferenceRecords:  productionDimension(inferenceRecords),
			InferencePaths:    productionDimension(counts.ReceiptRecords),
			DeterministicWork: productionDimension(work),
			ResourceWork:      productionDimension(output.Resources.WorkUnits),
			CPU:               productionDimension(output.Resources.CPUCoreNS),
			Memory:            productionDimension(output.Resources.PeakMemoryBytes),
		},
		FullSuite: output.Decision != DecisionPass,
	}
}

func productionVectorFromResult(result production.Result) productionVector {
	accepted := make([]string, 0, len(result.AcceptedSurfaceIDs))
	for _, id := range result.AcceptedSurfaceIDs {
		accepted = append(accepted, id.String())
	}
	sort.Strings(accepted)
	reason := ReasonNone
	if result.Status != production.StatusPass {
		reason = bridgeReason(result)
	}
	return productionVector{
		Decision: bridgeDecision(result.Status), Reason: reason, Accepted: accepted,
		Observation: result.Observation, FullSuite: result.FullSuiteRequired,
	}
}

func productionDimension(value uint64) production.CountDimension {
	return production.CountDimension{Known: true, Value: value}
}

func bridgeDecision(status production.Status) Decision {
	switch status {
	case production.StatusPass:
		return DecisionPass
	case production.StatusFailClosed:
		return DecisionFailClosed
	case production.StatusUnknown:
		return DecisionUnknown
	default:
		return DecisionUnknown
	}
}

func bridgeReason(result production.Result) Reason {
	if len(result.Reasons) == 0 {
		return ReasonInputAmbiguous
	}
	switch result.Reasons[0].Code {
	case production.ReasonRequiredInputMissing:
		return ReasonRequiredInputMissing
	case production.ReasonDuplicateReceipt:
		return ReasonDuplicateReceipt
	case production.ReasonOrphanReceipt:
		return ReasonOrphanReceipt
	case production.ReasonStaleInput:
		return ReasonStaleReceipt
	case production.ReasonDigestMismatch:
		return ReasonDigestMismatch
	case production.ReasonSourceMapMismatch:
		return ReasonRegistryBinding
	case production.ReasonContradictoryReceipt:
		return ReasonInvalidDelta
	case production.ReasonDeltaWithoutSource:
		return ReasonDeltaWithoutSource
	case production.ReasonNoDeltaWithoutEquality:
		return ReasonNoDeltaWithoutEquality
	case production.ReasonCandidateOnlyPath:
		return ReasonCandidateAuthority
	case production.ReasonMissingAuthorityPath:
		return ReasonPathMissing
	case production.ReasonMissingVerification, production.ReasonInferencePathMalformed:
		return ReasonPathMalformed
	case production.ReasonExternalReceiptMissing:
		return ReasonResourceUnbound
	default:
		return ReasonPathMalformed
	}
}

func detectorInputFromCanonical(input Input) production.Input {
	registry := production.Registry{Schema: production.RegistrySchemaV1}
	for _, binding := range input.Registry {
		surface := production.Surface{
			SurfaceID:       bridgeID(binding.RegisteredSurfaceID),
			CodeSymbolID:    bridgeID(binding.CodeSymbolID),
			SemanticOwnerID: bridgeID(binding.SemanticOwnerID),
			Binding: production.SourceMapBinding{
				SourceMapID: bridgeID(binding.SourceMapID),
			},
		}
		surface.Binding.BindingDigest = bridgeBindingDigest(surface)
		registry.Surfaces = append(registry.Surfaces, surface)
	}
	registry.Digest = bridgeRegistryDigest(registry)

	profileDigest := bridgeRawDigest(input.Config.Profile.Digest)
	config := production.Config{
		Schema:                 production.ConfigSchemaV1,
		RegistryDigest:         registry.Digest,
		ToolchainDigest:        bridgeRawDigest(input.Config.ToolchainDigest),
		ProfileDigest:          profileDigest,
		SnapshotDigest:         bridgeRawDigest(input.Config.ResourceBinding.SnapshotDigest),
		ExpectedProviderDigest: bridgeRawDigest(input.Config.ResourceBinding.ProviderDigest),
		ExpectedObserverDigest: bridgeRawDigest(input.Config.ResourceBinding.ObserverDigest),
		Baseline: production.BaselineConfig{
			Schema:            production.BaselineSchemaV1,
			FullSuiteRequired: true,
		},
		ExternalReceiptRequired: true,
	}
	config.Baseline.Digest = bridgeBaselineDigest(config.Baseline)

	manifest := production.ChangeManifest{
		Schema:               production.ManifestSchemaV1,
		Complete:             input.Manifest.Complete,
		ZeroChange:           input.Manifest.ZeroChange,
		RegistryDigest:       registry.Digest,
		ToolchainDigest:      config.ToolchainDigest,
		ProfileDigest:        config.ProfileDigest,
		BeforeSnapshotDigest: bridgeRawDigest(input.Manifest.BeforeSnapshotDigest),
		AfterSnapshotDigest:  config.SnapshotDigest,
	}
	changes := make(map[string]CodeChange, len(input.Changes))
	for _, change := range input.Changes {
		changes[change.CodeSymbolID] = change
	}
	for _, binding := range input.Registry {
		change, changed := changes[binding.CodeSymbolID]
		beforeBlob, afterBlob := bridgeHash("unchanged:"+binding.CodeSymbolID), bridgeHash("unchanged:"+binding.CodeSymbolID)
		if changed {
			beforeBlob, afterBlob = bridgeRawDigest(change.BeforeDigest), bridgeRawDigest(change.AfterDigest)
		}
		manifest.Entries = append(manifest.Entries, production.ManifestEntry{
			SurfaceID: bridgeID(binding.RegisteredSurfaceID), CodeSymbolID: bridgeID(binding.CodeSymbolID), SemanticOwnerID: bridgeID(binding.SemanticOwnerID),
			BeforeBindingDigest: registrySurface(registry, binding.RegisteredSurfaceID).Binding.BindingDigest,
			AfterBindingDigest:  registrySurface(registry, binding.RegisteredSurfaceID).Binding.BindingDigest,
			BeforeBlobDigest:    beforeBlob, AfterBlobDigest: afterBlob,
			BeforeSourcePath: "workspace/before.go", AfterSourcePath: "workspace/after.go",
		})
	}
	manifest.Digest = bridgeManifestDigest(manifest)

	return production.Input{
		Schema:          production.InputSchemaV1,
		Config:          config,
		Registry:        registry,
		Manifest:        manifest,
		Receipts:        bridgeReceipts(input, registry, config),
		InferencePath:   bridgePath(input),
		ExternalReceipt: bridgeExternalReceipt(input, config),
		WorkspaceRoot:   "/workspace",
	}
}

func registrySurface(registry production.Registry, id string) production.Surface {
	for _, surface := range registry.Surfaces {
		if surface.SurfaceID.String() == id {
			return surface
		}
	}
	return production.Surface{}
}

func bridgeReceipts(input Input, registry production.Registry, config production.Config) []production.CouplingReceipt {
	if len(input.Receipts) == 0 {
		return nil
	}
	authority, projection, verification := bridgeSelectedPathIDs(input.Path)
	root := bridgeIDOrEmpty(firstString(input.Roots))
	pathEvidence := bridgePathEvidence(input.Path, authority, projection, verification)
	result := make([]production.CouplingReceipt, 0, len(input.Receipts))
	for _, raw := range input.Receipts {
		binding := registrySurface(registry, raw.SurfaceID)
		localBinding := localBindingFor(input.Registry, raw.SurfaceID)
		bindingDigest := bridgeRawDigest(raw.SourceMapBindingDigest)
		if localBinding.BindingDigest == raw.SourceMapBindingDigest {
			bindingDigest = binding.Binding.BindingDigest
		}
		snapshot := bridgeRawDigest(raw.SnapshotDigest)
		if raw.SnapshotDigest == input.Config.ResourceBinding.SnapshotDigest {
			snapshot = config.SnapshotDigest
		}
		registryDigest := bridgeRawDigest(raw.RegistryDigest)
		if raw.RegistryDigest == input.RegistryDigest {
			registryDigest = config.RegistryDigest
		}
		originIDs := []semantic.ID{authority, projection, verification}
		if raw.OriginPathID != verification.String() {
			originIDs = []semantic.ID{bridgeIDOrEmpty(raw.OriginPathID)}
		}
		evidenceRefs := pathEvidence
		if len(raw.EvidenceRefs) != 1 || raw.EvidenceRefs[0] != verificationEvidenceID(input.Path, verification) {
			evidenceRefs = make([]semantic.EvidenceReference, 0, len(raw.EvidenceRefs))
			for _, id := range raw.EvidenceRefs {
				evidenceRefs = append(evidenceRefs, semantic.EvidenceReference{ID: bridgeIDOrEmpty(id), Digest: bridgeEvidenceDigest(input.Path, id)})
			}
		}
		var source *production.AuthoritySource
		if raw.AuthoritativeSourceRef != "" {
			source = &production.AuthoritySource{SourceID: root, Path: raw.AuthoritativeSourceRef}
		}
		result = append(result, production.CouplingReceipt{
			Schema: production.ReceiptSchemaV1, ReceiptID: bridgeIDOrEmpty(raw.ReceiptID), SurfaceID: bridgeIDOrEmpty(raw.SurfaceID),
			SemanticOwnerID: bridgeIDOrEmpty(raw.SemanticOwnerID), CodeSymbolID: bridgeIDOrEmpty(raw.CodeSymbolID), SourceMapBindingDigest: bindingDigest,
			SnapshotDigest: snapshot, RegistryDigest: registryDigest, ToolchainDigest: bridgeRawDigest(raw.ToolchainDigest), ProfileDigest: bridgeRawDigest(raw.ProfileDigest),
			BeforeBlobDigest: bridgeRawDigest(rawBeforeBlob(input, raw)), AfterBlobDigest: bridgeRawDigest(rawAfterBlob(input, raw)),
			BeforeAuthoritySourceDigest: bridgeRawDigest(raw.AuthoritySourceBeforeDigest), AfterAuthoritySourceDigest: bridgeRawDigest(raw.AuthoritySourceAfterDigest),
			BeforeCanonicalSemanticDigest: bridgeRawDigest(raw.BeforeIRDigest), AfterCanonicalSemanticDigest: bridgeRawDigest(raw.AfterIRDigest),
			ChangeClaim: production.ChangeClaim(raw.ChangeClaim), ReceiptKind: semantic.SemanticChangeKind(raw.ReceiptKind),
			CanonicalDelta: raw.SemanticDelta, DeltaDigest: bridgeRawDigest(raw.SemanticDeltaDigest), AuthoritativeSource: source,
			OriginPathIDs: originIDs, InferenceClaimID: bridgeIDOrEmpty(raw.ClaimRecordID), EvidenceRefs: evidenceRefs, State: raw.State,
		})
	}
	return result
}

func rawBeforeBlob(input Input, receipt CouplingReceipt) string {
	for _, change := range input.Changes {
		if binding := localBindingFor(input.Registry, receipt.SurfaceID); binding.CodeSymbolID == change.CodeSymbolID {
			return change.BeforeDigest
		}
	}
	return digestText("bridge-before-blob:" + receipt.SurfaceID)
}

func rawAfterBlob(input Input, receipt CouplingReceipt) string {
	for _, change := range input.Changes {
		if binding := localBindingFor(input.Registry, receipt.SurfaceID); binding.CodeSymbolID == change.CodeSymbolID {
			return change.AfterDigest
		}
	}
	return digestText("bridge-after-blob:" + receipt.SurfaceID)
}

func localBindingFor(bindings []CodeBinding, surface string) CodeBinding {
	for _, binding := range bindings {
		if binding.RegisteredSurfaceID == surface {
			return binding
		}
	}
	return CodeBinding{}
}

func bridgeExternalReceipt(input Input, config production.Config) *production.ExternalResourceReceipt {
	if len(input.ResourceReceipts) == 0 {
		return nil
	}
	external := &production.ExternalResourceReceipt{
		Schema: production.ResourceSchemaV1, SnapshotDigest: config.SnapshotDigest,
		ProviderDigest: config.ExpectedProviderDigest, ObserverDigest: config.ExpectedObserverDigest,
	}
	for _, raw := range input.ResourceReceipts {
		if raw.SnapshotDigest != input.Config.ResourceBinding.SnapshotDigest {
			external.SnapshotDigest = bridgeRawDigest(raw.SnapshotDigest)
		}
		if raw.ProviderDigest != input.Config.ResourceBinding.ProviderDigest {
			external.ProviderDigest = bridgeRawDigest(raw.ProviderDigest)
		}
		if raw.ObserverDigest != input.Config.ResourceBinding.ObserverDigest {
			external.ObserverDigest = bridgeRawDigest(raw.ObserverDigest)
		}
		if !raw.Present {
			continue
		}
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

func bridgePath(path semantic.InferencePathV1, input Input) semantic.InferencePathV1 {
	if len(path.Edges) == 0 && len(path.Claims) == 0 && len(path.Evidence) == 0 {
		return semantic.InferencePathV1{}
	}
	owner, code, surface, receiptID := semantic.ID(""), semantic.ID(""), semantic.ID(""), semantic.ID("")
	if len(input.Receipts) != 0 {
		owner, code, surface, receiptID = bridgeIDOrEmpty(input.Receipts[0].SemanticOwnerID), bridgeIDOrEmpty(input.Receipts[0].CodeSymbolID), bridgeIDOrEmpty(input.Receipts[0].SurfaceID), bridgeIDOrEmpty(input.Receipts[0].ReceiptID)
	}
	root := bridgeIDOrEmpty(firstString(input.Roots))
	out := semantic.InferencePathV1{Version: path.Version}
	for _, raw := range path.Edges {
		edge := raw
		edge.InferenceRecord = bridgeRecord(raw.InferenceRecord)
		edge.AcceptanceReceipt = bridgeIDOrEmpty(raw.AcceptanceReceipt.String())
		edge.SourceRoots = nil
		switch raw.Kind {
		case semantic.InferenceAuthoritativeDeclaration:
			edge.SubjectID, edge.ObjectID = owner, code
			edge.SourceRoots = []semantic.ID{root}
		case semantic.InferenceDerivedProjection:
			edge.SubjectID, edge.ObjectID = code, surface
		case semantic.InferenceIndependentVerification:
			edge.SubjectID, edge.ObjectID = surface, receiptID
		}
		edge.InferenceRecord.SubjectID, edge.InferenceRecord.ObjectID = edge.SubjectID, edge.ObjectID
		out.Edges = append(out.Edges, edge)
	}
	for _, raw := range path.Claims {
		claim := raw
		claim.InferenceRecord = bridgeRecord(raw.InferenceRecord)
		out.Claims = append(out.Claims, claim)
	}
	for _, raw := range path.Evidence {
		evidence := raw
		evidence.Digest = bridgeRawDigest(raw.Digest)
		evidence.Before = bridgeSnapshot(raw.Before)
		evidence.After = bridgeSnapshot(raw.After)
		evidence.Controls = bridgeControls(raw.Controls)
		out.Evidence = append(out.Evidence, evidence)
	}
	return out
}

func bridgeRecord(record semantic.InferenceRecord) semantic.InferenceRecord {
	record.RecordID = bridgeIDOrEmpty(record.RecordID.String())
	record.SubjectID = bridgeIDOrEmpty(record.SubjectID.String())
	record.ObjectID = bridgeIDOrEmpty(record.ObjectID.String())
	record.Rule.ID = bridgeIDOrEmpty(record.Rule.ID.String())
	record.Rule.Digest = bridgeRawDigest(record.Rule.Digest)
	record.Before = bridgeSnapshot(record.Before)
	record.After = bridgeSnapshot(record.After)
	record.Evidence = bridgeEvidenceRefs(record.Evidence)
	record.Controls = bridgeControls(record.Controls)
	return record
}

func bridgeSnapshot(snapshot semantic.SnapshotDigests) semantic.SnapshotDigests {
	return semantic.SnapshotDigests{Source: bridgeRawDigest(snapshot.Source), Semantic: bridgeRawDigest(snapshot.Semantic)}
}

func bridgeControls(controls semantic.InferenceControls) semantic.InferenceControls {
	controls.CatalogDigest = bridgeRawDigest(controls.CatalogDigest)
	controls.PolicyDigest = bridgeRawDigest(controls.PolicyDigest)
	controls.Profile.Digest = bridgeRawDigest(controls.Profile.Digest)
	return controls
}

func bridgeEvidenceRefs(refs []semantic.EvidenceReference) []semantic.EvidenceReference {
	out := make([]semantic.EvidenceReference, 0, len(refs))
	for _, ref := range refs {
		out = append(out, semantic.EvidenceReference{ID: bridgeIDOrEmpty(ref.ID.String()), Digest: bridgeRawDigest(ref.Digest)})
	}
	return out
}

func bridgeSelectedPathIDs(path semantic.InferencePathV1) (semantic.ID, semantic.ID, semantic.ID) {
	var authority, projection, verification semantic.ID
	for _, edge := range path.Edges {
		switch edge.Kind {
		case semantic.InferenceAuthoritativeDeclaration:
			authority = edge.RecordID
		case semantic.InferenceDerivedProjection:
			projection = edge.RecordID
		case semantic.InferenceIndependentVerification:
			verification = edge.RecordID
		}
	}
	return authority, projection, verification
}

func bridgePathEvidence(path semantic.InferencePathV1, ids ...semantic.ID) []semantic.EvidenceReference {
	wanted := make(map[semantic.ID]struct{})
	for _, id := range ids {
		for _, edge := range path.Edges {
			if edge.RecordID == id {
				for _, ref := range edge.Evidence {
					wanted[ref.ID] = struct{}{}
				}
			}
		}
	}
	result := make([]semantic.EvidenceReference, 0, len(wanted))
	for _, evidence := range path.Evidence {
		if _, ok := wanted[evidence.ID]; ok {
			result = append(result, semantic.EvidenceReference{ID: evidence.ID, Digest: bridgeRawDigest(evidence.Digest)})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func verificationEvidenceID(path semantic.InferencePathV1, verification semantic.ID) string {
	for _, edge := range path.Edges {
		if edge.RecordID == verification && len(edge.Evidence) != 0 {
			return edge.Evidence[0].ID.String()
		}
	}
	return ""
}

func bridgeEvidenceDigest(path semantic.InferencePathV1, id string) string {
	for _, evidence := range path.Evidence {
		if evidence.ID.String() == id {
			return bridgeRawDigest(evidence.Digest)
		}
	}
	return bridgeHash("missing-evidence:" + id)
}

func cloneProductionInput(input production.Input) production.Input {
	out := input
	out.Registry.Surfaces = append([]production.Surface(nil), input.Registry.Surfaces...)
	out.Manifest.Entries = append([]production.ManifestEntry(nil), input.Manifest.Entries...)
	out.Receipts = append([]production.CouplingReceipt(nil), input.Receipts...)
	for i := range out.Receipts {
		out.Receipts[i].OriginPathIDs = append([]semantic.ID(nil), input.Receipts[i].OriginPathIDs...)
		out.Receipts[i].EvidenceRefs = append([]semantic.EvidenceReference(nil), input.Receipts[i].EvidenceRefs...)
	}
	out.InferencePath.Edges = append([]semantic.InferenceEdge(nil), input.InferencePath.Edges...)
	for i := range out.InferencePath.Edges {
		out.InferencePath.Edges[i].SourceRoots = append([]semantic.ID(nil), input.InferencePath.Edges[i].SourceRoots...)
		out.InferencePath.Edges[i].Evidence = append([]semantic.EvidenceReference(nil), input.InferencePath.Edges[i].Evidence...)
	}
	out.InferencePath.Claims = append([]semantic.SemanticChangeClaim(nil), input.InferencePath.Claims...)
	for i := range out.InferencePath.Claims {
		out.InferencePath.Claims[i].Evidence = append([]semantic.EvidenceReference(nil), input.InferencePath.Claims[i].Evidence...)
	}
	out.InferencePath.Evidence = append([]semantic.InferenceEvidence(nil), input.InferencePath.Evidence...)
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
		out.ExternalReceipt = &external
	}
	return out
}

func bridgeBindingDigest(surface production.Surface) string {
	var builder strings.Builder
	bridgeField(&builder, surface.SurfaceID.String())
	bridgeField(&builder, surface.CodeSymbolID.String())
	bridgeField(&builder, surface.SemanticOwnerID.String())
	bridgeField(&builder, surface.Binding.SourceMapID.String())
	return bridgeHash(builder.String())
}

func bridgeRegistryDigest(registry production.Registry) string {
	surfaces := append([]production.Surface(nil), registry.Surfaces...)
	sort.Slice(surfaces, func(i, j int) bool { return surfaces[i].SurfaceID < surfaces[j].SurfaceID })
	var builder strings.Builder
	bridgeField(&builder, production.RegistrySchemaV1)
	for _, surface := range surfaces {
		bridgeField(&builder, surface.SurfaceID.String())
		bridgeField(&builder, surface.CodeSymbolID.String())
		bridgeField(&builder, surface.SemanticOwnerID.String())
		bridgeField(&builder, surface.Binding.SourceMapID.String())
		bridgeField(&builder, surface.Binding.BindingDigest)
	}
	return bridgeHash(builder.String())
}

func bridgeManifestDigest(manifest production.ChangeManifest) string {
	entries := append([]production.ManifestEntry(nil), manifest.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].SurfaceID < entries[j].SurfaceID })
	var builder strings.Builder
	bridgeField(&builder, production.ManifestSchemaV1)
	bridgeField(&builder, strconv.FormatBool(manifest.Complete))
	bridgeField(&builder, strconv.FormatBool(manifest.ZeroChange))
	bridgeField(&builder, manifest.RegistryDigest)
	bridgeField(&builder, manifest.ToolchainDigest)
	bridgeField(&builder, manifest.ProfileDigest)
	bridgeField(&builder, manifest.BeforeSnapshotDigest)
	bridgeField(&builder, manifest.AfterSnapshotDigest)
	for _, entry := range entries {
		bridgeField(&builder, entry.SurfaceID.String())
		bridgeField(&builder, entry.CodeSymbolID.String())
		bridgeField(&builder, entry.SemanticOwnerID.String())
		bridgeField(&builder, entry.BeforeBindingDigest)
		bridgeField(&builder, entry.AfterBindingDigest)
		bridgeField(&builder, entry.BeforeBlobDigest)
		bridgeField(&builder, entry.AfterBlobDigest)
	}
	return bridgeHash(builder.String())
}

func bridgeBaselineDigest(baseline production.BaselineConfig) string {
	var builder strings.Builder
	bridgeField(&builder, production.BaselineSchemaV1)
	bridgeField(&builder, strconv.FormatBool(baseline.FullSuiteRequired))
	return bridgeHash(builder.String())
}

func bridgeExternalDigest(receipt production.ExternalResourceReceipt) string {
	var builder strings.Builder
	bridgeField(&builder, production.ResourceSchemaV1)
	bridgeField(&builder, receipt.SnapshotDigest)
	bridgeField(&builder, receipt.ProviderDigest)
	bridgeField(&builder, receipt.ObserverDigest)
	if receipt.CPUWorkUnits != nil {
		bridgeField(&builder, "cpu_work_units")
		bridgeField(&builder, strconv.FormatUint(*receipt.CPUWorkUnits, 10))
	}
	if receipt.PeakMemoryBytes != nil {
		bridgeField(&builder, "peak_memory_bytes")
		bridgeField(&builder, strconv.FormatUint(*receipt.PeakMemoryBytes, 10))
	}
	if receipt.DeterministicWorkUnits != nil {
		bridgeField(&builder, "deterministic_work_units")
		bridgeField(&builder, strconv.FormatUint(*receipt.DeterministicWorkUnits, 10))
	}
	return bridgeHash(builder.String())
}

func bridgeField(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('|')
}

func bridgeHash(value string) string { return semantic.StableHashString(value) }

func bridgeRawDigest(value string) string {
	return strings.TrimPrefix(value, "sha256:")
}

func bridgeID(value string) semantic.ID { return semantic.MustIdentity(value) }

func bridgeIDOrEmpty(value string) semantic.ID {
	if value == "" {
		return ""
	}
	parsed, err := semantic.ParseIdentity(value)
	if err != nil {
		return semantic.ID(value)
	}
	return parsed
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
