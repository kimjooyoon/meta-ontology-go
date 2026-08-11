package semantic

// PROVName is the compact vocabulary exposed to serializers and query tools.
const (
	PROVEntity            = "prov:Entity"
	PROVActivity          = "prov:Activity"
	PROVAgent             = "prov:Agent"
	PROVUsed              = "prov:used"
	PROVWasGeneratedBy    = "prov:wasGeneratedBy"
	PROVWasDerivedFrom    = "prov:wasDerivedFrom"
	PROVWasAssociatedWith = "prov:wasAssociatedWith"
)

// Inverse returns a query projection without mutating canonical facts.
func Inverse(relation RelationKind) RelationKind {
	switch relation {
	case Used:
		return "usedBy"
	case WasGeneratedBy:
		return "generated"
	case WasDerivedFrom:
		return "derived"
	default:
		return relation
	}
}
