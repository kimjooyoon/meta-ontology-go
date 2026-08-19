package formatter

// DeclarationKind identifies the currently supported .gooo declaration forms.
type DeclarationKind string

const (
	EntityDeclaration   DeclarationKind = "entity"
	ActivityDeclaration DeclarationKind = "activity"
)

// Declaration is a syntax-neutral semantic declaration. Entity IDs are
// authoritative. Activity IDs are optional because the initial surface
// grammar derives them from namespace and display name.
type Declaration struct {
	Kind   DeclarationKind
	Name   string
	ID     string
	Inputs []string
	Output string
}

// Document is the minimal semantic view consumed by the formatter.
type Document struct {
	Package      string
	Namespace    string
	Declarations []Declaration
}

// Clone returns a detached document so adapters can safely reuse their AST.
func (d Document) Clone() Document {
	clone := Document{Package: d.Package, Namespace: d.Namespace}
	clone.Declarations = make([]Declaration, len(d.Declarations))
	for index, declaration := range d.Declarations {
		clone.Declarations[index] = declaration
		clone.Declarations[index].Inputs = append([]string(nil), declaration.Inputs...)
	}
	return clone
}
func (d Document) validate() Diagnostics {
	diagnostics := make(Diagnostics, 0)
	if !isIdentifier(d.Package) {
		diagnostics = appendInvalid(diagnostics, "package must be a non-empty identifier")
	}
	if !isIdentifier(d.Namespace) {
		diagnostics = appendInvalid(diagnostics, "namespace must be a non-empty identifier")
	}
	names := make(map[string]struct{}, len(d.Declarations))
	entityNames := make(map[string]struct{}, len(d.Declarations))
	entityIDs := make(map[string]struct{}, len(d.Declarations))
	for _, declaration := range d.Declarations {
		if _, exists := names[declaration.Name]; exists || declaration.Name == "" {
			diagnostics = appendInvalid(diagnostics, "declaration names must be unique and non-empty")
		}
		names[declaration.Name] = struct{}{}
		if declaration.Kind == EntityDeclaration {
			entityNames[declaration.Name] = struct{}{}
			diagnostics = validateEntity(diagnostics, declaration, entityIDs)
		}
	}
	activityIDs := make(map[string]struct{}, len(d.Declarations))
	for _, declaration := range d.Declarations {
		if declaration.Kind == ActivityDeclaration {
			diagnostics = validateActivity(diagnostics, declaration, d.Namespace, entityNames, entityIDs, activityIDs)
			continue
		}
		if declaration.Kind != EntityDeclaration {
			diagnostics = appendInvalid(diagnostics, "unsupported declaration kind "+string(declaration.Kind))
		}
	}
	return diagnostics
}
