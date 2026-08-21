package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strconv"
	"strings"
)

// SourceBundleDigest returns the schema-bound SHA-256 identity of raw source
// files. File order is presentation; filename, package path, and exact bytes
// are identity. Duplicate filenames are rejected rather than merged.
func SourceBundleDigest(sources []SourceFile) (string, error) {
	canonical, err := canonicalSourceFiles(sources)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(sourceBundleSchema)
	b.WriteByte('\n')
	for _, source := range canonical {
		writeSourceField(&b, source.Filename)
		writeSourceField(&b, source.PackagePath)
		b.WriteString(strconv.Itoa(len(source.Source)))
		b.WriteByte(':')
		b.Write(source.Source)
		b.WriteByte('\n')
	}
	return semantic.StableHashString(b.String()), nil
}
func canonicalSourceFiles(sources []SourceFile) ([]SourceFile, error) {
	if len(sources) == 0 {
		return nil, adapterError(AdapterSourceConfig, "", "", "at least one source file is required")
	}
	copyOf := make([]SourceFile, len(sources))
	seen := make(map[string]struct{}, len(sources))
	for index, source := range sources {
		filename := source.Filename
		if filename == "" {
			filename = "<source>"
		}
		if _, exists := seen[filename]; exists {
			return nil, adapterError(AdapterSourceConfig, "", filename, "duplicate source filename")
		}
		seen[filename] = struct{}{}
		copyOf[index] = SourceFile{
			Filename: filename, PackagePath: source.PackagePath,
			Source: append([]byte(nil), source.Source...),
		}
	}
	sort.Slice(copyOf, func(i, j int) bool {
		if copyOf[i].Filename != copyOf[j].Filename {
			return copyOf[i].Filename < copyOf[j].Filename
		}
		return copyOf[i].PackagePath < copyOf[j].PackagePath
	})
	return copyOf, nil
}
func writeSourceField(builder *strings.Builder, value string) {
	writeLengthPrefixedField(builder, value)
}
