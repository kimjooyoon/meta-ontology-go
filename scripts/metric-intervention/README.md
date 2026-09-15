# Metric intervention

This command binds the repository's real source metrics to a declarative,
sandbox-observed counterfactual. It compiles the mutation plan into a metric
delta, compares that prediction with the observed delta, and emits one
projection and indicator per registered dimension.

The projected values are an `ALGEBRAIC_ROOT_SCENARIO`, not an applied source
change. The repository workspace remains read-only and the output never
authorizes promotion.

The project root is deliberately exceptional: counts are `OBSERVED`, while
topology conformance and a root `README.md` requirement are
`NOT_APPLICABLE`.
