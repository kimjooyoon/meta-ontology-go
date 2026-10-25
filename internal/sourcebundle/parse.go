package sourcebundle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// Parsed contains a same-package syntax carrier, not a semantic admission.
// Every declaration keeps its physical member span. File.Span is deliberately
// unset: this carrier has no single physical source address.
type Parsed struct {
	File         *syntax.File
	SourceDigest string
	Members      []Member
}

func Parse(sources []Source) (Parsed, error) {
	return parseSources(sources, syntax.ParseFile)
}

// ParseWithEntityFieldsSupport uses one explicit profile across every member.
// Duplicate identities and references still require the consumer's lowering.
func ParseWithEntityFieldsSupport(sources []Source, support syntax.EntityFieldsSupport) (Parsed, error) {
	return parseSources(sources, func(name, content string) (*syntax.File, syntax.Diagnostics) {
		return syntax.ParseFileWithEntityFieldsSupport(name, content, support)
	})
}

func parseSources(sources []Source, parse func(string, string) (*syntax.File, syntax.Diagnostics)) (Parsed, error) {
	digest, err := Digest(sources)
	if err != nil {
		return Parsed{}, err
	}
	file, members, err := syntaxCarrier(sources, parse)
	if err != nil {
		return Parsed{}, err
	}
	return Parsed{File: file, SourceDigest: hex.EncodeToString(digest[:]), Members: members}, nil
}

func syntaxCarrier(sources []Source, parse func(string, string) (*syntax.File, syntax.Diagnostics)) (*syntax.File, []Member, error) {
	carrier := &syntax.File{}
	members := make([]Member, 0, len(sources))
	for i, source := range sources {
		file, diagnostics := parse(source.Filename, source.Content)
		if err := diagnostics.Error(); err != nil {
			return nil, nil, fmt.Errorf("source bundle member %q: %w", source.Filename, err)
		}
		if file == nil || file.Package == nil || file.Namespace == nil {
			return nil, nil, fmt.Errorf("source bundle member %q has invalid syntax or headers", source.Filename)
		}
		if len(file.Imports) != 0 {
			return nil, nil, fmt.Errorf("source bundle member %q needs package import resolution", source.Filename)
		}
		if i == 0 {
			carrier.Package, carrier.Namespace = file.Package, file.Namespace
		} else if file.Package.Name != carrier.Package.Name || file.Namespace.Name != carrier.Namespace.Name {
			return nil, nil, fmt.Errorf("source bundle member %q has different headers", source.Filename)
		}
		carrier.Decls = append(carrier.Decls, file.Decls...)
		carrier.Bindings = append(carrier.Bindings, file.Bindings...)
		sum := sha256.Sum256([]byte(source.Content))
		members = append(members, Member{source.Filename, hex.EncodeToString(sum[:]), len(source.Content)})
	}
	carrier.Declarations = carrier.Decls
	return carrier, members, nil
}
