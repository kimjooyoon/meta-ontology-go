package generator

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	runtimeBindingTypeName    = "GoooRuntimeBinding"
	runtimeExecutionTypeName  = "GoooRuntimeExecution"
	runtimeBindingErrorName   = "GoooRuntimeBindingError"
	runtimeBindingFunctionName = "ExecuteGoooRuntimeBindings"
)

func validateRuntimeBindingSupport(ir SemanticIR) error {
	reserved := map[string]struct{}{
		runtimeBindingTypeName: {}, runtimeExecutionTypeName: {}, runtimeBindingErrorName: {},
		runtimeBindingFunctionName: {}, runtimeBindingTypeName + "s": {}, "goooRuntimeActivities": {},
		"goooRuntimeActivityOrder": {}, "goooRuntimeInvoke": {}, "goooRuntimeReflect": {},
	}
	for _, entity := range ir.Entities {
		if _, exists := reserved[entity.GoName]; exists {
			return fmt.Errorf("generator: runtime binding helper name conflicts with entity %q", entity.GoName)
		}
	}
	for _, activity := range ir.Activities {
		if _, exists := reserved[activity.GoName]; exists {
			return fmt.Errorf("generator: runtime binding helper name conflicts with activity %q", activity.GoName)
		}
	}
	for _, item := range ir.Imports {
		if item.Path == "reflect" && (item.Name == "." || item.Name == "_") {
			return fmt.Errorf("generator: runtime binding support cannot use reflect import name %q", item.Name)
		}
	}
	return nil
}

func ensureRuntimeReflectionImport(imports []Import) []Import {
	for _, item := range imports {
		if item.Path == "reflect" {
			return imports
		}
	}
	return append(imports, Import{Path: "reflect", Name: "goooRuntimeReflect"})
}

func runtimeReflectionIdentifier(ir SemanticIR) string {
	for _, item := range ir.Imports {
		if item.Path == "reflect" {
			if item.Name == "" {
				return "reflect"
			}
			return item.Name
		}
	}
	return "goooRuntimeReflect"
}

