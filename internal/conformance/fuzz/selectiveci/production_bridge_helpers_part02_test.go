package selectiveci

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
	productionsci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// directWorkID independently implements SHA256(snapshotDigest || obligationID || pathStableID || policyDigest).
func directWorkID(snapshotDigest, obligationID, pathStableID, policyDigest string) string {
	sum := sha256.Sum256([]byte(snapshotDigest + obligationID + pathStableID + policyDigest))
	return hex.EncodeToString(sum[:])
}
func directReceiptFor(commandID, snapshot string, cpu, memory uint64) productionsci.Receipt {
	return productionsci.Receipt{CommandID: commandID, SnapshotDigest: snapshot, Envelope: resourceenvelope.Envelope{SchemaVersion: resourceenvelope.SchemaVersion, RunnerImageDigest: directDigest("runner"), AllocatedCPUCount: 1, WarmupCount: 1, SampleCount: 5, Limits: resourceenvelope.Limits{CPUCoreNS: cpu, PeakRSSBytes: memory, ReadBytes: 1, WriteBytes: 1}, Samples: []resourceenvelope.Sample{{CPUCoreNS: 0, WallNS: 1}, {CPUCoreNS: 1, WallNS: 1}, {CPUCoreNS: 2, WallNS: 1}, {CPUCoreNS: 3, WallNS: 1}, {CPUCoreNS: 4, WallNS: 1}, {CPUCoreNS: 5, WallNS: 1}}}}
}
func directProvenance(commandID string) productionsci.ProvenancePath {
	subject := semantic.MustIdentity("urn:selectiveci:path/" + commandID)
	object := semantic.MustIdentity("urn:selectiveci:path/result/" + commandID)
	record := semantic.MustIdentity("urn:selectiveci:record/" + commandID)
	evidenceID := semantic.MustIdentity("urn:selectiveci:evidence/" + commandID)
	ruleID := semantic.MustIdentity("urn:selectiveci:rule/v1")
	before := semantic.SnapshotDigests{Source: directDigest("before-" + commandID)}
	after := semantic.SnapshotDigests{Source: directDigest("after-" + commandID)}
	controls := semantic.InferenceControls{}
	edge := semantic.InferenceEdge{InferenceRecord: semantic.InferenceRecord{RecordID: record, SubjectID: subject, ObjectID: object, Rule: semantic.RuleBinding{ID: ruleID, Version: "1", Digest: directDigest("rule")}, Phase: semantic.PhasePlacement{Phase: semantic.PhaseDeclaration, Ordinal: 1}, Before: before, After: after, Authority: semantic.AuthorityBinding{Layer: semantic.AuthoritySource, Effect: semantic.AuthorityDeclare}, Evidence: []semantic.EvidenceReference{{ID: evidenceID, Digest: directDigest("evidence-" + commandID)}}, Controls: controls}, Kind: semantic.InferenceAuthoritativeDeclaration, SourceRoots: []semantic.ID{semantic.MustIdentity("urn:selectiveci:source/" + commandID)}}
	evidence := semantic.InferenceEvidence{ID: evidenceID, Digest: directDigest("evidence-" + commandID), Before: before, After: after, Controls: controls, SourceBacked: true}
	return productionsci.ProvenancePath{CommandID: commandID, Path: semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: []semantic.InferenceEdge{edge}, Evidence: []semantic.InferenceEvidence{evidence}}, Requirement: productionsci.PathRequirement{PathID: semantic.MustIdentity("urn:selectiveci:path-id/" + commandID).String(), RecordIDs: []string{record.String()}, ExpectedKinds: []string{string(semantic.InferenceAuthoritativeDeclaration)}, StartID: subject.String(), EndID: object.String()}}
}
