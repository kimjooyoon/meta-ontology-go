package symbolicinvocationusecase

import (
	"fmt"
	"slices"
)

func CanonicalContract() Contract {
	return Contract{
		Schema:                       "gooo/symbolic-invocation-usecase-contract/v1",
		Version:                      1,
		MetricID:                     "gooo.metric.user.symbolic-invocation-validation.v1",
		ExpectedGoVersion:            "1.27.0",
		ExpectedRegisteredEmitters:   3,
		ExpectedGoooFiles:            2,
		ExpectedGoFiles:              0,
		ExpectedGoooLines:            10,
		ExpectedFiles:                5,
		ExpectedDirectories:          0,
		ExpectedAcceptedInstances:    1,
		ExpectedRejectedInstances:    1,
		ExpectedGeneratedInstances:   1,
		ExpectedGeneratedGoldenMatches: 1,
		ExpectedDeterministicReplays: 1,
		ExpectedResourceSamples:      5,
		ExpectedRepositoryWrites:     0,
		ExpectedMutationAuthorities:  0,
		ExpectedNonClaims:            4,
		ExpectedValidator:            "github.com/santhosh-tekuri/jsonschema/cmd/jv@v0.7.0",
	}
}

func (contract Contract) Validate() error {
	if contract != CanonicalContract() {
		return fmt.Errorf("symbolic invocation use-case contract drifted")
	}
	return nil
}

func CanonicalNonClaims() []string {
	return []string{
		"value-level types",
		"domain correctness",
		"production readiness",
		"performance beyond this runner and fixed samples",
	}
}

func canonicalNonClaims(value []string) bool {
	return slices.Equal(value, CanonicalNonClaims())
}
