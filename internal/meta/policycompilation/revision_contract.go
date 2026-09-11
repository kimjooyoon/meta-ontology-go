package policycompilation

import (
	_ "embed"
	"errors"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

//go:embed revision-contract.gooo
var policyRevisionOperationSource []byte

// PolicyRevisionOperationBinding describes a source-owned native role contract.
// It does not authorize execution, repository writes, or policy adoption.
type PolicyRevisionOperationBinding struct {
	SourceDigest      string `json:"source_digest"`
	SemanticDigest    string `json:"semantic_digest"`
	ActivityID        string `json:"activity_id"`
	SourceEntityID    string `json:"source_entity_id"`
	RevisionEntityID  string `json:"revision_entity_id"`
	ProposalEntityID  string `json:"proposal_entity_id"`
	UsedSource        bool   `json:"used_source"`
	UsedRevision      bool   `json:"used_revision"`
	GeneratedProposal bool   `json:"generated_proposal"`
}

// PolicyRevisionNativeBinding binds the existing native proposal primitive to
// its released Gooo source/request/proposal relations. It does not register an
// operation in the common selector or make a new executor available.
func PolicyRevisionNativeBinding() (PolicyRevisionOperationBinding, error) {
	return bindPolicyRevisionOperation(policyRevisionOperationSource)
}

func bindPolicyRevisionOperation(source []byte) (PolicyRevisionOperationBinding, error) {
	file, diagnostics := syntax.ParseFile("revision-contract.gooo", string(source))
	if file == nil || diagnostics.HasErrors() {
		return PolicyRevisionOperationBinding{}, errors.New("policy revision operation contract cannot be parsed")
	}
	const namespace = "metapolicyrevision"
	const sourceName = "PolicySource"
	const activityName = "ProposePolicyDecisionRevision"
	revisionName := reflect.TypeFor[PolicyDecisionRevision]().Name()
	proposalName := reflect.TypeFor[PolicyDecisionProposal]().Name()
	if file.Package == nil || file.Namespace == nil || file.Package.Name != namespace || file.Namespace.Name != namespace ||
		len(file.Decls) != 4 || len(file.Bindings) != 0 {
		return PolicyRevisionOperationBinding{}, errors.New("policy revision operation headers or declaration boundary differ from native ABI")
	}
	entities := make(map[string]*syntax.EntityDecl, 3)
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Decls {
		switch current := declaration.(type) {
		case *syntax.EntityDecl:
			if current == nil || current.FieldsPresent || len(current.Fields) != 0 || entities[current.Name] != nil {
				return PolicyRevisionOperationBinding{}, errors.New("policy revision operation requires unique opaque native entities")
			}
			entities[current.Name] = current
		case *syntax.ActivityDecl:
			if current == nil || activity != nil {
				return PolicyRevisionOperationBinding{}, errors.New("policy revision operation requires one native activity")
			}
			activity = current
		default:
			return PolicyRevisionOperationBinding{}, errors.New("policy revision operation has an unsupported declaration")
		}
	}
	if len(entities) != 3 || entities[sourceName] == nil || entities[revisionName] == nil || entities[proposalName] == nil ||
		activity == nil || activity.Name != activityName || len(activity.Inputs) != 2 ||
		activity.Inputs[0].Name != sourceName || activity.Inputs[1].Name != revisionName || activity.Output != proposalName ||
		activity.ValueProgramPresent || activity.ValueProgram != "" {
		return PolicyRevisionOperationBinding{}, errors.New("policy revision operation signature differs from the admitted native ABI")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return PolicyRevisionOperationBinding{}, fmt.Errorf("lower policy revision operation: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return PolicyRevisionOperationBinding{}, fmt.Errorf("validate policy revision operation: %w", err)
	}
	operation, operationOK := ir.Graph.NodeByName(ir.Namespace, activityName)
	input, inputOK := ir.Graph.NodeByName(ir.Namespace, sourceName)
	revision, revisionOK := ir.Graph.NodeByName(ir.Namespace, revisionName)
	proposal, proposalOK := ir.Graph.NodeByName(ir.Namespace, proposalName)
	if !operationOK || !inputOK || !revisionOK || !proposalOK || operation.Kind != semantic.Activity ||
		input.Kind != semantic.Entity || revision.Kind != semantic.Entity || proposal.Kind != semantic.Entity ||
		input.ID.String() != entities[sourceName].ID || revision.ID.String() != entities[revisionName].ID ||
		proposal.ID.String() != entities[proposalName].ID || len(ir.Graph.Nodes()) != 4 || len(ir.Graph.AllFacts()) != 3 {
		return PolicyRevisionOperationBinding{}, errors.New("policy revision operation semantic graph differs from native roles")
	}
	binding := PolicyRevisionOperationBinding{
		SourceDigest: DigestBytes(source), SemanticDigest: "sha256:" + ir.StableHash(),
		ActivityID: operation.ID.String(), SourceEntityID: input.ID.String(),
		RevisionEntityID: revision.ID.String(), ProposalEntityID: proposal.ID.String(),
		UsedSource:        ir.Graph.HasFact(semantic.FactKey{Subject: operation.ID, Predicate: semantic.Used, Object: input.ID}),
		UsedRevision:      ir.Graph.HasFact(semantic.FactKey{Subject: operation.ID, Predicate: semantic.Used, Object: revision.ID}),
		GeneratedProposal: ir.Graph.HasFact(semantic.FactKey{Subject: proposal.ID, Predicate: semantic.WasGeneratedBy, Object: operation.ID}),
	}
	if !binding.UsedSource || !binding.UsedRevision || !binding.GeneratedProposal ||
		!ValidDigest(binding.SourceDigest) || !ValidDigest(binding.SemanticDigest) {
		return PolicyRevisionOperationBinding{}, errors.New("policy revision operation has incomplete source-owned relations")
	}
	return binding, nil
}
