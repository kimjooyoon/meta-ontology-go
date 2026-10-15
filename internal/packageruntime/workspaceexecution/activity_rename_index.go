package workspaceexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type workspaceRenameIndex struct {
	image      packageruntime.Image
	files      map[string][]*syntax.File
	sources    map[string]map[string]string
	calls      workspaceCalls
	identities map[string]string
}

func indexActivityRename(ctx context.Context, manifest packageruntime.Manifest) (workspaceRenameIndex, error) {
	index := workspaceRenameIndex{files: map[string][]*syntax.File{},
		sources: map[string]map[string]string{}, identities: map[string]string{}}
	if err := index.parseRenameSources(ctx, manifest); err != nil {
		return index, err
	}
	image, err := packageruntime.Build(manifest)
	if err != nil {
		return index, err
	}
	index.image = image
	index.calls = indexWorkspaceCalls(index.files, map[string]map[string]string{})
	for key, declaration := range index.calls {
		identity, err := workspaceActivityIdentity(declaration.namespace, declaration.ref.Activity, declaration.activity.ID)
		if err != nil {
			return index, fmt.Errorf("activity %s identity: %w", key, err)
		}
		index.identities[key] = identity
	}
	return index, nil
}

func (index workspaceRenameIndex) parseRenameSources(ctx context.Context, manifest packageruntime.Manifest) error {
	for _, pkg := range manifest.Packages {
		index.sources[pkg.Path] = map[string]string{}
		for _, source := range pkg.Sources {
			if err := ctx.Err(); err != nil {
				return err
			}
			file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(source.Filename, source.Content, syntax.EntityFieldsV4Support())
			if diagnostics.HasErrors() {
				return fmt.Errorf("rename source %q: %w", source.Filename, diagnostics.Error())
			}
			if file == nil {
				return fmt.Errorf("rename source %q is absent", source.Filename)
			}
			index.files[pkg.Path] = append(index.files[pkg.Path], file)
			index.sources[pkg.Path][source.Filename] = source.Content
		}
	}
	return nil
}

func (index workspaceRenameIndex) target(packagePath, id string) (string, workspaceCallDeclaration, error) {
	key := ""
	for candidate, identity := range index.identities {
		declaration := index.calls[candidate]
		if identity != id || declaration.ref.PackagePath != packagePath {
			continue
		}
		if key != "" {
			return "", workspaceCallDeclaration{}, fmt.Errorf("activity ID %q has multiple declarations in package %q", id, packagePath)
		}
		key = candidate
	}
	if key == "" {
		return "", workspaceCallDeclaration{}, fmt.Errorf("activity ID %q is absent from package %q", id, packagePath)
	}
	return key, index.calls[key], nil
}
