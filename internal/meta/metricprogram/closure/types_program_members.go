package closure

type operationDocument struct {
	ID                  string `json:"id"`
	Activity            string `json:"activity"`
	RepositoryWrites    bool   `json:"repository_writes"`
	PromotionAuthorized bool   `json:"promotion_authorized"`
}

type bindingDocument struct {
	IndicatorID string `json:"indicator_id"`
	OperationID string `json:"operation_id"`
}

type stepDocument struct {
	Activity string `json:"activity"`
}

type selectionDocument struct {
	ProofChoice   string `json:"proof_choice"`
	Decision      string `json:"decision"`
	MetaOperation string `json:"meta_operation"`
}

type coverageDocument struct {
	BindingCount               int    `json:"binding_count"`
	ResolvedBindingCount       int    `json:"resolved_binding_count"`
	RegistryOperationCount     int    `json:"registry_operation_count"`
	ReferencedOperationCount   int    `json:"referenced_operation_count"`
	SelectionOperationResolved bool   `json:"selection_operation_resolved"`
	Status                     string `json:"status"`
}
