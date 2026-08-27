package syntax

// EntityDecl declares a named entity and its stable semantic identifier.
type EntityDecl struct {
	Span     Span
	Name     string
	ID       string
	NameSpan Span
	IDSpan   Span
	Fields   []FieldDecl
	// FieldsPresent distinguishes an explicit empty fields block from no block.
	FieldsPresent bool
}

func (*EntityDecl) declarationNode()   {}
func (d *EntityDecl) SourceSpan() Span { return d.Span }

// Clone returns a detached copy of the entity declaration, preserving field
// source order and preventing callers from sharing the Fields backing array.
func (d EntityDecl) Clone() EntityDecl {
	clone := d
	if d.Fields != nil {
		clone.Fields = make([]FieldDecl, len(d.Fields))
		for index, field := range d.Fields {
			clone.Fields[index] = field.Clone()
		}
	}
	return clone
}

// ActivityDecl declares an activity, its entity inputs, and its entity result.
type ActivityDecl struct {
	ValueProgram        string
	ValueProgramSpan    Span
	ValueProgramPresent bool
	Span                Span
	Name                string
	NameSpan            Span

	// Inputs and Output are the compact grammar-facing names. Parameters and
	// Result retain descriptive names for newer consumers.
	Inputs     []NameRef
	Output     string
	Parameters []NameRef
	Result     NameRef
}

func (*ActivityDecl) declarationNode()   {}
func (d *ActivityDecl) SourceSpan() Span { return d.Span }

// FreshnessDecl declares one formal evidence-freshness policy value.
type FreshnessDecl struct {
	Span   Span
	Kind   string
	Values []NameRef
}

func (d *FreshnessDecl) SourceSpan() Span { return d.Span }
