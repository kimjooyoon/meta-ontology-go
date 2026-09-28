package domaincapability

type ScopeComparisonStatus string

const (
	ScopeComparable ScopeComparisonStatus = "COMPARABLE"
	ScopeChangedStatus ScopeComparisonStatus = "SCOPE_CHANGED"
)

type ScopeComparison struct {
	Status             ScopeComparisonStatus
	PreviousScopeDigest string
	CurrentScopeDigest  string
}

// CompareScopes makes scope drift explicit before any metric comparison.
func CompareScopes(previousExpected, currentExpected []string) ScopeComparison {
	previousDigest := ScopeDigest(previousExpected)
	currentDigest := ScopeDigest(currentExpected)
	status := ScopeComparable
	if ScopeChanged(previousDigest, currentDigest) {
		status = ScopeChangedStatus
	}
	return ScopeComparison{
		Status:              status,
		PreviousScopeDigest: previousDigest,
		CurrentScopeDigest:  currentDigest,
	}
}