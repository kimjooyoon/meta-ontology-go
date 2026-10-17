package main

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"

func completeWorkspacePackage(pkg *packageruntime.PackageSpec, declared workspacePackage) error {
	return packageruntime.CompleteSourceMetadata(pkg, declared.Name, declared.Imports)
}
