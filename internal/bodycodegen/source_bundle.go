package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// SourceBundleResult binds a deterministic pure body projection to the ordered
// source members and their physical declaration origins. SourceDigest uses the
// bundle digest, including names and exact bytes, rather than a virtual file.
type SourceBundleResult struct {
	Result
	Provenance sourcebundle.Provenance `json:"provenance"`
}

// GenerateSourceBundle compiles an explicitly selected same-package source set.
// Fixed pure calls may cross member boundaries; package imports need the package
// compiler. This entry point uses the existing deterministic lowering route.
func GenerateSourceBundle(sources []sourcebundle.Source, activity string) (SourceBundleResult, error) {
	parsed, err := parseBodySources(sources)
	if err != nil {
		return SourceBundleResult{}, err
	}
	p, err := prepareActivityBodyFile(parsed.File, activity)
	if err != nil {
		return SourceBundleResult{}, err
	}
	p.base, err = p.generate(preserveRoute)
	if err != nil {
		return SourceBundleResult{}, err
	}
	sha := "sha256:" + parsed.SourceDigest
	choice, err := chooseBodyRoute(context.Background(), p, sha, "", "", "")
	if err != nil {
		return SourceBundleResult{}, err
	}
	result, err := p.selectedResult(sha, choice)
	if err != nil {
		return SourceBundleResult{}, err
	}
	origins, err := p.declarationOrigins()
	if err != nil {
		return SourceBundleResult{}, err
	}
	return SourceBundleResult{Result: result, Provenance: sourcebundle.Provenance{
		Members: parsed.Members, Declarations: origins}}, nil
}

func parseBodySources(sources []sourcebundle.Source) (sourcebundle.Parsed, error) {
	parsed, err := sourcebundle.ParseWithEntityFieldsSupport(sources, syntax.EntityFieldsV3Support())
	if err == nil {
		return parsed, nil
	}
	// Match ParseBodyFile: required-field programs retain V3; an optional field
	// must be accepted by V4 for the whole set, not just by one isolated member.
	v4, v4Err := sourcebundle.ParseWithEntityFieldsSupport(sources, syntax.EntityFieldsV4Support())
	if v4Err == nil {
		return v4, nil
	}
	return sourcebundle.Parsed{}, fmt.Errorf("parse .gooo source bundle: %w", err)
}

func (p preparedBody) declarationOrigins() ([]sourcebundle.DeclarationOrigin, error) {
	origins := make([]sourcebundle.DeclarationOrigin, 0, len(p.file.Declarations))
	for _, declaration := range p.file.Declarations {
		var id string
		var span syntax.Span
		switch d := declaration.(type) {
		case *syntax.EntityDecl:
			id, span = p.entityIDs[d.Name], d.Span
		case *syntax.ActivityDecl:
			id, span = p.activityIDs[d.Name], d.Span
		case *syntax.PolicyDecl:
			identity, err := semantic.ParseIdentity(d.ID)
			if err != nil {
				return nil, err
			}
			id, span = identity.String(), d.Span
		default:
			return nil, fmt.Errorf("source bundle has an unsupported declaration %T", declaration)
		}
		if id == "" {
			return nil, fmt.Errorf("source bundle declaration has no resolved identity")
		}
		origins = append(origins, sourcebundle.DeclarationOrigin{ID: id, Span: bidir.SourceSpan{
			File: span.Filename, Start: span.Start.Offset, End: span.End.Offset,
			StartLine: span.Start.Line, StartColumn: span.Start.Column,
			EndLine: span.End.Line, EndColumn: span.End.Column}})
	}
	return origins, nil
}
