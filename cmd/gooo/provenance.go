package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const (
	provenanceCLISchema       = "gooo/provenance/v1"
	provenancePublishUsage    = "usage: gooo provenance publish [--json] <file.gooo> --store <ledger.jsonl> --evidence <evidence.json>"
	provenanceStatusCommitted = "committed"
	provenanceStatusRejected  = "rejected"
)

type provenancePublishOptions struct {
	source   string
	store    string
	evidence string
}

func runProvenance(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	options, err := parseProvenancePublishArguments(args)
	if err != nil {
		if jsonMode {
			return writeProvenanceFailure(stdout, provenancePublishResponse{
				Schema: provenanceCLISchema, Status: provenanceStatusRejected, Records: []provenanceCLIRecord{},
			}, "cli.usage", err.Error())
		}
		fmt.Fprintln(stderr, err)
		return exitUsage
	}

	deadline := time.Now().Add(2 * commandDeadline)
	response, err := publishProvenance(options, reader, parser, deadline)
	if err != nil {
		code := provenanceErrorCode(err)
		if jsonMode {
			return writeProvenanceFailure(stdout, response, code, err.Error())
		}
		fmt.Fprintf(stderr, "gooo: provenance: %s: %v\n", code, err)
		return exitFailure
	}
	if jsonMode {
		if err := writeProvenanceResponse(stdout, response, deadline); err != nil {
			return exitFailure
		}
		return exitOK
	}
	fmt.Fprintf(stdout, "%s: records=%d store_digest=%s\n", response.Status, len(response.Records), response.StoreDigest)
	return exitOK
}

func parseProvenancePublishArguments(args []string) (provenancePublishOptions, error) {
	if len(args) == 0 {
		return provenancePublishOptions{}, errors.New(provenancePublishUsage)
	}
	if args[0] == "publish" {
		args = args[1:]
	}
	if len(args) == 0 {
		return provenancePublishOptions{}, errors.New(provenancePublishUsage)
	}
	options := provenancePublishOptions{}
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--store":
			value, next, ok := nextProvenanceValue(args, index)
			if !ok {
				return provenancePublishOptions{}, errors.New(provenancePublishUsage)
			}
			options.store = value
			index = next
		case "--evidence", "--input":
			value, next, ok := nextProvenanceValue(args, index)
			if !ok {
				return provenancePublishOptions{}, errors.New(provenancePublishUsage)
			}
			options.evidence = value
			index = next
		default:
			if strings.HasPrefix(arg, "-") || options.source != "" {
				return provenancePublishOptions{}, errors.New(provenancePublishUsage)
			}
			options.source = arg
		}
	}
	if options.source == "" || options.store == "" || options.evidence == "" {
		return provenancePublishOptions{}, errors.New(provenancePublishUsage)
	}
	return options, nil
}

func nextProvenanceValue(args []string, index int) (string, int, bool) {
	if index+1 >= len(args) || args[index+1] == "" || strings.HasPrefix(args[index+1], "-") {
		return "", index, false
	}
	return args[index+1], index + 1, true
}

func publishProvenance(options provenancePublishOptions, reader SourceReader, parser SourceParser, deadline time.Time) (provenancePublishResponse, error) {
	response := provenancePublishResponse{
		Schema: provenanceCLISchema, Status: provenanceStatusRejected, Records: []provenanceCLIRecord{},
	}
	source, err := readSourceWithDeadline(reader, options.source, remainingDeadline(deadline))
	if err != nil {
		return response, fmt.Errorf("read source: %w", err)
	}
	response.SourceDigest = semantic.StableHash(source)

	file, diagnostics, err := parseWithDeadline(parser, options.source, string(source), remainingDeadline(deadline))
	if err != nil {
		return response, fmt.Errorf("parse source: %w", err)
	}
	if diagnostics.HasErrors() {
		return response, fmt.Errorf("source diagnostics: %s", diagnostics.Error())
	}
	ir, err := lowerInspectIRWith(file, remainingDeadline(deadline), bidir.Lower)
	if err != nil {
		return response, fmt.Errorf("lower source: %w", err)
	}
	response.SemanticDigest = ir.StableHash()
	response.GraphDigest = ir.Graph.StableHash()

	evidenceBytes, err := readSourceWithDeadline(reader, options.evidence, remainingDeadline(deadline))
	if err != nil {
		return response, fmt.Errorf("read evidence: %w", err)
	}
	records, err := decodeProvenanceEvidence(evidenceBytes)
	if err != nil {
		return response, fmt.Errorf("decode evidence: %w", err)
	}
	if err := validateProvenanceEvidence(records, response.SourceDigest, response.SemanticDigest, response.GraphDigest); err != nil {
		return response, err
	}

	store := provenance.New(options.store)
	if err := store.Append(records...); err != nil {
		return response, fmt.Errorf("append provenance: %w", err)
	}
	claims := verifiedClaims(records)
	snapshot, err := store.Read(provenance.ReadOptions{
		ExpectedSourceDigest: response.SourceDigest,
		RequiredVerified:     claims,
	})
	if err != nil {
		return response, fmt.Errorf("canonical provenance reread: %w", err)
	}
	if err := validatePublishedSnapshot(snapshot, records, response.SourceDigest, response.SemanticDigest, response.GraphDigest); err != nil {
		return response, err
	}
	response.StoreDigest = snapshot.Digest
	response.Records = provenanceCLIRecords(snapshot.Records)
	response.Status = provenanceStatusCommitted
	return response, nil
}

