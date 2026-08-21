package coupling

type oracleValidation struct {
	decision Decision
	reason   Reason
}
type normalizedSemantic struct {
	digest string
	facts  []string
}
type registryView struct {
	bySurface map[string]CodeBinding
	bySymbol  map[string]CodeBinding
	digest    string
}
type receiptView struct {
	bySurface map[string]CouplingReceipt
	valid     []string
}
type pathView struct {
	digest   string
	counts   ObservationCounts
	decision Decision
	reason   Reason
}
