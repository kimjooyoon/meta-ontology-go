package semanticbinding

import (
	"sort"
)

func validateDirective(current directive) (directive, error) {
	names := make([]string, 0, len(current.fields))
	for name := range current.fields {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		value := current.fields[name]
		switch name {
		case "id", "subject", "pressure":
			canonical, err := normalizeIdentity(value)
			if err != nil {
				return directive{}, withErrorSpan(err, current.span)
			}
			current.fields[name] = canonical
		case "role":
			if !Role(value).valid() {
				return directive{}, bindingError(CodeInvalidRole, current.span, "unknown binding role")
			}
		default:
			return directive{}, bindingError(CodeUnknownField, current.span, "unknown directive field")
		}
	}
	return current, nil
}
func ensureRegistered(current directive, registered map[string]struct{}) error {
	if registered == nil {
		return nil
	}
	for _, name := range []string{"id", "subject", "pressure"} {
		value, present := current.fields[name]
		if !present {
			continue
		}
		if _, ok := registered[value]; !ok {
			return bindingError(CodeUnregisteredID, current.span, "identity is not present in the explicit registry")
		}
	}
	return nil
}
func bindingError(code Code, span Span, message string) *Error {
	return &Error{Code: code, Span: span, Message: message, FullSuiteFallback: true}
}
func withErrorSpan(err error, span Span) error {
	if typed, ok := err.(*Error); ok {
		copy := *typed
		copy.Span = span
		return &copy
	}
	return bindingError(CodeMalformedDirective, span, err.Error())
}
