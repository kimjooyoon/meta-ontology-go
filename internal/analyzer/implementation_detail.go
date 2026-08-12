package analyzer

import (
	"sort"
	"strings"
)

// DeferredImplementationDetail is a source-bound, non-authoritative partial
// observation. It preserves unresolved references without inventing IDs.
type DeferredImplementationDetail struct {
	Binding DeltaBinding         `json:"binding"`
	Detail  ImplementationDetail `json:"detail"`
}

func (d DeferredImplementationDetail) canonical() string {
	var builder strings.Builder
	builder.WriteString("implementation-detail\n")
	builder.WriteString(d.Binding.canonical())
	writeBindingField(&builder, d.Detail.Reference)
	writeBindingField(&builder, string(d.Detail.IdentityState))
	writeBindingField(&builder, d.Detail.Reason)
	writeSemanticSpan(&builder, semanticSpan(d.Detail.Span))
	return builder.String()
}

func deferredImplementationDetails(result SemanticAdapterResult, binding DeltaBinding) []DeferredImplementationDetail {
	details := make([]DeferredImplementationDetail, 0, len(result.ImplementationDetails))
	for _, detail := range result.ImplementationDetails {
		detail = detail.normalized()
		details = append(details, DeferredImplementationDetail{Binding: binding, Detail: detail})
	}
	sort.Slice(details, func(i, j int) bool { return details[i].canonical() < details[j].canonical() })
	return details
}

func validateDeferredImplementationDetail(detail DeferredImplementationDetail) bool {
	return detail.Detail.Reference != "" &&
		detail.Detail.IdentityState.valid() && detail.Detail.Span.Start.Offset >= 0 &&
		detail.Detail.Span.End.Offset >= detail.Detail.Span.Start.Offset
}
