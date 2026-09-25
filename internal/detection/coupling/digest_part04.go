package coupling

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strconv"
	"strings"
)

func resultCanonical(result Result) string {
	ids := append([]semantic.ID(nil), result.AcceptedSurfaceIDs...)
	slices.Sort(ids)
	reasons := append([]Reason(nil), result.Reasons...)
	sort.Slice(reasons, func(i, j int) bool {
		if reasons[i].Code != reasons[j].Code {
			return reasons[i].Code < reasons[j].Code
		}
		return reasons[i].Detail < reasons[j].Detail
	})
	var builder strings.Builder
	field(&builder, ResultSchemaV1)
	field(&builder, string(result.Status))
	field(&builder, result.InputDigest)
	for _, id := range ids {
		field(&builder, id.String())
	}
	for _, reason := range reasons {
		field(&builder, string(reason.Code))
		field(&builder, reason.Detail)
	}
	writeDimension := func(dimension CountDimension) {
		field(&builder, strconv.FormatBool(dimension.Known))
		field(&builder, strconv.FormatUint(dimension.Value, 10))
	}
	writeDimension(result.Observation.ChangedSurfaces)
	writeDimension(result.Observation.Receipts)
	writeDimension(result.Observation.InferenceRecords)
	writeDimension(result.Observation.InferencePaths)
	writeDimension(result.Observation.DeterministicWork)
	writeDimension(result.Observation.ResourceWork)
	writeDimension(result.Observation.CPU)
	writeDimension(result.Observation.Memory)
	field(&builder, strconv.FormatBool(result.FullSuiteRequired))
	return builder.String()
}
func stableDigest(value string) string { return semantic.StableHashString(value) }
func inputIdentityDigest(input Input, authority AuthorityContext) string {
	receipts := append([]CouplingReceipt(nil), input.Receipts...)
	sort.Slice(receipts, func(i, j int) bool { return receiptCanonical(receipts[i]) < receiptCanonical(receipts[j]) })
	var builder strings.Builder
	field(&builder, InputSchemaV1)
	field(&builder, input.Schema)
	field(&builder, authorityCanonical(authority))
	field(&builder, configCanonical(input.Config))
	field(&builder, registryCanonical(input.Registry))
	field(&builder, input.Registry.Digest)
	field(&builder, manifestCanonical(input.Manifest))
	field(&builder, input.Manifest.Digest)
	for _, receipt := range receipts {
		field(&builder, receiptCanonical(receipt))
	}
	field(&builder, input.InferencePath.Canonical())
	if input.ExternalReceipt != nil {
		field(&builder, externalCanonical(*input.ExternalReceipt))
		field(&builder, input.ExternalReceipt.Digest)
	}
	return stableDigest(builder.String())
}
