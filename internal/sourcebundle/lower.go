package sourcebundle

import (
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type Member struct {
	Filename   string
	SHA256     string
	ByteLength int
}

type DeclarationOrigin struct {
	ID   string
	Span bidir.SourceSpan
}

type Provenance struct {
	Members      []Member
	Declarations []DeclarationOrigin
}

func (p Provenance) Clone() Provenance {
	return Provenance{Members: slices.Clone(p.Members), Declarations: slices.Clone(p.Declarations)}
}

type Bundle struct {
	Package, Namespace string
	SourceDigest       string
	Declarations       []syntax.Declaration
	IR                 semantic.IR
	Provenance         Provenance
}

// Lower preserves each parser's physical spans. No concatenated virtual source
// is attributed to a real file. Duplicate declarations and unknown references
// still pass through the existing semantic validator.
func Lower(sources []Source) (Bundle, error) {
	return LowerWithEntityFieldsSupport(sources, syntax.CurrentEntityFieldsSupport())
}

// LowerWithEntityFieldsSupport preserves one caller-selected profile through
// parsing and lowering. Source contents never widen that contract.
func LowerWithEntityFieldsSupport(sources []Source, support syntax.EntityFieldsSupport) (Bundle, error) {
	if err := support.Validate(); err != nil {
		return Bundle{}, err
	}
	parsed, err := ParseWithEntityFieldsSupport(sources, support)
	if err != nil {
		return Bundle{}, err
	}
	carrier := parsed.File
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(carrier, support)
	if err != nil {
		return Bundle{}, err
	}
	ir, err := bidir.LowerDocumentWithEntityFieldsSupport(document, support)
	if err != nil {
		return Bundle{}, err
	}
	if err := ir.Validate(); err != nil {
		return Bundle{}, err
	}
	origins := make([]DeclarationOrigin, 0, len(carrier.Decls))
	for _, declaration := range document.Declarations {
		node, ok := ir.Graph.NodeByName(ir.Namespace, declaration.Name)
		if !ok {
			return Bundle{}, fmt.Errorf("source bundle declaration %q was not lowered", declaration.Name)
		}
		origins = append(origins, DeclarationOrigin{ID: node.ID.String(), Span: declaration.Span})
	}
	// Policies lower outside Document.Declarations and the entity/activity graph.
	// Keep the parser's physical address; semantic span normalization may trim it.
	for _, declaration := range carrier.Decls {
		policy, ok := declaration.(*syntax.PolicyDecl)
		if !ok {
			continue
		}
		id, err := semantic.ParseIdentity(policy.ID)
		if err != nil {
			return Bundle{}, err
		}
		span := policy.Span
		origins = append(origins, DeclarationOrigin{ID: id.String(), Span: bidir.SourceSpan{
			File: span.Filename, Start: span.Start.Offset, End: span.End.Offset,
			StartLine: span.Start.Line, StartColumn: span.Start.Column,
			EndLine: span.End.Line, EndColumn: span.End.Column}})
	}
	return Bundle{Package: document.Package, Namespace: document.Namespace,
		SourceDigest: parsed.SourceDigest, Declarations: carrier.Decls, IR: ir,
		Provenance: Provenance{Members: parsed.Members, Declarations: origins}}, nil
}
