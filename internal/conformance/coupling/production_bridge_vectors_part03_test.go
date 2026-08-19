package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strconv"
	"strings"
)

func bridgeField(b *strings.Builder, value string) {
	b.WriteString(strconv.Itoa(len(value)))
	b.WriteByte(':')
	b.WriteString(value)
	b.WriteByte('|')
}
func bridgeHash(value string) string {
	return strings.TrimPrefix(semantic.StableHashString(value), "sha256:")
}
func bridgeRawDigest(value string) string { return strings.TrimPrefix(value, "sha256:") }
func bridgeID(value string) semantic.ID   { return semantic.MustIdentity(value) }
func bridgeSourceMapID(value string) semantic.ID {
	return semantic.MustIdentity("urn:gooo:source-map:" + value)
}
func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
func registrySurface(registry production.Registry, id string) production.Surface {
	for _, s := range registry.Surfaces {
		if s.SurfaceID.String() == id {
			return s
		}
	}
	return production.Surface{}
}
func rawBeforeBlobDigest(input Input, receipt CouplingReceipt) string {
	for _, change := range input.Changes {
		for _, binding := range input.Registry {
			if binding.RegisteredSurfaceID == receipt.SurfaceID && binding.CodeSymbolID == change.CodeSymbolID {
				return change.BeforeDigest
			}
		}
	}
	return digestText("bridge-before-blob:" + receipt.SurfaceID)
}
func rawAfterBlobDigest(input Input, receipt CouplingReceipt) string {
	for _, change := range input.Changes {
		for _, binding := range input.Registry {
			if binding.RegisteredSurfaceID == receipt.SurfaceID && binding.CodeSymbolID == change.CodeSymbolID {
				return change.AfterDigest
			}
		}
	}
	return digestText("bridge-after-blob:" + receipt.SurfaceID)
}
