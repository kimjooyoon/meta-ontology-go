package workspaceexecution

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// ActivityRef preserves package identity when workspace activities are lowered
// into the single typed graph accepted by bodyexecution.
type ActivityRef struct {
	PackagePath string `json:"package_path"`
	Activity    string `json:"activity"`
	LoweredName string `json:"lowered_name"`
}

type Program struct {
	Schema     string               `json:"schema"`
	Workspace  packageruntime.Image `json:"workspace_image"`
	Entry      ActivityRef          `json:"entry"`
	Activities []ActivityRef        `json:"activities"`
	Source     string               `json:"lowered_gooo_source"`
	Scope      string               `json:"scope"`
}

// Prepare validates the package manifest and flattens its reachable package
// declarations into one Gooo typed graph. Imported binds become ordinary
// explicit Gooo binds; no activity body is evaluated here.
func Prepare(manifest packageruntime.Manifest) (Program, error) {
	image, err := packageruntime.Build(manifest)
	if err != nil {
		return Program{}, err
	}
	if len(image.Entry.PackagePath) == 0 {
		return Program{}, fmt.Errorf("workspace entry is unresolved")
	}
	packages := make(map[string]packageruntime.PackageSpec, len(manifest.Packages))
	for _, spec := range manifest.Packages {
		packages[spec.Path] = spec
	}
	ordered := make([]packageruntime.PackageSpec, 0, len(image.InitOrder))
	for _, packagePath := range image.InitOrder {
		ordered = append(ordered, packages[packagePath])
	}
	files := make(map[string][]*syntax.File, len(ordered))
	activityNames := map[string]map[string]string{}
	allActivityRefs := make([]ActivityRef, 0)
	activityOrdinal := 0
	for _, spec := range ordered {
		activityNames[spec.Path] = map[string]string{}
		sources := append([]packageruntime.Source(nil), spec.Sources...)
		sort.Slice(sources, func(i, j int) bool { return sources[i].Filename < sources[j].Filename })
		for _, source := range sources {
			file, diagnostics := syntax.ParseFile(source.Filename, source.Content)
			if diagnostics.HasErrors() || file == nil {
				return Program{}, fmt.Errorf("workspace source %q has syntax errors", source.Filename)
			}
			files[spec.Path] = append(files[spec.Path], file)
			for _, declaration := range file.Declarations {
				activity, ok := declaration.(*syntax.ActivityDecl)
				if !ok {
					continue
				}
				lowered := numberedActivityName(activityOrdinal, activity.Name)
				activityOrdinal++
				activityNames[spec.Path][activity.Name] = lowered
				allActivityRefs = append(allActivityRefs, ActivityRef{PackagePath: spec.Path, Activity: activity.Name, LoweredName: lowered})
			}
		}
	}
	entities := map[string]string{}
	entityDeclarations := map[string]*syntax.EntityDecl{}
	entityDecls := map[string]bool{}
	activityDeclarations := map[string]*syntax.ActivityDecl{}
	activityOrder := make([]string, 0, len(allActivityRefs))
	type pendingBinding struct {
		decl        syntax.BindingDecl
		producerKey string
		consumerKey string
	}
	pendingBindings := make([]pendingBinding, 0)
	var declarations []syntax.Declaration
	var bindings []syntax.BindingDecl
	for _, spec := range ordered {
		for _, file := range files[spec.Path] {
			imports := importAliases(spec.Path, file.Imports)
			for _, declaration := range file.Declarations {
				switch value := declaration.(type) {
				case *syntax.EntityDecl:
					if previous, exists := entities[value.Name]; exists && previous != value.ID {
						return Program{}, fmt.Errorf("entity name %q has different stable IDs across workspace packages", value.Name)
					}
					entities[value.Name] = value.ID
					if previous, exists := entityDeclarations[value.Name]; exists && !sameEntityShape(previous, value) {
						return Program{}, fmt.Errorf("entity name %q has conflicting declarations across workspace packages", value.Name)
					}
					if !entityDecls[value.Name] {
						declarations = append(declarations, declaration)
						entityDecls[value.Name] = true
						entityDeclarations[value.Name] = value
					}
				case *syntax.ActivityDecl:
					key := packageActivityKey(spec.Path, value.Name)
					value.Name = activityNames[spec.Path][value.Name]
					value.NameSpan = syntax.Span{}
					activityDeclarations[key] = value
					activityOrder = append(activityOrder, key)
				default:
					// Package execution retains only the entity and activity declarations
					// needed by the entry's producer chain.
				}
			}
			for _, binding := range file.Bindings {
				producerPackage := spec.Path
				if binding.Producer.PackageAlias != "" {
					var ok bool
					producerPackage, ok = imports[binding.Producer.PackageAlias]
					if !ok {
						return Program{}, fmt.Errorf("binding in %q uses unknown import alias %q", spec.Path, binding.Producer.PackageAlias)
					}
				}
				consumerPackage := spec.Path
				if binding.Consumer.PackageAlias != "" {
					var ok bool
					consumerPackage, ok = imports[binding.Consumer.PackageAlias]
					if !ok {
						return Program{}, fmt.Errorf("binding in %q uses unknown consumer import alias %q", spec.Path, binding.Consumer.PackageAlias)
					}
				}
				producerName, producerOK := activityNames[producerPackage][binding.Producer.Activity.Name]
				consumerName, consumerOK := activityNames[consumerPackage][binding.Consumer.Activity.Name]
				if !producerOK || !consumerOK {
					return Program{}, fmt.Errorf("binding in %q refers to an activity outside the workspace", spec.Path)
				}
				producerKey := packageActivityKey(producerPackage, binding.Producer.Activity.Name)
				consumerKey := packageActivityKey(consumerPackage, binding.Consumer.Activity.Name)
				binding.Producer.PackageAlias = ""
				binding.Producer.Activity.Name = producerName
				binding.Consumer.PackageAlias = ""
				binding.Consumer.Activity.Name = consumerName
				pendingBindings = append(pendingBindings, pendingBinding{decl: binding, producerKey: producerKey, consumerKey: consumerKey})
			}
		}
	}
	entryKey := packageActivityKey(image.Entry.PackagePath, image.Entry.Activity)
	needed := map[string]bool{entryKey: true}
	for changed := true; changed; {
		changed = false
		for _, binding := range pendingBindings {
			if needed[binding.consumerKey] && !needed[binding.producerKey] {
				needed[binding.producerKey] = true
				changed = true
			}
		}
	}
	if len(needed) < 2 {
		return Program{}, fmt.Errorf("workspace entry requires at least one explicitly bound producer for body execution")
	}
	if len(needed) > 16 {
		return Program{}, fmt.Errorf("workspace entry execution path supports at most 16 activities; got %d", len(needed))
	}
	activeRefs := make([]ActivityRef, 0, len(needed))
	refsByName := make(map[string]ActivityRef, len(needed))
	for _, ref := range allActivityRefs {
		if needed[packageActivityKey(ref.PackagePath, ref.Activity)] {
			activeRefs = append(activeRefs, ref)
			refsByName[ref.LoweredName] = ref
		}
	}
	for _, key := range activityOrder {
		if needed[key] {
			declarations = append(declarations, activityDeclarations[key])
		}
	}
	for _, binding := range pendingBindings {
		if needed[binding.producerKey] && needed[binding.consumerKey] {
			bindings = append(bindings, binding.decl)
		}
	}
	if len(bindings) == 0 {
		return Program{}, fmt.Errorf("workspace body execution requires at least one explicit activity binding")
	}
	flattened := &syntax.File{
		Package:   &syntax.PackageDecl{Name: "gooo_workspace"},
		Namespace: &syntax.NamespaceDecl{Name: "gooo_workspace"},
		Decls:     declarations, Declarations: declarations, Bindings: bindings,
	}
	source, err := syntax.Format(flattened)
	if err != nil {
		return Program{}, fmt.Errorf("format lowered workspace graph: %w", err)
	}
	activities := append([]ActivityRef(nil), activeRefs...)
	sort.Slice(activities, func(i, j int) bool { return activities[i].LoweredName < activities[j].LoweredName })
	entryName := activityNames[image.Entry.PackagePath][image.Entry.Activity]
	entry, ok := refsByName[entryName]
	if !ok {
		return Program{}, fmt.Errorf("workspace entry activity disappeared during lowering")
	}
	return Program{Schema: "gooo/workspace-body-program/v1", Workspace: image, Entry: entry,
		Activities: activities, Source: source,
		Scope: "explicitly bound workspace activity bodies lowered to one typed Gooo graph; execution requires finite cases and native Go compilation"}, nil
}

func sameEntityShape(left, right *syntax.EntityDecl) bool {
	if left.FieldsPresent != right.FieldsPresent || len(left.Fields) != len(right.Fields) {
		return false
	}
	for index := range left.Fields {
		a, b := left.Fields[index], right.Fields[index]
		if a.ID != b.ID || a.Name != b.Name || a.TypeRef.Spelling != b.TypeRef.Spelling ||
			a.Presence != b.Presence || a.Cardinality != b.Cardinality {
			return false
		}
	}
	return true
}

func importAliases(packagePath string, imports []syntax.ImportDecl) map[string]string {
	result := make(map[string]string, len(imports))
	for _, declaration := range imports {
		alias := declaration.Alias
		if alias == "" {
			alias = packagePathBase(declaration.Path)
		}
		result[alias] = declaration.Path
	}
	return result
}

func packagePathBase(value string) string {
	if index := strings.LastIndexByte(value, '/'); index >= 0 {
		return value[index+1:]
	}
	return value
}

func packageActivityKey(packagePath, activity string) string { return packagePath + ":" + activity }

func numberedActivityName(index int, name string) string {
	return "GoooPackage" + strconv.Itoa(index) + "Activity" + name
}
