package syntax

import "github.com/kimjooyoon/meta-ontology-go/internal/inputbinding"

func (p *Parser) parseActivityInput() (NameRef, NameRef) {
	first := p.expectIdentifier("activity input", DiagExpectedIdentifier)
	if p.at(TokenIdentifier) {
		return p.parseActivityTypeReference("activity input type", DiagExpectedIdentifier), first
	}
	return p.finishActivityTypeReference(first, "activity input type", DiagExpectedIdentifier), NameRef{}
}

func (p *Parser) finishActivityInputNames(activity *ActivityDecl) {
	anyName := false
	for _, name := range activity.InputNames {
		anyName = anyName || name.Name != ""
	}
	if !anyName {
		activity.InputNames = nil
	}
	if _, err := ActivityInputNames(activity); err != nil {
		p.error(DiagExpectedIdentifier, activity.Span, err.Error())
	}
}

// ActivityInputNames returns the effective body bindings in signature order.
// Runtime ports continue to use the canonical positional names.
func ActivityInputNames(activity *ActivityDecl) ([]string, error) {
	var declared []string
	if len(activity.InputNames) > 0 {
		declared = make([]string, len(activity.InputNames))
		for i, name := range activity.InputNames {
			if name.Name != "" {
				if err := validateIdentifier(name.Name, "activity input name"); err != nil {
					return nil, err
				}
			}
			declared[i] = name.Name
		}
	}
	return inputbinding.Resolve(len(activity.Inputs), declared)
}

// ActivityDeclaredInputNames preserves omitted positions as empty strings.
// The parser validates this optional aligned list before it reaches consumers.
func ActivityDeclaredInputNames(activity *ActivityDecl) []string {
	if len(activity.InputNames) == 0 {
		return nil
	}
	names := make([]string, len(activity.InputNames))
	for i, name := range activity.InputNames {
		names[i] = name.Name
	}
	return names
}
