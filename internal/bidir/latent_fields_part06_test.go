package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func latentSyntaxFile() *syntax.File {
	return &syntax.File{
		Package:   &syntax.PackageDecl{Name: "billing"},
		Namespace: &syntax.NamespaceDecl{Name: "billing"},
		Declarations: []syntax.Declaration{
			&syntax.EntityDecl{
				Name: "Order", ID: "billing://entity/order",
				Fields: []syntax.FieldDecl{
					{Span: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 10}, End: syntax.Position{Offset: 50}}, ID: "billing://field/order-number", Name: "Order Number", TypeRef: syntax.TypeRefDecl{Span: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 31}, End: syntax.Position{Offset: 37}}, Spelling: "string"}, Presence: syntax.FieldPresenceRequired, Cardinality: syntax.FieldCardinalityOne,
						IDSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 10}, End: syntax.Position{Offset: 18}}, NameSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 19}, End: syntax.Position{Offset: 30}}, PresenceSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 38}, End: syntax.Position{Offset: 46}}, CardinalitySpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 47}, End: syntax.Position{Offset: 50}}},
					{Span: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 60}, End: syntax.Position{Offset: 102}}, ID: "billing://field/amount", Name: "Amount", TypeRef: syntax.TypeRefDecl{Span: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 81}, End: syntax.Position{Offset: 92}}, Spelling: "gooo:string"}, Presence: syntax.FieldPresenceRequired, Cardinality: syntax.FieldCardinalityOne,
						IDSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 60}, End: syntax.Position{Offset: 68}}, NameSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 69}, End: syntax.Position{Offset: 80}}, PresenceSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 93}, End: syntax.Position{Offset: 97}}, CardinalitySpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 98}, End: syntax.Position{Offset: 102}}},
				},
			},
			&syntax.EntityDecl{
				Name: "Payment", ID: "billing://entity/payment",
				Fields: []syntax.FieldDecl{{Span: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 110}, End: syntax.Position{Offset: 150}}, ID: "billing://field/receipt", Name: "Amount", TypeRef: syntax.TypeRefDecl{Span: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 131}, End: syntax.Position{Offset: 137}}, Spelling: "string"}, Presence: syntax.FieldPresenceRequired, Cardinality: syntax.FieldCardinalityOne,
					IDSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 110}, End: syntax.Position{Offset: 118}}, NameSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 119}, End: syntax.Position{Offset: 130}}, PresenceSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 138}, End: syntax.Position{Offset: 146}}, CardinalitySpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 147}, End: syntax.Position{Offset: 150}}}},
			},
		},
	}
}
