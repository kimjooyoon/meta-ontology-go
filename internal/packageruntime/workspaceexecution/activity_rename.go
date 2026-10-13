package workspaceexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"go/token"
	"go/types"
	"reflect"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// RenameWorkspaceActivity returns a source proposal for every supplied package.
// It includes consumers outside the entry closure and never writes or executes.
func RenameWorkspaceActivity(ctx context.Context, manifest packageruntime.Manifest,
	packagePath, activityID, to string) (WorkspaceActivityRename, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return WorkspaceActivityRename{}, err
	}
	before, err := indexActivityRename(ctx, manifest)
	if err != nil {
		return WorkspaceActivityRename{}, err
	}
	key, target, err := before.target(packagePath, activityID)
	if err != nil {
		return WorkspaceActivityRename{}, err
	}
	if err := validateWorkspaceRenameName(before, target, to); err != nil {
		return WorkspaceActivityRename{}, err
	}
	refs, err := before.references(ctx)
	if err != nil {
		return WorkspaceActivityRename{}, err
	}
	result, err := proposeWorkspaceRename(before, manifest, key, target, to, refs)
	if err != nil {
		return WorkspaceActivityRename{}, err
	}
	after, err := indexActivityRename(ctx, result.Manifest)
	if err != nil {
		return WorkspaceActivityRename{}, fmt.Errorf("renamed workspace: %w", err)
	}
	if err := verifyWorkspaceRename(ctx, before, after, refs, &result); err != nil {
		return WorkspaceActivityRename{}, err
	}
	result.OutputImageDigest = after.image.Digest
	return result, ctx.Err()
}

func validateWorkspaceRenameName(index workspaceRenameIndex, target workspaceCallDeclaration, to string) error {
	tokens, diagnostics := syntax.Lex(to)
	if target.ref.Activity == to || to == "_" || to == "let" || to == "init" || types.Universe.Lookup(to) != nil ||
		!token.IsIdentifier(to) || diagnostics.HasErrors() || len(tokens) != 2 ||
		tokens[0].Kind != syntax.TokenIdentifier || tokens[0].Text != to {
		return fmt.Errorf("choose a different Gooo activity identifier")
	}
	for _, pkg := range index.image.Packages {
		if pkg.Path != target.ref.PackagePath {
			continue
		}
		for _, export := range pkg.Exports {
			if export.Name == to {
				return fmt.Errorf("name %q already belongs to a declaration in package %q", to, pkg.Path)
			}
		}
	}
	return nil
}

func proposeWorkspaceRename(index workspaceRenameIndex, manifest packageruntime.Manifest, key string,
	target workspaceCallDeclaration, to string, refs []renameReference) (WorkspaceActivityRename, error) {
	r := WorkspaceActivityRename{Schema: "gooo/workspace-activity-rename/v1", ActivityID: index.identities[key],
		PackagePath: target.ref.PackagePath, From: target.ref.Activity, To: to, InputImageDigest: index.image.Digest,
		Scope: "all supplied source files, including consumers outside the entry; export identities/type roles and classified references retained; external cases/API consumers and native behavior require separate observations"}
	var err error
	if r.Manifest, err = detachedRenameManifest(manifest); err != nil {
		return r, err
	}
	if r.Manifest.Entry.PackagePath == r.PackagePath && r.Manifest.Entry.Activity == r.From {
		r.Manifest.Entry.Activity, r.EntryUpdated = to, true
	}
	changes := map[string]map[string][]WorkspaceRenameEdit{}
	add := func(pkg, filename string, start, end int, kind, replacement string) {
		if changes[pkg] == nil {
			changes[pkg] = map[string][]WorkspaceRenameEdit{}
		}
		source := index.sources[pkg][filename]
		changes[pkg][filename] = append(changes[pkg][filename], WorkspaceRenameEdit{start, end, kind, source[start:end], replacement})
	}
	span := target.activity.NameSpan
	add(r.PackagePath, target.source, span.Start.Offset, span.End.Offset, "declaration", to)
	if !target.activity.IDPresent {
		position := target.activity.Result.Span.End.Offset
		identity, _ := json.Marshal(r.ActivityID)
		add(r.PackagePath, target.source, position, position, "pin_identity", " id "+string(identity))
		r.IdentityPinned = true
	}
	addWorkspaceReferenceEdits(index, refs, r.PackagePath, r.ActivityID, to, &r.ResolvedReferences, add)
	r.Files, err = applyWorkspaceRenameFiles(&r.Manifest, changes)
	r.UpdatedReferences = r.ResolvedReferences
	return r, err
}

func detachedRenameManifest(manifest packageruntime.Manifest) (packageruntime.Manifest, error) {
	raw, err := json.Marshal(manifest)
	if err != nil {
		return packageruntime.Manifest{}, err
	}
	var result packageruntime.Manifest
	err = json.Unmarshal(raw, &result)
	return result, err
}

func verifyWorkspaceRename(ctx context.Context, before, after workspaceRenameIndex,
	refs []renameReference, result *WorkspaceActivityRename) error {
	afterRefs, err := after.references(ctx)
	if err != nil {
		return fmt.Errorf("renamed references: %w", err)
	}
	if !reflect.DeepEqual(renameReferenceIdentities(refs), renameReferenceIdentities(afterRefs)) {
		return fmt.Errorf("rename changed a resolved source reference identity")
	}
	if !reflect.DeepEqual(renameExportIdentities(before), renameExportIdentities(after)) {
		return fmt.Errorf("rename changed an exported identity or type role")
	}
	_, target, err := after.target(result.PackagePath, result.ActivityID)
	if err != nil || target.ref.Activity != result.To {
		return fmt.Errorf("renamed activity does not retain its original package and ID")
	}
	return ctx.Err()
}

func renameReferenceIdentities(refs []renameReference) []string {
	keys := make([]string, 0, len(refs))
	for _, ref := range refs {
		raw, _ := json.Marshal([]string{ref.packagePath, ref.filename, ref.callerID, ref.calleePackage, ref.calleeID, ref.surface})
		keys = append(keys, string(raw))
	}
	sort.Strings(keys)
	return keys
}

func renameExportIdentities(index workspaceRenameIndex) []string {
	var keys []string
	for _, pkg := range index.image.Packages {
		for _, export := range pkg.Exports {
			id := export.ID
			if export.Kind == "activity" {
				id = index.identities[packageActivityKey(pkg.Path, export.Name)]
			}
			raw, _ := json.Marshal([]any{pkg.Path, export.Kind, id, export.InputTypes, export.OutputType})
			keys = append(keys, string(raw))
		}
	}
	sort.Strings(keys)
	return keys
}

func addWorkspaceReferenceEdits(index workspaceRenameIndex, refs []renameReference, packagePath, id, to string,
	count *int, add func(string, string, int, int, string, string)) {
	for _, ref := range refs {
		if ref.calleeID != id || ref.calleePackage != packagePath {
			continue
		}
		replacement := to
		if ref.literal {
			source := index.sources[ref.packagePath][ref.filename]
			if strings.Contains(source[ref.start:ref.end], "\\") {
				raw, _ := json.Marshal(to)
				replacement = string(raw[1 : len(raw)-1])
			}
		}
		add(ref.packagePath, ref.filename, ref.start, ref.end, ref.surface, replacement)
		*count++
	}
}
