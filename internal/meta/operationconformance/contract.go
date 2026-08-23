package operationconformance

var splitGoV1Indicators = [...]IndicatorDefinition{
	{
		ID:             "filesystem.atomic-replacement/v1",
		Role:           RoleGuardrail,
		ProofRoute:     RouteCoherence,
		AuthorityURI:   "gooo://source-splitter/atomic-replacement",
		EvidenceSchema: "gooo/atomic-replacement-receipt/v1",
	},
	{
		ID:             "go.filename.build-semantics/v1",
		Role:           RoleDriver,
		ProofRoute:     RouteFoundation,
		AuthorityURI:   "https://pkg.go.dev/go/build#Context.MatchFile",
		EvidenceSchema: "gooo/go-build-file-set/v1",
	},
	{
		ID:             "go.header.preserved/v1",
		Role:           RoleGuardrail,
		ProofRoute:     RouteRegression,
		AuthorityURI:   "gooo://source-splitter/header-preservation",
		EvidenceSchema: "gooo/source-header-diff/v1",
	},
	{
		ID:             "go.import.identity/v1",
		Role:           RoleOutcome,
		ProofRoute:     RouteFoundation,
		AuthorityURI:   "https://go.dev/ref/spec#Import_declarations",
		EvidenceSchema: "gooo/go-import-binding/v1",
	},
	{
		ID:             "go.initialization.order/v1",
		Role:           RoleOutcome,
		ProofRoute:     RouteFoundation,
		AuthorityURI:   "https://go.dev/ref/spec#Package_initialization",
		EvidenceSchema: "gooo/go-initialization-graph/v1",
	},
	{
		ID:             "go.package.conformance/v1",
		Role:           RoleOutcome,
		ProofRoute:     RouteFoundation,
		AuthorityURI:   "https://go.dev/ref/spec#Package_clause",
		EvidenceSchema: "gooo/go-package-set/v1",
	},
}

func SplitGoV1Contract() Contract {
	indicators := append([]IndicatorDefinition(nil), splitGoV1Indicators[:]...)
	return Contract{
		Schema:        SplitGoV1Schema,
		ID:            SplitGoV1ContractID,
		Version:       1,
		Denominator:   SplitGoV1Denominator,
		UnknownPolicy: PreserveUnknownAndBlock,
		Indicators:    indicators,
	}
}

func SplitGoV1IndicatorIDs() []string {
	identifiers := make([]string, len(splitGoV1Indicators))
	for index, indicator := range splitGoV1Indicators {
		identifiers[index] = indicator.ID
	}
	return identifiers
}
