package analyzer

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

type annotation struct {
	kind      SymbolKind
	namespace string
	id        string
}

func collectRegistrations(file parsedFile, fileSet *token.FileSet) []Registration {
	var registrations []Registration
	for _, declaration := range file.file.Decls {
		switch current := declaration.(type) {
		case *ast.FuncDecl:
			parsed := parseAnnotations(current.Doc)
			registration, ok := registrationFor(file, fileSet, parsed, KindActivity, SymbolRef{
				PackagePath: file.packagePath,
				PackageName: file.packageName,
				Receiver:    receiverTypeName(current.Recv),
				Name:        current.Name.Name,
			})
			if ok {
				registrations = append(registrations, registration)
			}
		case *ast.GenDecl:
			registrations = append(registrations, collectGenRegistrations(file, fileSet, current)...)
		}
	}
	return registrations
}

func collectGenRegistrations(file parsedFile, fileSet *token.FileSet, declaration *ast.GenDecl) []Registration {
	var registrations []Registration
	for _, specification := range declaration.Specs {
		switch current := specification.(type) {
		case *ast.TypeSpec:
			parsed := mergeAnnotations(parseAnnotations(declaration.Doc), parseAnnotations(current.Doc))
			registration, ok := registrationFor(file, fileSet, parsed, KindEntity, SymbolRef{
				PackagePath: file.packagePath,
				PackageName: file.packageName,
				Name:        current.Name.Name,
			})
			if ok {
				if registration.Kind == "" {
					registration.Kind = KindEntity
				}
				registrations = append(registrations, registration)
			}
		case *ast.ValueSpec:
			parsed := mergeAnnotations(parseAnnotations(declaration.Doc), parseAnnotations(current.Doc))
			if parsed.kind == "" {
				continue
			}
			for _, name := range current.Names {
				registration, ok := registrationFor(file, fileSet, parsed, "", SymbolRef{
					PackagePath: file.packagePath,
					PackageName: file.packageName,
					Name:        name.Name,
				})
				if ok {
					registrations = append(registrations, registration)
				}
			}
		}
	}
	return registrations
}

func registrationFor(file parsedFile, fileSet *token.FileSet, parsed annotation, defaultKind SymbolKind, ref SymbolRef) (Registration, bool) {
	if parsed.id == "" {
		return Registration{}, false
	}
	if parsed.kind == "" {
		parsed.kind = defaultKind
	}
	registration := Registration{
		Ref:      ref,
		Kind:     parsed.kind,
		Identity: Identity{Namespace: parsed.namespace, ID: parsed.id},
		Span:     spanFor(fileSet, file.file),
	}
	if ref.Receiver != "" || ref.Name != "" {
		registration.Span = declarationSpan(fileSet, ref, file.file)
	}
	if err := validateRegistration(registration); err != nil {
		return Registration{}, false
	}
	return registration, true
}

func declarationSpan(fileSet *token.FileSet, ref SymbolRef, file *ast.File) Span {
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == ref.Name && receiverTypeName(function.Recv) == ref.Receiver {
			return spanFor(fileSet, function)
		}
		if group, ok := declaration.(*ast.GenDecl); ok {
			for _, specification := range group.Specs {
				if namedSpec(specification, ref.Name) {
					return spanFor(fileSet, specification)
				}
			}
		}
	}
	return spanFor(fileSet, file)
}

func namedSpec(specification ast.Spec, name string) bool {
	switch current := specification.(type) {
	case *ast.TypeSpec:
		return current.Name.Name == name
	case *ast.ValueSpec:
		for _, candidate := range current.Names {
			if candidate.Name == name {
				return true
			}
		}
	}
	return false
}

func parseAnnotations(group *ast.CommentGroup) annotation {
	var result annotation
	if group == nil {
		return result
	}
	for _, comment := range group.List {
		text := strings.TrimSpace(comment.Text)
		text = strings.TrimSpace(strings.TrimSuffix(text, "*/"))
		text = strings.TrimSpace(strings.TrimPrefix(text, "//"))
		text = strings.TrimSpace(strings.TrimPrefix(text, "/*"))
		text = strings.TrimSpace(strings.TrimPrefix(text, "*"))
		if !strings.HasPrefix(text, "gooo:") {
			continue
		}
		parts := annotationFields(strings.TrimSpace(strings.TrimPrefix(text, "gooo:")))
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "semantic":
			applyAnnotationParts(&result, parts[1:])
		case "activity", "entity":
			result.kind = SymbolKind(parts[0])
			applyAnnotationParts(&result, parts[1:])
		default:
			applyAnnotationParts(&result, parts)
		}
	}
	return result
}

func applyAnnotationParts(result *annotation, parts []string) {
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if key, value, ok := strings.Cut(part, "="); ok {
			applyAnnotationKey(result, key, value)
			continue
		}
		switch part {
		case string(KindActivity), string(KindEntity):
			result.kind = SymbolKind(part)
		default:
			if result.id == "" {
				result.id = part
			}
		}
	}
}

func applyAnnotationKey(result *annotation, key, value string) {
	value = strings.TrimSpace(value)
	switch strings.TrimSpace(key) {
	case "kind":
		result.kind = SymbolKind(value)
	case "namespace":
		result.namespace = value
	case "id", "identity":
		result.id = value
	}
}

func mergeAnnotations(left, right annotation) annotation {
	result := left
	if right.kind != "" {
		result.kind = right.kind
	}
	if right.namespace != "" {
		result.namespace = right.namespace
	}
	if right.id != "" {
		result.id = right.id
	}
	return result
}

func annotationFields(value string) []string {
	var fields []string
	var current strings.Builder
	inQuote, escaped := false, false
	flush := func() {
		if current.Len() > 0 {
			fields = append(fields, current.String())
			current.Reset()
		}
	}
	for _, character := range value {
		switch {
		case escaped:
			current.WriteRune(character)
			escaped = false
		case character == '\\' && inQuote:
			current.WriteRune(character)
			escaped = true
		case character == '"':
			inQuote = !inQuote
		case (character == ' ' || character == '\t' || character == '\n') && !inQuote:
			flush()
		default:
			current.WriteRune(character)
		}
	}
	flush()
	for index, field := range fields {
		if unquoted, err := strconv.Unquote(field); err == nil {
			fields[index] = unquoted
		}
	}
	return fields
}
