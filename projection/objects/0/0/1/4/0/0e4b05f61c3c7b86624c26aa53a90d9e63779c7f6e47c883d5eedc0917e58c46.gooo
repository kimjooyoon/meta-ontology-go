package bidir

// ReconciliationFixture supplies the source view and two Go-side updates for
// a measurable bidirectional contract experiment.
type ReconciliationFixture interface {
	Name() string
	Document() Document
	AcceptedDelta() FactDelta
	PartialDelta() FactDelta
}
