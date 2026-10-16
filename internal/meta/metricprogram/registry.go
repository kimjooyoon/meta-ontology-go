package metricprogram

var canonicalOperations = sourceOperations()

func CanonicalOperations() []OperationSpec {
	return append([]OperationSpec(nil), canonicalOperations...)
}

func findOperation(id string) (OperationSpec, bool) {
	for _, operation := range canonicalOperations {
		if operation.ID == id {
			return operation, true
		}
	}
	return OperationSpec{}, false
}