func validateProvenanceEvidence(records []provenance.Evidence, sourceDigest, semanticDigest, graphDigest string) error {
	if len(records) == 0 {
		return errors.New("evidence is incomplete: at least one record is required")
	}
	for index := range records {
		record := &records[index]
		if strings.ToLower(strings.TrimSpace(record.SourceDigest)) != sourceDigest {
			return fmt.Errorf("evidence record %d has source digest different from authoritative input", index)
		}
		if strings.ToLower(strings.TrimSpace(record.SemanticDigest)) != semanticDigest {
			return fmt.Errorf("evidence record %d has semantic digest different from authoritative source", index)
		}
		if strings.ToLower(strings.TrimSpace(record.GraphDigest)) != graphDigest {
			return fmt.Errorf("evidence record %d has graph digest different from authoritative source", index)
		}
		parsed, err := semantic.ParseIdentity(strings.TrimSpace(record.SemanticID))
		if err != nil {
			return fmt.Errorf("evidence record %d has invalid stable semantic ID: %w", index, err)
		}
		record.SemanticID = parsed.String()
	}
	return nil
}

func verifiedClaims(records []provenance.Evidence) []provenance.VerifiedClaim {
	claims := make([]provenance.VerifiedClaim, 0, len(records))
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		if record.Status != provenance.StatusVerified {
			continue
		}
		key := record.SemanticID + "\x00" + record.SemanticDigest + "\x00" + record.GraphDigest
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		claims = append(claims, provenance.VerifiedClaim{
			SemanticID: record.SemanticID, SemanticDigest: record.SemanticDigest, GraphDigest: record.GraphDigest,
		})
	}
	return claims
}

func validatePublishedSnapshot(snapshot provenance.Snapshot, input []provenance.Evidence, sourceDigest, semanticDigest, graphDigest string) error {
	byID := make(map[string]provenance.Evidence, len(snapshot.Records))
	for _, record := range snapshot.Records {
		if record.SourceDigest != sourceDigest || record.SemanticDigest != semanticDigest || record.GraphDigest != graphDigest {
			return errors.New("canonical provenance reread is not bound to the authoritative source")
		}
		byID[record.ID] = record
	}
	for _, record := range input {
		if _, ok := byID[record.ID]; !ok {
			return fmt.Errorf("canonical provenance reread omitted event %q", record.ID)
		}
	}
	return nil
}

func provenanceCLIRecords(records []provenance.Evidence) []provenanceCLIRecord {
	result := make([]provenanceCLIRecord, 0, len(records))
	for _, record := range records {
		result = append(result, provenanceCLIRecord{
			ID: record.ID, SemanticID: record.SemanticID, Producer: record.Producer,
			Kind: string(record.Kind), Status: string(record.Status), Sequence: record.Sequence, Hash: record.Hash,
		})
	}
	return result
}

func provenanceErrorCode(err error) string {
	var conflict *provenance.ConflictError
	var chain *provenance.ChainError
	var claim *provenance.ClaimError
	var freshness *provenance.FreshnessError
	var corruption *provenance.CorruptionError
	switch {
	case errors.As(err, &conflict):
		return "provenance.conflict"
	case errors.As(err, &chain):
		return "provenance.chain-gap"
	case errors.As(err, &claim):
		return "provenance.claim-not-verified"
	case errors.As(err, &freshness):
		return "provenance.stale-source"
	case errors.As(err, &corruption):
		return "provenance.corruption"
	case strings.Contains(err.Error(), "decode evidence"), strings.Contains(err.Error(), "evidence is incomplete"):
		return "evidence.malformed"
	case strings.Contains(err.Error(), "different from authoritative"):
		return "evidence.binding"
	case strings.Contains(err.Error(), "read source"):
		return "source.read"
	case strings.Contains(err.Error(), "parse source"), strings.Contains(err.Error(), "source diagnostics"):
		return "source.parse"
	case strings.Contains(err.Error(), "lower source"):
		return "source.semantic"
	case strings.Contains(err.Error(), "read evidence"):
		return "evidence.read"
	case strings.Contains(err.Error(), "canonical provenance reread"):
		return "provenance.reread"
	default:
		return "provenance.rejected"
	}
}
