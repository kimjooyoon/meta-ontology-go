// Command metric-unresolved-policy-codegen projects the source-owned disposition record.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	policyprojection "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy/projection"
	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

func main() {
	if len(os.Args) < 3 || len(os.Args) > 4 {
		fmt.Fprintln(os.Stderr, "usage: metric-unresolved-policy-codegen source.gooo output-directory [receipt-output]")
		os.Exit(2)
	}
	receipt := ""
	if len(os.Args) == 4 {
		receipt = os.Args[3]
	}
	if err := generatePolicy(os.Args[1], os.Args[2], receipt); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generatePolicy(source, directory, receipt string) error {
	raw, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	native, err := os.ReadFile(filepath.Join(filepath.Dir(source), "disposition.go"))
	if err != nil {
		return err
	}
	projection, err := policyprojection.Generate(source, raw, native)
	if err != nil {
		return err
	}
	if err := writePolicy(directory, raw, projection.Source); err != nil {
		return err
	}
	if receipt != "" {
		report, err := json.MarshalIndent(projection.Report, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(receipt, append(report, '\n'), 0600)
	}
	return nil
}

func writePolicy(directory string, source, projection []byte) error {
	parts, err := generator.PartitionProjection(projection, 75)
	if err != nil {
		return err
	}
	manifest := projectionbundle.Manifest{Schema: projectionbundle.Schema, SourceDigest: projectionbundle.Digest(source)}
	files := make(map[string][]byte, len(parts))
	for i, part := range parts {
		name := "policy.generated.go"
		if i > 0 {
			name = fmt.Sprintf("policy_part%02d.generated.go", i+1)
		}
		files[name] = part.Source
		manifest.Members = append(manifest.Members, projectionbundle.Member{
			Path: name, Digest: projectionbundle.Digest(part.Source), RegionIDs: part.RegionIDs,
		})
	}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if _, err := projectionbundle.Decode(raw); err != nil {
		return err
	}
	manifestPath := filepath.Join(directory, "policy.projection.json")
	obsolete, err := projectionbundle.ObsoleteMembers(manifestPath, manifest)
	if err != nil {
		return err
	}
	for _, member := range manifest.Members {
		if err := projectionbundle.WriteAtomic(filepath.Join(directory, member.Path), files[member.Path]); err != nil {
			return err
		}
	}
	for _, path := range obsolete {
		if err := os.Remove(filepath.Join(directory, path)); err != nil {
			return err
		}
	}
	return projectionbundle.WriteAtomic(manifestPath, append(raw, '\n'))
}
