package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func entityFieldsProjectionMalformedCases() []struct {
	name string
	edit func(*semantic.IR, *bidir.Model, *syntax.EntityFieldsSupport)
	want string
} {
	return []struct {
		name string
		edit func(*semantic.IR, *bidir.Model, *syntax.EntityFieldsSupport)
		want string
	}{
		{name: "missing ID", edit: func(_ *semantic.IR, model *bidir.Model, _ *syntax.EntityFieldsSupport) {
			model.Nodes[0].Fields[0].ID = ""
		}, want: "GOOO-EF-V1"},
		{name: "duplicate ID", edit: func(_ *semantic.IR, model *bidir.Model, _ *syntax.EntityFieldsSupport) {
			model.Nodes[0].Fields[1].ID = model.Nodes[0].Fields[0].ID
		}, want: "duplicate"},
		{name: "wrong parent", edit: func(_ *semantic.IR, model *bidir.Model, _ *syntax.EntityFieldsSupport) {
			model.Nodes[0].Fields[0].Parent = "billing://entity/other"
		}, want: "parent"},
		{name: "wrong snapshot", edit: func(_ *semantic.IR, model *bidir.Model, _ *syntax.EntityFieldsSupport) {
			model.Nodes[0].Fields[0].IDSpan.File = "other.gooo"
		}, want: "source"},
		{name: "unsupported shape", edit: func(ir *semantic.IR, model *bidir.Model, _ *syntax.EntityFieldsSupport) {
			model.Nodes[0].Fields[0].Presence = bidir.FieldPresenceOptional
			node, _ := ir.Graph.Node("billing://entity/order")
			node.Fields[0].Presence = semantic.Optional
			ir.Graph = semantic.NewGraph()
			_ = ir.AddNode(node)
		}, want: "UNSUPPORTED-SHAPE"},
		{name: "unsupported type", edit: func(ir *semantic.IR, model *bidir.Model, _ *syntax.EntityFieldsSupport) {
			model.Nodes[0].Fields[0].TypeRef.ID = "urn:gooo:type:integer"
			model.Nodes[0].Fields[0].TypeRefUse = bidir.TypeRefUse{Form: bidir.TypeRefFormStableID, Spelling: "urn:gooo:type:integer", ResolvedID: "urn:gooo:type:integer", Span: model.Nodes[0].Fields[0].TypeRefSpan}
			node, _ := ir.Graph.Node("billing://entity/order")
			node.Fields[0].TypeRef.ID = "urn:gooo:type:integer"
			ir.Graph = semantic.NewGraph()
			_ = ir.AddNode(node)
		}, want: "UNKNOWN-TYPE"},
		{name: "unbound profile", edit: func(_ *semantic.IR, _ *bidir.Model, support *syntax.EntityFieldsSupport) {
			support.Profile = syntax.EntityFieldsProfile{}
		}, want: "UNBOUND-PROFILE"},
		{name: "profile mismatch", edit: func(_ *semantic.IR, _ *bidir.Model, support *syntax.EntityFieldsSupport) {
			support.Profile.ID = "other"
		}, want: "PROFILE-MISMATCH"},
		{name: "profile digest mismatch", edit: func(_ *semantic.IR, _ *bidir.Model, support *syntax.EntityFieldsSupport) {
			support.Profile.Digest = "tampered"
		}, want: "PROFILE-DIGEST-MISMATCH"},
		{name: "unknown state", edit: func(_ *semantic.IR, _ *bidir.Model, support *syntax.EntityFieldsSupport) {
			support.State = "UNKNOWN"
		}, want: "UNKNOWN-STATE"},
	}
}
