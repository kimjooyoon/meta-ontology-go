package workspaceexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

func isWorkspaceScalarIdentity(id string) bool {
	return id == string(semantic.BuiltinStringTypeID) || id == string(semantic.BuiltinBooleanTypeID) ||
		id == string(semantic.BuiltinIntegerTypeID)
}
