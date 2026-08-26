package artifactemit

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func parseSymbolicReaderRequest(payload []byte) (SymbolicReaderRequestDeclaration, bool, bool) {
	request := SymbolicReaderRequestDeclaration{SourceDigest: symbolicReaderBytesDigest(payload)}
	file, diagnostics := syntax.ParseFile("reader-request.gooo", string(payload))
	if diagnostics.HasErrors() || file.Package == nil || file.Namespace == nil {
		return request, false, false
	}
	request.Package = file.Package.Name
	request.Namespace = file.Namespace.Name
	activities := symbolicReaderRequestActivities(file.Decls)
	if len(activities) != 1 {
		return request, true, false
	}
	activity := activities[0]
	request.Activity = activity.Name
	request.ValueProgram = activity.ValueProgram
	operation, remainder, first := strings.Cut(activity.ValueProgram, ":")
	audience, resolution, second := strings.Cut(remainder, ":")
	if !first || !second || operation != "reader.project" || strings.Contains(resolution, ":") {
		return request, true, false
	}
	request.Audience = audience
	request.ExpectedResolution = resolution
	return request, true, true
}

func symbolicReaderRequestActivities(declarations []syntax.Declaration) []*syntax.ActivityDecl {
	activities := make([]*syntax.ActivityDecl, 0, 1)
	for _, declaration := range declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if ok && activity.ValueProgramPresent && strings.HasPrefix(activity.ValueProgram, "reader.project:") {
			activities = append(activities, activity)
		}
	}
	return activities
}

func symbolicReaderRequestAudienceKnown(value string) bool {
	return value == "USER" || value == "TOOL_AUTHOR" || value == "GOVERNOR"
}

func symbolicReaderRequestResolutionKnown(value string) bool {
	return value == "DECISION_AND_COUNTS_ONLY" ||
		value == "INDICATOR_CONTRACT_ONLY" ||
		value == "SOURCE_BOUND_RECEIPT_ONLY"
}
