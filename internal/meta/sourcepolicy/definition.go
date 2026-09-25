package sourcepolicy

type definition struct {
	family    Family
	limit     int
	relation  Relation
	blocking  bool
	role      IndicatorRole
	proof     ProofChoice
	operation Operation
	consumer  string
}
