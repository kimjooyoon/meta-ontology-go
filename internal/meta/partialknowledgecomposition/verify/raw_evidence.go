package verify

import (
	"encoding/json"
	"fmt"
	"slices"
)

const rawEvidenceProvider = "ci-partial-knowledge-observer"

func parseRawEvidence(input Input, model sourceModel) (RawEvidenceReceipt, error) {
	if len(input.RawEvidence) == 0 {
		return RawEvidenceReceipt{}, fmt.Errorf("independent verifier requires raw observation evidence")
	}
	var receipt RawEvidenceReceipt
	if err := json.Unmarshal(input.RawEvidence, &receipt); err != nil {
		return RawEvidenceReceipt{}, fmt.Errorf("decode raw observation evidence independently: %w", err)
	}
	if receipt.Digest == "" || receipt.Digest != rawEvidenceReceiptDigest(receipt) {
		return RawEvidenceReceipt{}, fmt.Errorf("independent raw observation evidence digest is invalid")
	}
	expectedSourceDigest := digestBytes(input.Source)
	if receipt.Schema != "gooo/partial-knowledge/raw-evidence/v3" || receipt.Repository != input.Repository || receipt.HeadSHA != input.HeadSHA || receipt.SourcePath != input.SourcePath || receipt.SourceDigest != expectedSourceDigest || receipt.SemanticIRDigest != model.SemanticIRDigest || receipt.SourceCases != len(model.Cases) || receipt.SourceCasesTotal != 5 || receipt.Provider != rawEvidenceProvider || len(receipt.Cases) != 5 {
		return RawEvidenceReceipt{}, fmt.Errorf("independent raw evidence identity is not bound")
	}
	if err := validateWorkspaceEvidence(receipt.Workspace); err != nil {
		return RawEvidenceReceipt{}, err
	}
	if err := validateAuthorityEvidence(receipt.Authority); err != nil {
		return RawEvidenceReceipt{}, err
	}
	for index, recipe := range model.Cases {
		observed := receipt.Cases[index]
		if observed.ID != recipe.ID || observed.SourceActivity != recipe.SourceActivity || observed.SourceActivityID != recipe.SourceActivityID || observed.Producer != recipe.Producer || observed.Consumer != recipe.Consumer || observed.MetaOperation != recipe.MetaOperation || observed.ProofChoice != recipe.ProofChoice {
			return RawEvidenceReceipt{}, fmt.Errorf("independent raw evidence case %d is not bound to source", index+1)
		}
		if err := validateEvidence(recipe.Left, observed.Left, receipt); err != nil {
			return RawEvidenceReceipt{}, fmt.Errorf("independent raw evidence case %q left: %w", recipe.ID, err)
		}
		if err := validateEvidence(recipe.Right, observed.Right, receipt); err != nil {
			return RawEvidenceReceipt{}, fmt.Errorf("independent raw evidence case %q right: %w", recipe.ID, err)
		}
	}
	return receipt, nil
}

func validateWorkspaceEvidence(workspace WorkspaceObservation) error {
	if workspace.Before.Digest != snapshotDigest(workspace.Before) || workspace.After.Digest != snapshotDigest(workspace.After) {
		return fmt.Errorf("independent workspace snapshot digest is invalid")
	}
	changed := changedSnapshotPaths(workspace.Before, workspace.After)
	if !slices.Equal(changed, workspace.ChangedPaths) || workspace.RepositoryWrites != len(changed) || workspace.Stage != "CI_OBSERVATION" || workspace.Step != "SNAPSHOT_TRACKED_AND_UNTRACKED" || workspace.Reason == "" || workspace.EvidenceDigest != workspaceEvidenceDigest(workspace) {
		return fmt.Errorf("independent workspace observation is not a verified pre/post snapshot")
	}
	return nil
}

func validateAuthorityEvidence(authority CapabilityObservation) error {
	if authority.Name != "promotion-permission" || authority.Available || authority.State != "UNKNOWN" || authority.Resolution != "LOWER_RESOLUTION" || authority.Stage != "CAPABILITY_OBSERVATION" || authority.Step != "CHECK_PROMOTION_PERMISSION" || authority.Reason == "" || authority.EvidenceDigest != capabilityEvidenceDigest(authority) {
		return fmt.Errorf("independent authority evidence is not explicitly unknown")
	}
	return nil
}

