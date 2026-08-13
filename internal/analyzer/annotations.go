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
	active    bool
	conflict  bool
}

func collectRegistrations(file parsedFile, fileSet *token.FileSet) ([]Registration, Diagnostics) {
	var registrations []Registration
	var diagnostics Diagnostics
	for _, declaration := range file.file.Decls {
		switch current := declaration.(type) {
		case *ast.FuncDecl:
			parsed := parseAnnotations(current.Doc)
			registration, ok, found := registrationFor(file, fileSet, parsed, KindActivity, SymbolRef{
				PackagePath: file.packagePath,
				PackageName: file.packageName,
				Receiver:    receiverTypeName(current.Recv),
				Name:        current.Name.Name,
			})
			diagnostics = append(diagnostics, found...)
			if ok {
				registrations = append(registrations, registration)
			}
		case *ast.GenDecl:
			foundRegistrations, foundDiagnostics := collectGenRegistrations(file, fileSet, current)
			registrations = append(registrations, foundRegistrations...)
			diagnostics = append(diagnostics, foundDiagnostics...)
		}
	}
	return registrations, diagnostics
}

func collectGenRegistrations(file parsedFile, fileSet *token.FileSet, declaration *ast.GenDecl) ([]Registration, Diagnostics) {
	var registrations []Registration
	var diagnostics Diagnostics
	for _, specification := range declaration.Specs {
		switch current := specification.(type) {
		case *ast.TypeSpec:
			parsed := mergeAnnotations(parseAnnotations(declaration.Doc), parseAnnotations(current.Doc))
			registration, ok, found := registrationFor(file, fileSet, parsed, KindEntity, SymbolRef{
				PackagePath: file.packagePath,
				PackageName: file.packageName,
				Name:        current.Name.Name,
			})
			diagnostics = append(diagnostics, found...)
			if ok {
				if registration.Kind == "" {
					registration.Kind = KindEntity
				}
				registrations = append(registrations, registration)
			}
		case *ast.ValueSpec:
			parsed := mergeAnnotations(parseAnnotations(declaration.Doc), parseAnnotations(current.Doc))
			if !parsed.active {
				continue
			}
			for _, name := range current.Names {
				registration, ok, found := registrationFor(file, fileSet, parsed, "", SymbolRef{
					PackagePath: file.packagePath,
					PackageName: file.packageName,
					Name:        name.Name,
				})
				diagnostics = append(diagnostics, found...)
				if ok {
					registrations = append(registrations, registration)
				}
			}
		}
	}
	return registrations, diagnostics
}

func registrationFor(file parsedFile, fileSet *token.FileSet, parsed annotation, defaultKind SymbolKind, ref SymbolRef) (Registration, bool, Diagnostics) {
	if !parsed.active {
		return Registration{}, false, nil
	}
	span := declarationSpan(fileSet, ref, file.file)
	if parsed.conflict {
		return Registration{}, false, Diagnostics{{
			Code:    DiagConflictingAnnotation,
			Message: "semantic annotation contains conflicting values",
			Span:    span,
		}}
	}
	if parsed.id == "" {
		return Registration{}, false, Diagnostics{{
			Code:    DiagInvalidAnnotation,
			Message: "semantic annotation requires an id",
			Span:    span,
		}}
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
		registration.Span = span
	}
	if err := validateRegistration(registration); err != nil {
		return Registration{}, false, Diagnostics{{
			Code:    DiagInvalidAnnotation,
			Message: err.Error(),
			Span:    span,
		}}
	}
	return registration, true, nil
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
			result.active = true
			applyAnnotationParts(&result, parts[1:])
		case "activity", "entity":
			result.active = true
			setAnnotationKind(&result, SymbolKind(parts[0]))
			applyAnnotationParts(&result, parts[1:])
		case "generated:start", "generated:end", "slot:start", "slot:end":
			continue
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
			setAnnotationKind(result, SymbolKind(part))
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
		setAnnotationKind(result, SymbolKind(value))
	case "namespace":
		result.namespace = value
	case "id", "identity":
		if result.id != "" && result.id != value {
			result.conflict = true
		}
		result.id = value
	}
}

func setAnnotationKind(result *annotation, kind SymbolKind) {
	if result.kind != "" && result.kind != kind {
		result.conflict = true
	}
	result.kind = kind
}

func mergeAnnotations(left, right annotation) annotation {
	result := left
	result.active = left.active || right.active
	result.conflict = left.conflict || right.conflict
	if right.kind != "" {
		setAnnotationKind(&result, right.kind)
	}
	if right.namespace != "" {
		result.namespace = right.namespace
	}
	if right.id != "" {
		if result.id != "" && result.id != right.id {
			result.conflict = true
		}
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
