package syntax

// PolicyDecl is a first-class meta-programming policy declaration. Its
// children are typed syntax nodes rather than an opaque activity value
// program, so later compiler stages can preserve the policy structure.
type PolicyDecl struct {
	Span        Span
	Name        string
	ID          string
	NameSpan    Span
	IDSpan      Span
	States      []StateDecl
	Transitions []TransitionDecl
	Cases       []CaseDecl
}

func (*PolicyDecl) declarationNode()   {}
func (d *PolicyDecl) SourceSpan() Span { return d.Span }

// StateDecl names a policy state or condition.
type StateDecl struct {
	Span     Span
	Name     string
	NameSpan Span
}

func (d StateDecl) SourceSpan() Span { return d.Span }

// TransitionDecl connects two declared policy states. Declaration order is
// semantically meaningful because the evaluator uses the first matching case.
type TransitionDecl struct {
	Span     Span
	From     string
	To       string
	FromSpan Span
	ToSpan   Span
}

func (d TransitionDecl) SourceSpan() Span { return d.Span }

// CaseDecl binds evidence to one resolution row.
type CaseDecl struct {
	Span       Span
	Name       string
	NameSpan   Span
	Evidence   []EvidenceDecl
	Resolution *ResolutionDecl
}

func (d CaseDecl) SourceSpan() Span { return d.Span }

// EvidenceDecl is a named, source-owned input to a policy case.
type EvidenceDecl struct {
	Span      Span
	Name      string
	Value     string
	NameSpan  Span
	ValueSpan Span
}

func (d EvidenceDecl) SourceSpan() Span { return d.Span }

// ResolutionDecl contains the complete decision coordinate. UNKNOWN rows
// must carry all six resolution fields; known rows leave those fields absent.
type ResolutionDecl struct {
	Span           Span
	Decision       string
	Stage          string
	Step           int
	Reason         string
	DecisionStage  string
	DecisionStep   int
	DecisionReason string
	UnknownClass   string
	NextOperation  string
	BlockedBy      []string
	Role           string
	MetaOperation  string
	ProofChoice    string
	Claim          string
}

func (d ResolutionDecl) SourceSpan() Span { return d.Span }

// Clone returns a detached policy syntax tree.
func (d PolicyDecl) Clone() *PolicyDecl {
	clone := d
	clone.States = append([]StateDecl(nil), d.States...)
	clone.Transitions = append([]TransitionDecl(nil), d.Transitions...)
	if d.Cases != nil {
		clone.Cases = make([]CaseDecl, len(d.Cases))
		for index, current := range d.Cases {
			clone.Cases[index] = current
			clone.Cases[index].Evidence = append([]EvidenceDecl(nil), current.Evidence...)
			if current.Resolution != nil {
				resolution := *current.Resolution
				resolution.BlockedBy = append([]string(nil), current.Resolution.BlockedBy...)
				clone.Cases[index].Resolution = &resolution
			}
		}
	}
	return &clone
}
