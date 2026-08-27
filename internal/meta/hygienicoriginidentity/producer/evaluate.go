package producer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const expectedRecords = 2

// Evaluate is the producer-side canonical observation. It never reads
// comments or annotations; records come from lowered activity value programs.
func Evaluate(files fs.FS, sourcePath string) (Report, error) {
	raw, err := fs.ReadFile(files, sourcePath)
	if err != nil {
		return Report{}, fmt.Errorf("read source: %w", err)
	}
	file, diagnostics := syntax.ParseFile(sourcePath, string(raw))
	if file == nil || diagnostics.HasErrors() {
		return Report{}, fmt.Errorf("syntax diagnostics prevent producer evaluation: %d", len(diagnostics))
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Report{}, fmt.Errorf("lower source to semantic IR: %w", err)
	}
	producers := map[string]producerValue{}
	consumers := map[string]consumerValue{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.ValueProgram == "" {
			continue
		}
		switch {
		case strings.HasPrefix(node.Name, "Produce"):
			fields, err := valueFields(node.ValueProgram, "hoi.produce", []string{"case", "spelling", "origin", "definition-scope", "use-scope"})
			if err != nil {
				return Report{}, fmt.Errorf("producer activity %q: %w", node.Name, err)
			}
			producers[fields["case"]] = producerValue{
				CaseID: fields["case"], Spelling: fields["spelling"],
				Origin: resolveOrigin(fields["origin"]), DefinitionScope: resolveScope(fields["definition-scope"]), UseScope: resolveScope(fields["use-scope"]),
			}
		case strings.HasPrefix(node.Name, "Consume"):
			fields, err := valueFields(node.ValueProgram, "hoi.resolve", []string{"case"})
			if err != nil {
				return Report{}, fmt.Errorf("consumer activity %q: %w", node.Name, err)
			}
			value := consumerValue{CaseID: fields["case"]}
			if fields["provenance"] == "missing" {
				value.Missing = true
			} else {
				value.Resolved = resolveBinding(fields["binding"])
				value.UseScope = resolveScope(fields["use-scope"])
			}
			consumers[fields["case"]] = value
		}
	}
	if len(producers) != expectedRecords || len(consumers) != expectedRecords {
		return Report{}, fmt.Errorf("producer/consumer record denominator changed")
	}
	report := Report{SourcePath: sourcePath, RawDigest: digestBytes(raw), SemanticDigest: "sha256:" + ir.StableHash(), Records: make([]Record, 0, expectedRecords)}
	for _, id := range []string{"captured", "hygienic"} {
		producer, ok := producers[id]
		if !ok {
			return Report{}, fmt.Errorf("producer record %q is missing", id)
		}
		consumer, ok := consumers[id]
		if !ok {
			return Report{}, fmt.Errorf("consumer record %q is missing", id)
		}
		if producer.Spelling == "" || producer.Origin == "" || producer.DefinitionScope == "" || producer.UseScope == "" {
			return Report{}, fmt.Errorf("producer record %q is incomplete", id)
		}
		report.Records = append(report.Records, Record{
			CaseID: id, Spelling: producer.Spelling, OriginIdentity: producer.Origin,
			DefinitionScope: producer.DefinitionScope, UseScope: producer.UseScope,
			ResolvedUseScope: consumer.UseScope, ResolvedIdentity: consumer.Resolved,
			Captured: consumer.Resolved == consumerBinding,
		})
	}
	return report, nil
}

type producerValue struct {
	CaseID          string
	Spelling        string
	Origin          string
	DefinitionScope string
	UseScope        string
}

type consumerValue struct {
	CaseID   string
	Resolved string
	UseScope string
	Missing  bool
}

const (
	consumerBinding    = "gooo://hygienic-origin-identity/consumer/binding-site"
	producerExpansion  = "gooo://hygienic-origin-identity/producer/expansion-1"
	consumerCallSite   = "gooo://hygienic-origin-identity/scope/consumer-call-site"
	freshProducerScope = "gooo://hygienic-origin-identity/scope/fresh-producer-expansion-1"
)

func valueFields(program, operation string, required []string) (map[string]string, error) {
	fields := map[string]string{}
	tokens := slices.Collect(strings.FieldsSeq(program))
	if len(tokens) == 0 || tokens[0] != operation {
		return nil, fmt.Errorf("value program operation is not %q", operation)
	}
	for _, token := range tokens[1:] {
		key, value, ok := strings.Cut(token, "=")
		if !ok || key == "" || value == "" {
			return nil, fmt.Errorf("malformed semantic value token %q", token)
		}
		fields[key] = value
	}
	for _, key := range required {
		if fields[key] == "" {
			return nil, fmt.Errorf("semantic value missing key %q", key)
		}
	}
	return fields, nil
}

func resolveOrigin(value string) string {
	if value == "consumer-binding" {
		return consumerBinding
	}
	if value == "producer-expansion-1" {
		return producerExpansion
	}
	return ""
}

func resolveScope(value string) string {
	if value == "consumer-call-site" {
		return consumerCallSite
	}
	if value == "fresh-producer-expansion-1" {
		return freshProducerScope
	}
	return ""
}

func resolveBinding(value string) string {
	if value == "consumer-binding" {
		return consumerBinding
	}
	if value == "producer-expansion-1" {
		return producerExpansion
	}
	return ""
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}
