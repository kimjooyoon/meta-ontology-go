package sourcebundle

import (
	"crypto/sha256"
	"encoding/hex"
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
	digest, err := Digest(sources)
	if err != nil {
		return Bundle{}, err
	}
	carrier, members, err := syntaxCarrier(sources)
	if err != nil {
		return Bundle{}, err
	}
	document, err := bidir.DocumentFromSyntax(carrier)
	if err != nil {
		return Bundle{}, err
	}
	ir, err := bidir.LowerDocument(document)
	if err != nil {
		return Bundle{}, err
	}
	if err := ir.Validate(); err != nil {
		return Bundle{}, err
	}
	origins := make([]DeclarationOrigin, 0, len(document.Declarations))
	for _, declaration := range document.Declarations {
		node, ok := ir.Graph.NodeByName(ir.Namespace, declaration.Name)
		if !ok {
			return Bundle{}, fmt.Errorf("source bundle declaration %q was not lowered", declaration.Name)
		}
		origins = append(origins, DeclarationOrigin{ID: node.ID.String(), Span: declaration.Span})
	}
	return Bundle{Package: document.Package, Namespace: document.Namespace,
		SourceDigest: hex.EncodeToString(digest[:]), Declarations: carrier.Decls, IR: ir,
		Provenance: Provenance{Members: members, Declarations: origins}}, nil
}

func syntaxCarrier(sources []Source) (*syntax.File, []Member, error) {
	carrier := &syntax.File{}
	members := make([]Member, 0, len(sources))
	for i, source := range sources {
		file, diagnostics := syntax.ParseFile(source.Filename, source.Content)
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
