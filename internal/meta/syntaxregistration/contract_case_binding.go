package syntaxregistration

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func bindContract() (contractBinding, error) {
	return bindContractSource(contractSource)
}

func bindCaseComposition(ir semantic.IR) error {
	if len(ir.Graph.Nodes()) != 32 {
		return fmt.Errorf("registration contract node inventory is not exact: %d", len(ir.Graph.Nodes()))
	}
	input, inputFound := ir.Graph.NodeByName(ir.Namespace, "CaseIntent")
	output, outputFound := ir.Graph.NodeByName(ir.Namespace, "CanonicalCase")
	activity, activityFound := ir.Graph.NodeByName(ir.Namespace, "ComposeCase")
	if !inputFound || !outputFound || !activityFound || input.Kind != semantic.Entity ||
		output.Kind != semantic.Entity || activity.Kind != semantic.Activity ||
		input.ID.String() != "gooo://syntax-registration/case-intent" ||
		output.ID.String() != "gooo://syntax-registration/canonical-case" ||
		activity.ID.String() != "gooo://syntax-registration/compose-case" || activity.ValueProgram == "" ||
		!ir.Graph.HasFact(semantic.FactKey{Subject: activity.ID, Predicate: semantic.Used, Object: input.ID}) ||
		!ir.Graph.HasFact(semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: activity.ID}) {
		return fmt.Errorf("registration case-composition activity ABI mismatch")
	}
	return nil
}

func bindAppendComposition(ir semantic.IR) error {
	input, inputFound := ir.Graph.NodeByName(ir.Namespace, "AppendObservation")
	output, outputFound := ir.Graph.NodeByName(ir.Namespace, "AppendRecord")
	activity, activityFound := ir.Graph.NodeByName(ir.Namespace, "ComposeRegistrationAppend")
	if !inputFound || !outputFound || !activityFound || input.Kind != semantic.Entity ||
		output.Kind != semantic.Entity || activity.Kind != semantic.Activity ||
		input.ID.String() != "gooo://syntax-registration/append-observation" ||
		output.ID.String() != "gooo://syntax-registration/append-record" ||
		activity.ID.String() != "gooo://syntax-registration/compose-append" || activity.ValueProgram == "" ||
		!ir.Graph.HasFact(semantic.FactKey{Subject: activity.ID, Predicate: semantic.Used, Object: input.ID}) ||
		!ir.Graph.HasFact(semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: activity.ID}) {
		return fmt.Errorf("registration append-composition activity ABI mismatch")
	}
	for _, name := range []string{"GenerateSyntaxDenominator", "GenerateDenominatorAdmission",
		"GenerateDenominatorSelection", "GenerateDenominatorDigest", "GenerateDenominatorEvidence",
		"GenerateMigrationConformance"} {
		projection, found := ir.Graph.NodeByName(ir.Namespace, name)
		if !found || !ir.Graph.HasFact(semantic.FactKey{
			Subject: projection.ID, Predicate: semantic.Used, Object: output.ID,
		}) {
			return fmt.Errorf("registration append record is not an input to %s", name)
		}
	}
	return nil
}

func bindProfileHint(ir semantic.IR) error {
	input, inputFound := ir.Graph.NodeByName(ir.Namespace, "ProfileObservation")
	output, outputFound := ir.Graph.NodeByName(ir.Namespace, "ProfileHint")
	activity, activityFound := ir.Graph.NodeByName(ir.Namespace, "SuggestRegistrationProfile")
	if !inputFound || !outputFound || !activityFound || input.Kind != semantic.Entity ||
		output.Kind != semantic.Entity || activity.Kind != semantic.Activity ||
		input.ID.String() != "gooo://syntax-registration/profile-observation" ||
		output.ID.String() != "gooo://syntax-registration/profile-hint" ||
		activity.ID.String() != "gooo://syntax-registration/suggest-profile" || activity.ValueProgram == "" ||
		!ir.Graph.HasFact(semantic.FactKey{Subject: activity.ID, Predicate: semantic.Used, Object: input.ID}) ||
		!ir.Graph.HasFact(semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: activity.ID}) {
		return fmt.Errorf("registration profile-hint activity ABI mismatch")
	}
	return nil
}

func registrationRoles() ([10]string, [10]string) {
	return [10]string{"RegistrationCandidate", "SyntaxCorpus", "NativeRegistry",
			"SyntaxDenominator", "SyntaxConformance", "DenominatorAdmission",
			"DenominatorSelection", "DenominatorDigest", "DenominatorEvidence", "MigrationConformance"},
		[10]string{Operation, "GenerateSyntaxCorpus", "GenerateNativeRegistry",
			"GenerateSyntaxDenominator", "GenerateSyntaxConformance", "GenerateDenominatorAdmission",
			"GenerateDenominatorSelection", "GenerateDenominatorDigest",
			"GenerateDenominatorEvidence", "GenerateMigrationConformance"}
}
