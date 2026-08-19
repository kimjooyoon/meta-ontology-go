package selectiveci

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	productionsci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"reflect"
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