func renderRuntimeBindingSupport(ir SemanticIR) string {
	reflection := runtimeReflectionIdentifier(ir)
	bindings := append([]RuntimeBinding(nil), ir.RuntimeBindings...)
	sort.SliceStable(bindings, func(i, j int) bool {
		left, right := bindings[i], bindings[j]
		if left.ProducerActivity != right.ProducerActivity {
			return left.ProducerActivity < right.ProducerActivity
		}
		if left.ConsumerActivity != right.ConsumerActivity {
			return left.ConsumerActivity < right.ConsumerActivity
		}
		if left.ProducerPort != right.ProducerPort {
			return left.ProducerPort < right.ProducerPort
		}
		return left.ConsumerPort < right.ConsumerPort
	})
	order := runtimeActivityOrder(ir)

	var output strings.Builder
	fmt.Fprintf(&output, "type %s struct {\n", runtimeBindingTypeName)
	output.WriteString("\tSchema string\n\tProducerActivity string\n\tProducerPort string\n\tConsumerActivity string\n\tConsumerPort string\n\tEntity string\n}\n\n")
	fmt.Fprintf(&output, "var %s = []%s{\n", runtimeBindingTypeName+"s", runtimeBindingTypeName)
	for _, binding := range bindings {
		fmt.Fprintf(&output, "\t{Schema: %s, ProducerActivity: %s, ProducerPort: %s, ConsumerActivity: %s, ConsumerPort: %s, Entity: %s},\n",
			goString(binding.Schema), goString(binding.ProducerActivity), goString(binding.ProducerPort),
			goString(binding.ConsumerActivity), goString(binding.ConsumerPort), goString(binding.Entity))
	}
	output.WriteString("}\n\n")

	fmt.Fprintf(&output, "type %s struct {\n", runtimeExecutionTypeName)
	output.WriteString("\tActivities []string\n\tDeliveries int\n\tValues map[string]any\n}\n\n")
	fmt.Fprintf(&output, "type %s struct {\n", runtimeBindingErrorName)
	output.WriteString("\tCode string\n\tActivity string\n\tDetail string\n}\n\n")
	fmt.Fprintf(&output, "func (err %s) Error() string {\n", runtimeBindingErrorName)
	output.WriteString("\tif err.Activity == \"\" {\n\t\treturn err.Code + \": \" + err.Detail\n\t}\n\treturn err.Code + \": \" + err.Activity + \": \" + err.Detail\n}\n\n")

	output.WriteString("var goooRuntimeActivities = map[string]any{\n")
	for _, activity := range ir.Activities {
		fmt.Fprintf(&output, "\t%s: %s,\n", goString(activity.ID), activity.GoName)
	}
	output.WriteString("}\n\n")
	output.WriteString("var goooRuntimeActivityOrder = []string{\n")
	for _, activityID := range order {
		fmt.Fprintf(&output, "\t%s,\n", goString(activityID))
	}
	output.WriteString("}\n\n")

	fmt.Fprintf(&output, "func goooRuntimeInvoke(activityID string, input any) (any, bool, error) {\n")
	output.WriteString("\tfunction, ok := goooRuntimeActivities[activityID]\n\tif !ok {\n")
	fmt.Fprintf(&output, "\t\treturn nil, false, %s{Code: \"RUNTIME_ACTIVITY_UNKNOWN\", Activity: activityID, Detail: \"activity is not registered\"}\n\t}\n", runtimeBindingErrorName)
	fmt.Fprintf(&output, "\tvalue := %s.ValueOf(function)\n\tif value.Kind() != %s.Func {\n", reflection, reflection)
	fmt.Fprintf(&output, "\t\treturn nil, false, %s{Code: \"RUNTIME_ACTIVITY_INVALID\", Activity: activityID, Detail: \"registered value is not a function\"}\n\t}\n", runtimeBindingErrorName)
	fmt.Fprintf(&output, "\tfunctionType := value.Type()\n\tif functionType.NumIn() > 1 || functionType.NumOut() > 1 {\n")
	fmt.Fprintf(&output, "\t\treturn nil, false, %s{Code: \"RUNTIME_SIGNATURE_UNSUPPORTED\", Activity: activityID, Detail: \"runtime activities require at most one input and one output\"}\n\t}\n", runtimeBindingErrorName)
	output.WriteString("\tvar results []" + reflection + ".Value\n\tswitch functionType.NumIn() {\n\tcase 0:\n\t\tif input != nil {\n")
	fmt.Fprintf(&output, "\t\t\treturn nil, false, %s{Code: \"RUNTIME_ROOT_INPUT_UNEXPECTED\", Activity: activityID, Detail: \"activity declares no input\"}\n\t\t}\n", runtimeBindingErrorName)
	output.WriteString("\t\tresults = value.Call(nil)\n\tcase 1:\n\t\tif input == nil {\n")
	fmt.Fprintf(&output, "\t\t\treturn nil, false, %s{Code: \"RUNTIME_INPUT_MISSING\", Activity: activityID, Detail: \"activity input is nil\"}\n\t\t}\n", runtimeBindingErrorName)
	output.WriteString("\t\targument := " + reflection + ".ValueOf(input)\n\t\tif !argument.Type().AssignableTo(functionType.In(0)) {\n")
	fmt.Fprintf(&output, "\t\t\treturn nil, false, %s{Code: \"RUNTIME_INPUT_TYPE_MISMATCH\", Activity: activityID, Detail: \"runtime input type does not match activity input\"}\n\t\t}\n", runtimeBindingErrorName)
	output.WriteString("\t\tresults = value.Call([]" + reflection + ".Value{argument})\n\tdefault:\n")
	fmt.Fprintf(&output, "\t\treturn nil, false, %s{Code: \"RUNTIME_SIGNATURE_UNSUPPORTED\", Activity: activityID, Detail: \"runtime activity input arity is unsupported\"}\n\t}\n", runtimeBindingErrorName)
	output.WriteString("\tif len(results) == 0 {\n\t\treturn nil, false, nil\n\t}\n\treturn results[0].Interface(), true, nil\n}\n\n")

	fmt.Fprintf(&output, "func %s(rootActivity string, input any) (%s, error) {\n", runtimeBindingFunctionName, runtimeExecutionTypeName)
	fmt.Fprintf(&output, "\texecution := %s{Values: map[string]any{}}\n", runtimeExecutionTypeName)
	output.WriteString("\tif _, ok := goooRuntimeActivities[rootActivity]; !ok {\n")
	fmt.Fprintf(&output, "\t\treturn execution, %s{Code: \"RUNTIME_ROOT_UNKNOWN\", Activity: rootActivity, Detail: \"root activity is not registered\"}\n\t}\n", runtimeBindingErrorName)
	output.WriteString("\tincoming := map[string]string{}\n\toutgoing := map[string][]" + runtimeBindingTypeName + "{}\n\tfor _, binding := range " + runtimeBindingTypeName + "s {\n")
	output.WriteString("\t\tif binding.Schema == \"gooo.runtime-feedback/v1\" {\n")
	fmt.Fprintf(&output, "\t\t\treturn execution, %s{Code: \"RUNTIME_FEEDBACK_UNSUPPORTED\", Activity: binding.ConsumerActivity, Detail: \"feedback bindings require bounded iteration\"}\n\t\t}\n", runtimeBindingErrorName)
	output.WriteString("\t\tif _, exists := incoming[binding.ConsumerActivity]; exists {\n")
	fmt.Fprintf(&output, "\t\t\treturn execution, %s{Code: \"RUNTIME_INPUT_CONFLICT\", Activity: binding.ConsumerActivity, Detail: \"consumer input has multiple producers\"}\n\t\t}\n", runtimeBindingErrorName)
	output.WriteString("\t\tincoming[binding.ConsumerActivity] = binding.ProducerActivity\n\t\toutgoing[binding.ProducerActivity] = append(outgoing[binding.ProducerActivity], binding)\n\t}\n\tif _, exists := incoming[rootActivity]; exists {\n")
	fmt.Fprintf(&output, "\t\treturn execution, %s{Code: \"RUNTIME_ROOT_NOT_ROOT\", Activity: rootActivity, Detail: \"root activity has a declared predecessor\"}\n\t}\n", runtimeBindingErrorName)
	output.WriteString("\tvalues := map[string]any{rootActivity: input}\n\tfor _, activityID := range goooRuntimeActivityOrder {\n\t\tvalue, available := values[activityID]\n\t\tif !available {\n\t\t\tcontinue\n\t\t}\n\t\toutput, hasOutput, err := goooRuntimeInvoke(activityID, value)\n\t\tif err != nil {\n\t\t\treturn execution, err\n\t\t}\n\t\texecution.Activities = append(execution.Activities, activityID)\n\t\tif hasOutput {\n\t\t\tvalues[activityID] = output\n\t\t}\n\t\texecution.Values[activityID] = output\n\t\tfor _, binding := range outgoing[activityID] {\n\t\t\tif !hasOutput {\n")
	fmt.Fprintf(&output, "\t\t\t\treturn execution, %s{Code: \"RUNTIME_OUTPUT_MISSING\", Activity: activityID, Detail: \"producer has no result for a declared edge\"}\n\t\t\t}\n", runtimeBindingErrorName)
	output.WriteString("\t\t\tif _, exists := values[binding.ConsumerActivity]; exists {\n")
	fmt.Fprintf(&output, "\t\t\t\treturn execution, %s{Code: \"RUNTIME_INPUT_CONFLICT\", Activity: binding.ConsumerActivity, Detail: \"consumer input was delivered more than once\"}\n\t\t\t}\n", runtimeBindingErrorName)
	output.WriteString("\t\t\tvalues[binding.ConsumerActivity] = output\n\t\t\texecution.Deliveries++\n\t\t}\n\t}\n\tfor _, binding := range " + runtimeBindingTypeName + "s {\n\t\tif _, executed := execution.Values[binding.ConsumerActivity]; !executed {\n")
	fmt.Fprintf(&output, "\t\t\treturn execution, %s{Code: \"RUNTIME_EDGE_UNREACHABLE\", Activity: binding.ConsumerActivity, Detail: \"declared consumer was not reached from root\"}\n\t\t}\n\t}\n\treturn execution, nil\n}\n", runtimeBindingErrorName)
	return output.String()
}

func runtimeActivityOrder(ir SemanticIR) []string {
	indegree := make(map[string]int, len(ir.Activities))
	outgoing := make(map[string][]string, len(ir.Activities))
	for _, activity := range ir.Activities {
		indegree[activity.ID] = 0
	}
	for _, binding := range ir.RuntimeBindings {
		if _, exists := indegree[binding.ProducerActivity]; !exists {
			continue
		}
		if _, exists := indegree[binding.ConsumerActivity]; !exists {
			continue
		}
		outgoing[binding.ProducerActivity] = append(outgoing[binding.ProducerActivity], binding.ConsumerActivity)
		indegree[binding.ConsumerActivity]++
	}
	queue := make([]string, 0, len(indegree))
	for id, degree := range indegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	order := make([]string, 0, len(indegree))
	for len(queue) > 0 {
		sort.Strings(queue)
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)
		for _, next := range outgoing[id] {
			indegree[next]--
			if indegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	if len(order) == len(indegree) {
		return order
	}
	order = order[:0]
	for _, activity := range ir.Activities {
		order = append(order, activity.ID)
	}
	return order
}

func goString(value string) string { return strconv.Quote(value) }