func validateEvidence(recipe RecipeOperand, observed Evidence, receipt RawEvidenceReceipt) error {
	if observed.Operation != recipe.Operation || observed.Required != recipe.Required || observed.Stage != "CI_OBSERVATION" || observed.Step != "OBSERVE_RECIPE" || observed.Reason == "" || observed.Provenance.Provider != rawEvidenceProvider || observed.Provenance.SourcePath != receipt.SourcePath || observed.Provenance.SourceDigest != receipt.SourceDigest || observed.Provenance.SemanticIRDigest != receipt.SemanticIRDigest || observed.Provenance.WorkspaceSnapshotDigest != receipt.Workspace.EvidenceDigest || observed.EvidenceDigest != evidenceDigest(observed) {
		return fmt.Errorf("independent observation provenance or digest is invalid")
	}
	switch recipe.ObservationRecipe {
	case "exact":
		if !observed.ObservedAvailable || observed.Observed != observed.Required || observed.Dependency != nil || observed.InvariantEvidence != "" {
			return fmt.Errorf("independent exact recipe did not produce an exact observation")
		}
	case "missing":
		if observed.ObservedAvailable || observed.Observed != "" || observed.Dependency != nil || observed.InvariantEvidence != "" {
			return fmt.Errorf("independent missing recipe manufactured an observation")
		}
	case "dependency":
		if observed.ObservedAvailable || observed.Observed != "" || observed.InvariantEvidence != "" || observed.Dependency == nil || observed.Dependency.ClaimID != "upstream/"+recipe.DependencyRecipe {
			return fmt.Errorf("independent dependency recipe did not carry upstream claim")
		}
		if err := validateUpstreamClaim(*observed.Dependency, receipt, recipe.DependencyRecipe); err != nil {
			return err
		}
	case "invariant":
		if receipt.Workspace.RepositoryWrites == 0 {
			if !observed.ObservedAvailable || observed.Observed != observed.Required || observed.InvariantEvidence != recipe.InvariantCapability || observed.Dependency != nil {
				return fmt.Errorf("independent invariant recipe did not use unchanged snapshot")
			}
		} else if observed.ObservedAvailable || observed.Observed != "" || observed.InvariantEvidence != "" || observed.Dependency != nil {
			return fmt.Errorf("independent invariant recipe ignored changed snapshot")
		}
	default:
		return fmt.Errorf("independent unsupported observation recipe %q", recipe.ObservationRecipe)
	}
	return nil
}

func validateUpstreamClaim(claim UpstreamClaim, receipt RawEvidenceReceipt, target string) error {
	if claim.Proposition == "" || claim.PropositionDigest != digestValue(claim.Proposition) || claim.Predicate != "upstream-evidence-available" || (claim.State != "OPEN" && claim.State != "UNKNOWN") || claim.Resolution != "LOWER_RESOLUTION" || claim.Stage != "UPSTREAM_OBSERVATION" || claim.Step != "OBSERVE_DEPENDENCY_CLAIM" || claim.Reason == "" || claim.EvidenceDigest != upstreamClaimEvidenceDigest(claim) || claim.RawSourceDigest != receipt.SourceDigest || claim.SemanticDigest != receipt.SemanticIRDigest || claim.WorkspaceSnapshotDigest != receipt.Workspace.EvidenceDigest || claim.TargetOperation != "bind-"+target || claim.TargetOutput != "ObservationReceipt" {
		return fmt.Errorf("independent upstream claim lifecycle is not bound")
	}
	return nil
}

func changedSnapshotPaths(before, after Snapshot) []string {
	beforeSet := snapshotEntries(before)
	afterSet := snapshotEntries(after)
	values := make(map[string]struct{})
	for path := range beforeSet {
		if _, ok := afterSet[path]; !ok {
			values[path] = struct{}{}
		}
	}
	for path := range afterSet {
		if _, ok := beforeSet[path]; !ok {
			values[path] = struct{}{}
		}
	}
	result := make([]string, 0, len(values))
	for path := range values {
		result = append(result, path)
	}
	slices.Sort(result)
	return result
}

func snapshotEntries(snapshot Snapshot) map[string]struct{} {
	values := make(map[string]struct{}, len(snapshot.Tracked)+len(snapshot.Untracked)+len(snapshot.Status))
	for _, value := range snapshot.Tracked {
		values["tracked:"+value] = struct{}{}
	}
	for _, value := range snapshot.Untracked {
		values["untracked:"+value] = struct{}{}
	}
	for _, value := range snapshot.Status {
		values["status:"+value] = struct{}{}
	}
	return values
}
