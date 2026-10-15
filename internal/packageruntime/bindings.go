package packageruntime

import (
	"sort"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func localBindings(bindings []syntax.BindingDecl) []syntax.BindingDecl {
	result := make([]syntax.BindingDecl, 0, len(bindings))
	for _, binding := range bindings {
		if binding.Producer.PackageAlias == "" && binding.Consumer.PackageAlias == "" {
			result = append(result, binding)
		}
	}
	return result
}

func resolvePackageBindings(spec PackageSpec, file *syntax.File, local []Export,
	dependencies map[string][]Export, activitySources map[string]*syntax.File) ([]PackageBinding, error) {
	imports, err := sourceImportAliases(spec, file)
	if err != nil {
		return nil, err
	}
	var result []PackageBinding
	for _, binding := range file.Bindings {
		crossFile := binding.Producer.PackageAlias == "" && binding.Consumer.PackageAlias == "" &&
			(activitySources[binding.Producer.Activity.Name] != file || activitySources[binding.Consumer.Activity.Name] != file)
		if binding.Producer.PackageAlias == "" && binding.Consumer.PackageAlias == "" && !crossFile {
			continue
		}
		if !crossFile && (binding.Producer.PackageAlias == "" || binding.Consumer.PackageAlias != "") {
			return nil, reject("PACKAGE_BINDING_SCOPE_UNSUPPORTED", "%s supports imported activity outputs bound to local activity inputs", spec.Path)
		}
		producerPath, exists := imports[binding.Producer.PackageAlias]
		producerExports := dependencies[producerPath]
		if crossFile {
			producerPath, exists, producerExports = spec.Path, true, local
		}
		if !exists {
			return nil, reject("PACKAGE_BINDING_IMPORT_UNKNOWN", "%s binding uses undeclared import alias %q", spec.Path, binding.Producer.PackageAlias)
		}
		producer, ok := findActivity(producerExports, binding.Producer.Activity.Name)
		if !ok {
			return nil, reject("PACKAGE_BINDING_ACTIVITY_UNKNOWN", "%s has no exported activity %s.%s", spec.Path, producerPath, binding.Producer.Activity.Name)
		}
		if binding.Producer.Port.Name != "result" {
			return nil, reject("PACKAGE_BINDING_PORT_UNKNOWN", "%s.%s has no output port %q", producerPath, producer.Name, binding.Producer.Port.Name)
		}
		consumer, ok := findActivity(local, binding.Consumer.Activity.Name)
		if !ok {
			return nil, reject("PACKAGE_BINDING_ACTIVITY_UNKNOWN", "%s has no local activity %q", spec.Path, binding.Consumer.Activity.Name)
		}
		inputIndex, ok := activityInputIndex(consumer, binding.Consumer.Port.Name)
		if !ok {
			return nil, reject("PACKAGE_BINDING_PORT_UNKNOWN", "%s.%s has no input port %q", spec.Path, consumer.Name, binding.Consumer.Port.Name)
		}
		if producer.OutputType != consumer.InputTypes[inputIndex] {
			return nil, reject("PACKAGE_BINDING_TYPE_MISMATCH", "%s.%s.result has type %s; %s.%s has type %s", producerPath, producer.Name, producer.OutputType, spec.Path, consumer.Name, consumer.InputTypes[inputIndex])
		}
		resolved := PackageBinding{
			ProducerPackage: producerPath, ProducerActivity: producer.Name, ProducerPort: "result",
			ConsumerPackage: spec.Path, ConsumerActivity: consumer.Name, ConsumerPort: binding.Consumer.Port.Name,
			EntityID: producer.OutputType, Feedback: binding.Feedback,
		}
		if crossFile {
			resolved.Source = &PackageBindingSource{Filename: binding.Span.Filename,
				Start: binding.Span.Start.Offset, End: binding.Span.End.Offset,
				ProducerID: producer.ID, ConsumerID: consumer.ID}
		}
		result = append(result, resolved)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].ConsumerActivity != result[j].ConsumerActivity {
			return result[i].ConsumerActivity < result[j].ConsumerActivity
		}
		return result[i].ConsumerPort < result[j].ConsumerPort
	})
	return result, nil
}

func findActivity(exports []Export, name string) (Export, bool) {
	var found Export
	count := 0
	for _, export := range exports {
		if export.Kind == "activity" && export.Name == name {
			found = export
			count++
		}
	}
	return found, count == 1
}

func activityInputIndex(activity Export, port string) (int, bool) {
	if len(activity.InputTypes) == 1 && port == "input" {
		return 0, true
	}
	for index := range activity.InputTypes {
		if port == "input"+strconv.Itoa(index) {
			return index, true
		}
	}
	return 0, false
}
