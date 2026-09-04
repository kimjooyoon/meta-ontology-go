package syntax

// Clone returns a detached copy of the parsed syntax tree. The declaration
// aliases retain their existing same-order relationship while every mutable
// slice and declaration node is copied.
func (f *File) Clone() *File {
	if f == nil {
		return nil
	}
	clone := *f
	if f.Package != nil {
		packageDecl := *f.Package
		clone.Package = &packageDecl
	}
	if f.Namespace != nil {
		namespaceDecl := *f.Namespace
		clone.Namespace = &namespaceDecl
	}
	if f.Decls == nil && f.Declarations == nil {
		return &clone
	}
	declarations := f.Decls
	if declarations == nil {
		declarations = f.Declarations
	}
	clonedDeclarations := make([]Declaration, len(declarations))
	for index, declaration := range declarations {
		clonedDeclarations[index] = cloneDeclaration(declaration)
	}
	clone.Decls = clonedDeclarations
	clone.Declarations = clonedDeclarations
	return &clone
}
func cloneDeclaration(declaration Declaration) Declaration {
	switch value := declaration.(type) {
	case *EntityDecl:
		if value == nil {
			return (*EntityDecl)(nil)
		}
		clone := value.Clone()
		return &clone
	case *ActivityDecl:
		if value == nil {
			return (*ActivityDecl)(nil)
		}
		clone := *value
		clone.Inputs = append([]NameRef(nil), value.Inputs...)
		clone.Parameters = append([]NameRef(nil), value.Parameters...)
		return &clone
	case *PolicyDecl:
		if value == nil {
			return (*PolicyDecl)(nil)
		}
		return value.Clone()
	default:
		return declaration
	}
}
