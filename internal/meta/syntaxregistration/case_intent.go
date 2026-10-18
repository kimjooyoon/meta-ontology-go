package syntaxregistration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

// CaseIntent is the small, explicit input to Gooo's case-composition activity.
// Source identity and runtime pinning are added in a separate inspection step.
type CaseIntent struct {
	ID                string `json:"id"`
	Path              string `json:"path"`
	ProofChoice       string `json:"proof_choice"`
	Profile           string `json:"profile"`
	BaseVersion       int    `json:"base_version"`
	PromoteMetaSource bool   `json:"promote_meta_source,omitempty"`
}

type caseValue struct {
	ID                    string `json:"case_id"`
	Path                  string `json:"path"`
	Kind                  string `json:"kind"`
	ExpectedDecision      string `json:"expected_decision"`
	ProofChoice           string `json:"proof_choice"`
	MetaOperation         string `json:"meta_operation"`
	Scope                 string `json:"scope"`
	OrderedCollections    bool   `json:"ordered_collections"`
	EntityFields          bool   `json:"entity_fields"`
	ImplicitActivityPorts bool   `json:"implicit_activity_ports"`
	EntityFieldsProfile   string `json:"entity_fields_profile"`
	Accepted              bool   `json:"accepted"`
}

func DecodeCaseIntent(raw []byte) (CaseIntent, error) {
	var intent CaseIntent
	if err := decodeStrict(raw, &intent); err != nil {
		return CaseIntent{}, err
	}
	return intent, nil
}

// ComposeDraft evaluates Gooo's fixed case activity. It does not inspect the
// repository, select a denominator version or grant registration authority.
func ComposeDraft(ctx context.Context, intent CaseIntent) (Request, bodycodegen.FixedValue, error) {
	return composeDraftSource(ctx, intent, contractSource)
}

func composeDraftSource(ctx context.Context, intent CaseIntent, source []byte) (Request, bodycodegen.FixedValue, error) {
	input, err := json.Marshal([]any{struct {
		ID          string `json:"case_id"`
		Path        string `json:"path"`
		ProofChoice string `json:"proof_choice"`
		Profile     string `json:"profile"`
	}{intent.ID, intent.Path, intent.ProofChoice, intent.Profile}})
	if err != nil {
		return Request{}, bodycodegen.FixedValue{}, err
	}
	value, err := bodycodegen.EvaluateFixedValue(ctx, "internal/meta/syntaxregistration/contract.gooo",
		source, "ComposeCase", input)
	if err != nil {
		return Request{}, value, fmt.Errorf("evaluate Gooo registration case: %w", err)
	}
	if value.ActivityID != "gooo://syntax-registration/compose-case" || value.SourceSHA256 != digest(source) ||
		value.ProjectionSHA256 == "" {
		return Request{}, value, failure("REFUTED", "compose-case", "REGISTRATION_CASE_ORIGIN_MISMATCH", "", "restore-case-source")
	}
	var result caseValue
	if err := decodeStrict(value.Value, &result); err != nil {
		return Request{}, value, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(value.Value, &fields); err != nil || len(fields) != 12 {
		return Request{}, value, failure("REFUTED", "compose-case", "REGISTRATION_CASE_VALUE_INCOMPLETE", "", "repair-case-activity")
	}
	for _, field := range fields {
		if bytes.Equal(bytes.TrimSpace(field), []byte("null")) {
			return Request{}, value, failure("REFUTED", "compose-case", "REGISTRATION_CASE_VALUE_INCOMPLETE", "", "repair-case-activity")
		}
	}
	if !result.Accepted || result.ID != intent.ID || result.Path != intent.Path || result.ProofChoice != intent.ProofChoice {
		return Request{}, value, failure("REFUTED", "compose-case", "REGISTRATION_CASE_INTENT_INVALID", "", "correct-explicit-intent")
	}
	definition := languagesyntax.CaseDefinition{ID: result.ID, Path: result.Path, Kind: result.Kind,
		ExpectedDecision: result.ExpectedDecision, ProofChoice: result.ProofChoice,
		MetaOperation: result.MetaOperation, Scope: result.Scope,
		OrderedCollections: result.OrderedCollections, EntityFields: result.EntityFields,
		ImplicitActivityPorts: result.ImplicitActivityPorts, EntityFieldsProfile: result.EntityFieldsProfile}
	if definition.Kind != languagesyntax.KindValid || definition.ExpectedDecision != languagesyntax.DecisionPass ||
		definition.MetaOperation != "replay-language-syntax" || definition.Scope != languagesyntax.ScopeLanguageCapability ||
		!languagesyntax.ValidProfileSelection(definition) {
		return Request{}, value, failure("REFUTED", "compose-case", "REGISTRATION_CASE_CONTRACT_MISMATCH", "", "repair-case-activity")
	}
	copy := intent
	return Request{Case: definition, BaseVersion: intent.BaseVersion, PromoteMetaSource: intent.PromoteMetaSource,
		CaseIntent: &copy, CaseSourceDigest: value.SourceSHA256}, value, nil
}

// VerifyCaseComposition prevents an inspected Gooo intent from silently
// becoming an unrelated case before candidate generation.
func VerifyCaseComposition(ctx context.Context, request Request) error {
	if request.CaseIntent == nil {
		if request.CaseSourceDigest != "" {
			return failure("REFUTED", "compose-case", "REGISTRATION_CASE_ORIGIN_MISMATCH", "", "restore-case-intent")
		}
		return nil
	}
	replayed, _, err := ComposeDraft(ctx, *request.CaseIntent)
	if err != nil {
		return err
	}
	if request.CaseSourceDigest != replayed.CaseSourceDigest {
		return failure("UNKNOWN", "compose-case", "REGISTRATION_CASE_SOURCE_STALE", "STALE", "recompose-case-intent")
	}
	if !reflect.DeepEqual(request.Case, replayed.Case) || request.BaseVersion != replayed.BaseVersion ||
		request.PromoteMetaSource != replayed.PromoteMetaSource {
		return failure("REFUTED", "compose-case", "REGISTRATION_CASE_INTENT_MISMATCH", "", "restore-composed-request")
	}
	return nil
}
