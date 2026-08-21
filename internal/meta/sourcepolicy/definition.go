package sourcepolicy

type definition struct {
	family    Family
	limit     int
	relation  Relation
	blocking  bool
	proof     ProofChoice
	operation Operation
	consumer  string
}
