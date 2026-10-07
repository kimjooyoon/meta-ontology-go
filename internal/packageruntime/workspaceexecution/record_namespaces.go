package workspaceexecution

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type EntityAlias struct {
	PackagePath string `json:"package_path"`
	Name        string `json:"name"`
	EntityID    string `json:"entity_id"`
	LoweredName string `json:"lowered_name"`
}

type workspaceRecordNames struct {
	canonical map[string]string
	known     map[string]bool
	visible   map[string]map[string]string
	aliases   []EntityAlias
}

func scopeWorkspaceRecords(ordered []packageruntime.PackageSpec, files map[string][]*syntax.File,
	image packageruntime.Image, active map[string]ActivityRef) (workspaceRecordNames, error) {
	n := workspaceRecordNames{canonical: map[string]string{}, known: map[string]bool{}, visible: map[string]map[string]string{}}
	byName := map[string]map[string]bool{}
	byID := map[string]*syntax.EntityDecl{}
	for _, pkg := range ordered {
		for _, file := range files[pkg.Path] {
			for _, declaration := range file.Declarations {
				entity, ok := declaration.(*syntax.EntityDecl)
				if !ok {
					continue
				}
				if byName[entity.Name] == nil {
					byName[entity.Name] = map[string]bool{}
				}
				byName[entity.Name][entity.ID] = true
				if !entity.FieldsPresent {
					continue
				}
				if entity.Name == "Integer" || entity.Name == "Boolean" || entity.Name == "Text" {
					return n, fmt.Errorf("scalar entity %q cannot also declare record fields", entity.Name)
				}
				if previous := byID[entity.ID]; previous != nil && !sameEntityShape(previous, entity) {
					return n, fmt.Errorf("stable record ID %q has conflicting field declarations", entity.ID)
				}
				byID[entity.ID], n.known[entity.Name] = entity, true
				if previous := n.canonical[entity.ID]; previous == "" || entity.Name < previous {
					n.canonical[entity.ID] = entity.Name
				}
			}
		}
	}
	for id, name := range n.canonical {
		if len(byName[name]) > 1 {
			lowered := "GoooEntity" + strings.TrimPrefix(sourceSHA256([]byte(id)), "sha256:")
			for byName[lowered] != nil {
				lowered += "X"
			}
			n.canonical[id] = lowered
		}
	}
	for _, pkg := range image.Packages {
		n.visible[pkg.Path] = visibleRecordNames(pkg, image, n.canonical)
	}
	if err := n.rewriteFiles(ordered, files, image, active); err != nil {
		return n, err
	}
	return n, nil
}

func visibleRecordNames(pkg packageruntime.PackageImage, image packageruntime.Image, canonical map[string]string) map[string]string {
	identities := map[string]string{}
	for _, imported := range pkg.Imports {
		for _, dependency := range image.Packages {
			if dependency.Path != imported {
				continue
			}
			for _, export := range dependency.Exports {
				if export.Kind != "entity" {
					continue
				}
				if previous, exists := identities[export.Name]; exists && previous != export.ID {
					identities[export.Name] = ""
				} else if !exists {
					identities[export.Name] = export.ID
				}
			}
		}
	}
	for _, export := range pkg.Exports {
		if export.Kind == "entity" {
			identities[export.Name] = export.ID
		}
	}
	result := map[string]string{}
	for name, id := range identities {
		if lowered, exists := canonical[id]; exists {
			result[name] = lowered
		} else if id == "" {
			result[name] = ""
		}
	}
	return result
}

func (n *workspaceRecordNames) rewriteFiles(ordered []packageruntime.PackageSpec, files map[string][]*syntax.File,
	image packageruntime.Image, active map[string]ActivityRef) error {
	for _, pkg := range ordered {
		for _, file := range files[pkg.Path] {
			for _, declaration := range file.Declarations {
				switch value := declaration.(type) {
				case *syntax.EntityDecl:
					if lowered := n.canonical[value.ID]; lowered != "" && value.Name != lowered {
						n.aliases = append(n.aliases, EntityAlias{pkg.Path, value.Name, value.ID, lowered})
						value.Name, value.NameSpan = lowered, syntax.Span{}
					}
				case *syntax.ActivityDecl:
					if ref, ok := active[value.Name]; ok {
						if err := n.rewriteActivity(ref, value, image); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	sort.Slice(n.aliases, func(i, j int) bool {
		a, b := n.aliases[i], n.aliases[j]
		if a.PackagePath != b.PackagePath {
			return a.PackagePath < b.PackagePath
		}
		return a.Name < b.Name
	})
	return nil
}

func (n workspaceRecordNames) rewriteActivity(ref ActivityRef, activity *syntax.ActivityDecl, image packageruntime.Image) error {
	for _, pkg := range image.Packages {
		if pkg.Path != ref.PackagePath {
			continue
		}
		for _, export := range pkg.Exports {
			if export.Kind != "activity" || export.Name != ref.Activity {
				continue
			}
			for i, id := range export.InputTypes {
				if lowered := n.canonical[id]; lowered != "" {
					activity.Inputs[i].Name = lowered
					if i < len(activity.Parameters) {
						activity.Parameters[i].Name = lowered
					}
				}
			}
			if lowered := n.canonical[export.OutputType]; lowered != "" {
				activity.Output, activity.Result.Name = lowered, lowered
			}
		}
	}
	body, err := n.rewriteConstructors(ref.PackagePath, activity.ValueProgram, true)
	if err != nil {
		return fmt.Errorf("package %s activity %s: %w", ref.PackagePath, ref.Activity, err)
	}
	activity.ValueProgram = body
	return n.rewriteAssembly(ref.PackagePath, activity.Assembly)
}

func collectWorkspaceEntities(ordered []packageruntime.PackageSpec, files map[string][]*syntax.File) ([]syntax.Declaration, error) {
	seen := map[string]*syntax.EntityDecl{}
	var declarations []syntax.Declaration
	for _, pkg := range ordered {
		for _, file := range files[pkg.Path] {
			for _, declaration := range file.Declarations {
				entity, ok := declaration.(*syntax.EntityDecl)
				if !ok {
					continue
				}
				if previous := seen[entity.Name]; previous != nil {
					if previous.ID != entity.ID {
						return nil, fmt.Errorf("entity name %q has different stable IDs across workspace packages", entity.Name)
					}
					if !sameEntityShape(previous, entity) {
						return nil, fmt.Errorf("entity name %q has conflicting declarations across workspace packages", entity.Name)
					}
					continue
				}
				seen[entity.Name] = entity
				declarations = append(declarations, entity)
			}
		}
	}
	return declarations, nil
}
