package selectiveci

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"

	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
	productionsci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type directFixture struct {
	Input        productionsci.Input
	AnalyzerBase analyzersci.SnapshotInput
	AnalyzerHead analyzersci.SnapshotInput
	OracleToProd map[string]string
	ProdToOracle map[string]string
}

func translateDirect(c Case) (directFixture, error) {
	if c.Name != "direct" || len(c.Graph.Commands) != 1 || len(c.Evidence.Paths) != 1 || len(c.Evidence.Changes) != 1 {
		return directFixture{}, fmt.Errorf("direct fixture shape is not mechanically representable")
	}
	command := c.Graph.Commands[0]
	path := c.Evidence.Paths[0]
	change := c.Evidence.Changes[0]
	if command.ID != "cmd.compile" || path.Owners[0] != command.ID || change.Path != path.Path {
		return directFixture{}, fmt.Errorf("direct fixture identity mapping is incomplete")
	}
	prodID := "urn:selectiveci:command/cmd-compile"
	obligationID := "urn:selectiveci:obligation/cmd-compile"
	baseBlob, headBlob := directDigest("base-blob"), directDigest("head-blob")
	base := productionsci.SnapshotManifest{SchemaVersion: productionsci.ManifestSchemaVersion, Files: []productionsci.SnapshotFile{{Path: "src/main.gooo", BlobDigest: baseBlob, SemanticIDs: []string{prodID}}}}
	head := productionsci.SnapshotManifest{SchemaVersion: productionsci.ManifestSchemaVersion, Files: []productionsci.SnapshotFile{{Path: "src/main.gooo", BlobDigest: headBlob, SemanticIDs: []string{prodID}}}}
	base.Digest, head.Digest = base.ComputedDigest(), head.ComputedDigest()
	registry := productionsci.Registry{SchemaVersion: productionsci.RegistrySchemaVersion, PolicyDigest: directDigest("policy"), Nodes: []impactgraph.Node{{ID: prodID, Kind: impactgraph.NodeKindSemantic}, {ID: obligationID, Kind: impactgraph.NodeKindObligation}}, DependencyEdges: []productionsci.DependencyEdge{}, Obligations: []productionsci.ObligationBinding{{ID: obligationID, Subject: prodID, CommandIDs: []string{prodID}}}, Commands: []productionsci.Command{{ID: prodID, Argv: append([]string(nil), command.Argv...), WorkingDir: ".", CPUWorkUnits: command.CPUUnits, MemoryBytes: command.MemoryCeiling}}, GlobalGuardCommands: []productionsci.Command{}}
	registry.Digest = registry.ComputedDigest()
	input := productionsci.Input{SchemaVersion: productionsci.SchemaVersion, Base: base, Head: head, Registry: registry, CPUCapacity: command.CPUUnits, Receipts: []productionsci.Receipt{directReceiptFor(prodID, head.Digest, command.CPUUnits, command.MemoryCeiling)}, ProvenancePaths: []productionsci.ProvenancePath{directProvenance(prodID)}}
	return directFixture{Input: input, AnalyzerBase: analyzerInput("src/main.gooo", baseBlob, prodID), AnalyzerHead: analyzerInput("src/main.gooo", headBlob, prodID), OracleToProd: map[string]string{command.ID: prodID}, ProdToOracle: map[string]string{prodID: command.ID}}, nil
}

func analyzerInput(path, blob, id string) analyzersci.SnapshotInput {
	binding := semanticbinding.Binding{ID: id, Role: semanticbinding.RoleHandwrittenImpl, PackagePath: "fixture", DeclarationKey: id, Span: semanticbinding.Span{Filename: path, Start: semanticbinding.Position{Line: 1, Column: 1}, End: semanticbinding.Position{Line: 1, Column: 2}}, DirectiveSpan: semanticbinding.Span{Filename: path, Start: semanticbinding.Position{Line: 1, Column: 1}, End: semanticbinding.Position{Line: 1, Column: 2}}}
	binding.Digest, binding.CanonicalDigest = binding.StableHash(), binding.StableHash()
	return analyzersci.SnapshotInput{Sources: []analyzersci.SourceInput{{Path: path, BlobDigest: analyzerDigest(blob), Bindings: []semanticbinding.Binding{binding}}}, SourceMapDigest: analyzerDigest("source-map"), RegistryDigest: analyzerDigest("registry"), RegisteredIDs: []string{id}}
}

func runDirect(fixture directFixture) (productionsci.PlanResult, error) {
	base, err := analyzersci.Build(fixture.AnalyzerBase)
	if err != nil {
		return productionsci.PlanResult{}, err
	}
	head, err := analyzersci.Build(fixture.AnalyzerHead)
	if err != nil {
		return productionsci.PlanResult{}, err
	}
	delta, err := analyzersci.Diff(base, head)
	if err != nil || delta.Status != analyzersci.StatusBound || !reflect.DeepEqual(delta.ChangedIDs, []string{fixture.OracleToProd["cmd.compile"]}) {
		return productionsci.PlanResult{}, fmt.Errorf("analyzer delta=%#v err=%v", delta, err)
	}
	return productionsci.Plan(fixture.Input), nil
}

func directReceiptDigest(receipt directReceipt) string {
	body, _ := json.Marshal(receipt)
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func directDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func analyzerDigest(value string) string { return "sha256:" + directDigest(value) }

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
