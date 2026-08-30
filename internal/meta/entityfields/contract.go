package entityfields

const (
	ReportSchema = "gooo/entity-fields-observation/v1"
	DecisionPass = "PASS"
	DecisionRefuted = "REFUTED"
	ResolutionExact = "EXACT"
	ResolutionLower = "LOWER_RESOLUTION"
)

type CellSpec struct {
	ID             string `json:"id"`
	Activity       string `json:"activity"`
	MetaOperation  string `json:"meta_operation"`
	ProofChoice    string `json:"proof_choice"`
	IndicatorClass string `json:"indicator_class"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	EvidenceKey    string `json:"evidence_key"`
}

var cellSpecs = []CellSpec{
	{"EF_SOURCE_PIN", "ParseEntityFields", "parse-entity-fields", "FOUNDATION", "DRIVER", "entityfieldsv1", "entityfields", "OBSERVATION_SOURCE_PIN"},
	{"EF_FORMAT_REPLAY", "FormatEntityFields", "format-entity-fields", "FOUNDATION", "OUTCOME", "entityfieldsv1", "entityfields", "SYNTAX_FORMAT_REPLAY"},
	{"EF_PROFILE_BINDING", "BindEntityFieldsProfile", "bind-entity-fields-profile", "FOUNDATION", "GUARDRAIL", "entityfieldsv1", "entityfields", "CONTRACT_PROFILE_PIN"},
	{"EF_SEMANTIC_LOWERING", "LowerEntityFields", "lower-entity-fields", "FOUNDATION", "DRIVER", "bidir", "entityfields", "SEMANTIC_IR_LOWERING"},
	{"EF_GLOBAL_IDS", "ValidateStableIDs", "validate-stable-ids", "COHERENCE", "GUARDRAIL", "bidir", "entityfields", "GLOBAL_ID_VALIDATION"},
	{"EF_DECLARATION_ORDER", "PreserveDeclarationOrder", "preserve-declaration-order", "COHERENCE", "OUTCOME", "entityfieldsv1", "entityfields", "DECLARATION_ORDER"},
	{"EF_BX_GET", "RoundTripGet", "roundtrip-get", "COHERENCE", "DRIVER", "bidir", "entityfields", "BX_GET_ROUNDTRIP"},
	{"EF_BX_PUT", "RoundTripPut", "roundtrip-put", "COHERENCE", "GUARDRAIL", "bidir", "entityfields", "BX_PUT_ROUNDTRIP"},
	{"EF_GO_PROJECTION", "GenerateGoProjection", "generate-go-projection", "REGRESSION", "OUTCOME", "generator", "entityfields", "GO_STRUCT_PROJECTION"},
	{"EF_SOURCE_MAP", "GenerateSourceMap", "generate-source-map", "REGRESSION", "DRIVER", "generator", "entityfields", "SOURCE_MAP_PROJECTION"},
	{"EF_LSP_NAVIGATION", "ResolveLSPNavigation", "resolve-lsp-navigation", "REGRESSION", "GUARDRAIL", "lsp", "entityfields", "LSP_NAVIGATION"},
	{"EF_RECEIPT", "PublishEntityFieldsReceipt", "publish-entity-fields-receipt", "REGRESSION", "OUTCOME", "entityfieldsv1", "entityfields", "ENTITY_FIELDS_RECEIPT"},
}

func CellSpecs() []CellSpec {
	return append([]CellSpec(nil), cellSpecs...)
}
