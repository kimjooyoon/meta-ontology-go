package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	jev "github.com/kimjooyoon/gooo-jev/gooo"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const discoverUsage = "usage: gooo discover [--json] --query <question> [--domain-contract <contract.gooo>] [--generation <body-codegen.json>] <file.gooo>"

type capabilityDiscoveryReport struct {
	Schema                 string                            `json:"schema"`
	Decision               string                            `json:"decision"`
	Query                  jev.CapabilityQueryTrail          `json:"capability_trail"`
	SourcePath             string                            `json:"source_path"`
	SourceHash             string                            `json:"source_digest"`
	Semantic               string                            `json:"semantic_digest"`
	DomainContractPath     string                            `json:"domain_contract_path,omitempty"`
	DomainContractHash     string                            `json:"domain_contract_digest,omitempty"`
	DomainContractSemantic string                            `json:"domain_contract_semantic_digest,omitempty"`
	Generation             *discoveryGeneration              `json:"generation,omitempty"`
	Receipt                *completeness.CompletenessReceipt `json:"completeness_receipt"`
}

func runDiscover(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	query := ""
	filename := ""
	domainContractPath := ""
	generationPath := ""
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--query":
			if query != "" || index+1 >= len(args) {
				return reportDiscoverUsage(jsonMode, stdout, stderr)
			}
			index++
			query = strings.TrimSpace(args[index])
			if query == "" {
				return reportDiscoverUsage(jsonMode, stdout, stderr)
			}
		case "--domain-contract":
			if domainContractPath != "" || index+1 >= len(args) {
				return reportDiscoverUsage(jsonMode, stdout, stderr)
			}
			index++
			domainContractPath = strings.TrimSpace(args[index])
			if domainContractPath == "" {
				return reportDiscoverUsage(jsonMode, stdout, stderr)
			}
		case "--generation":
			if generationPath != "" || index+1 >= len(args) {
				return reportDiscoverUsage(jsonMode, stdout, stderr)
			}
			index++
			generationPath = strings.TrimSpace(args[index])
			if generationPath == "" {
				return reportDiscoverUsage(jsonMode, stdout, stderr)
			}
		case "--help", "-h":
			_, _ = fmt.Fprintln(stdout, discoverUsage)
			return exitOK
		default:
			if strings.HasPrefix(args[index], "-") || filename != "" {
				return reportDiscoverUsage(jsonMode, stdout, stderr)
			}
			filename = args[index]
		}
	}
	if query == "" || filename == "" {
		return reportDiscoverUsage(jsonMode, stdout, stderr)
	}
	source, err := reader.ReadFile(filename)
	if err != nil {
		return reportDiscoverFailure(jsonMode, stdout, stderr, filename, "SOURCE_READ_FAILED", err.Error())
	}
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(filename, string(source), syntax.EntityFieldsV4Support())
	if diagnostics.HasErrors() || file == nil {
		message := "source parser did not produce a declaration"
		if diagnosticErr := diagnostics.Error(); diagnosticErr != nil {
			message = diagnosticErr.Error()
		}
		return reportDiscoverFailure(jsonMode, stdout, stderr, filename, "SOURCE_PARSE_FAILED", message)
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, syntax.EntityFieldsV4Support())
	if err != nil {
		return reportDiscoverFailure(jsonMode, stdout, stderr, filename, "SOURCE_LOWER_FAILED", err.Error())
	}
	inputSequences, err := discoveryInputSequences(file, ir)
	if err != nil {
		return reportDiscoverFailure(jsonMode, stdout, stderr, filename, "SOURCE_PORT_SIGNATURE_FAILED", err.Error())
	}
	domainContract, contractFailure, err := loadDiscoveryDomainContract(reader, filename, source, domainContractPath)
	if err != nil {
		return reportDiscoverFailure(jsonMode, stdout, stderr, domainContractPath, contractFailure, err.Error())
	}
	trail := jev.DiscoverCapabilityQueryTrail(query, string(source))
	if err := trail.Validate(); err != nil {
		return reportDiscoverFailure(jsonMode, stdout, stderr, filename, "CAPABILITY_TRAIL_INVALID", err.Error())
	}
	generation, err := loadDiscoveryGeneration(reader, filename, source, generationPath)
	if err != nil {
		return reportDiscoverFailure(jsonMode, stdout, stderr, generationPath, "GENERATION_REPLAY_FAILED", err.Error())
	}
	receipt := capabilityDiscoveryCompletenessReceipt(filename, source, ir, inputSequences, trail, domainContract, generation)
	if err := completeness.Validate(receipt); err != nil {
		return reportDiscoverFailure(jsonMode, stdout, stderr, filename, "COMPLETENESS_RECEIPT_INVALID", err.Error())
	}
	report := capabilityDiscoveryReport{
		Schema: "gooo/capability-discovery-report/v1", Decision: "PROGRESS", Query: trail,
		SourcePath: filename, SourceHash: "sha256:" + cache.HashBytes(source).String(),
		Semantic: ir.StableHash(), Generation: generation, Receipt: receipt,
	}
	if domainContract != nil {
		report.DomainContractPath = domainContract.Path
		report.DomainContractHash = "sha256:" + cache.HashBytes(domainContract.Source).String()
		report.DomainContractSemantic = domainContract.IR.StableHash()
	}
	if jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return exitFailure
		}
		return exitOK
	}
	if _, err := fmt.Fprintf(stdout, "capability discovery: %s (%s)\ncompleteness: PROGRESS, first unresolved: %s\nsource: %s\n",
		trail.Response.Status, trail.Intent, receipt.FirstUnresolved.ID, filename); err != nil {
		return exitFailure
	}
	if generation != nil {
		for _, dimension := range receipt.Dimensions {
			if dimension.ID == "generation_coverage" {
				if _, err := fmt.Fprintf(stdout, "generation replay: %s\ngeneration coverage: %s %d/%d\n",
					generation.ActivityID, dimension.Status, dimension.Numerator, dimension.Denominator); err != nil {
					return exitFailure
				}
			}
		}
	}
	for _, capability := range trail.Response.Capabilities {
		if _, err := fmt.Fprintf(stdout, "- %s [%s]: %s\n", capability.ID, capability.State, capability.Description); err != nil {
			return exitFailure
		}
	}
	return exitOK
}

func capabilityDiscoveryCompletenessReceipt(filename string, source []byte, ir semantic.IR, inputSequences map[semantic.ID][]semantic.ID, trail jev.CapabilityQueryTrail, domainContract *discoveryDomainContract, generation *discoveryGeneration) *completeness.CompletenessReceipt {
	sourceDigest := "sha256:" + cache.HashBytes(source).String()
	semanticDigest := ir.StableHash()
	matchStatus, matchNumerator := capabilityCatalogMatchState(trail.Response.Status)
	declarationCoverage := discoveryDeclarationCoverage(ir, inputSequences, domainContract)
	dimensions := []completeness.CompletenessDimension{
		declarationCoverage,
		{ID: "capability_discovery_observation", Status: "PASS", Numerator: 1, Denominator: 1,
			Unit: "validated deterministic JEV discovery trails", Reason: "The JEV trail validates its query, declaration observation, route and evidence digests; a match is descriptive, not proof of implementation.",
			Evidence: []string{"jev_query_digest:" + trail.Response.QueryDigest, "jev_evidence_digest:" + trail.EvidenceDigest,
				"intent:" + trail.Intent, "discovery_path:" + strings.Join(trail.DiscoveryPath, ">")}},
		{ID: "catalog_match", Status: matchStatus, Numerator: matchNumerator, Denominator: 1,
			Unit: "requested capabilities found in the JEV catalog", Reason: capabilityMatchReason(trail.Response.Status),
			Evidence: []string{"jev_status:" + string(trail.Response.Status), "query_digest:" + trail.Response.QueryDigest}},
		discoveryGenerationCoverage(ir, inputSequences, domainContract, generation),
		unknownCompletenessDimension("reverse_observation_coverage", "generated runtime observations mapped back to source", "Discovery does not execute or reverse-observe generated artifacts."),
		unknownCompletenessDimension("real_use_case_coverage", "independently replayed input and output cases", "A catalog answer is not an independent use case or behavior result."),
		{ID: "execution_boundary_coverage", Status: boundaryObservationStatus(trail), Numerator: boolInt(trail.Response.NonExecuting && trail.Response.NonAuthorizing), Denominator: 1,
			Unit: "capability discovery requests that remain non-executing and non-authorizing", Reason: "JEV discovery is a read-only description and grants neither execution nor authorization.",
			Evidence: []string{"non_executing:" + strconv.FormatBool(trail.Response.NonExecuting), "non_authorizing:" + strconv.FormatBool(trail.Response.NonAuthorizing)}},
		unknownCompletenessDimension("permission_boundary_coverage", "observed host permission profiles", "The discovery receipt does not observe operating-system permissions."),
		unknownCompletenessDimension("network_boundary_coverage", "observed external network boundaries", "The discovery receipt does not observe external network configuration."),
		{ID: "provenance_integrity", Status: "PASS", Numerator: 1, Denominator: 1,
			Unit: "source, semantic IR, optional domain contract and JEV trail identities bound in one receipt", Reason: "The receipt binds exact source bytes, normalized semantic IR, optional domain-contract bytes and semantic IR, JEV query identity and trail evidence.",
			Evidence: discoveryProvenanceEvidence(sourceDigest, semanticDigest, trail, domainContract)},
	}
	if generation != nil {
		last := &dimensions[len(dimensions)-1]
		last.Evidence = append(last.Evidence, "generation_artifact_digest:"+generation.Digest,
			"generated_digest:"+generation.GeneratedDigest, "generated_activity_id:"+generation.ActivityID)
	}
	core := make([]string, 0, len(dimensions))
	statusCounts := map[string]int{"PASS": 0, "PROGRESS": 0, "UNKNOWN": 0, "FAIL_CLOSED": 0}
	unresolved := make([]completeness.UnresolvedCompletenessClaim, 0)
	for _, dimension := range dimensions {
		core = append(core, dimension.ID)
		statusCounts[dimension.Status]++
		if dimension.Status != "PASS" {
			next := nextCapabilityReceiptOperation(dimension.ID)
			if dimension.ID == "generation_coverage" && generation != nil && dimension.Denominator == 0 {
				next = "provide_a_separate_gooo_domain_contract_with_expected_activity_signatures"
			}
			unresolved = append(unresolved, completeness.UnresolvedCompletenessClaim{
				ID: dimension.ID, Status: dimension.Status, Reason: dimension.Reason,
				NextOperation: next,
			})
		}
	}
	var first *completeness.UnresolvedCompletenessClaim
	if len(unresolved) > 0 {
		first = &unresolved[0]
	}
	receipt := &completeness.CompletenessReceipt{
		Schema: completeness.CompletenessReceiptSchema, ProfileID: "gooo-jev-capability-discovery/v1",
		Decision: "PROGRESS", DecisionBasis: "A validated source-bound discovery observation exists; declaration coverage is measured only when a separate Gooo domain contract supplies its denominator, while generation, independent runtime use cases and reverse observation remain open.",
		Scope: map[string]any{
			"receipt_declaration": completeness.ContractBinding(), "domain": "source-bound Gooo capability discovery",
			"source_path": filename, "source_digest": sourceDigest, "semantic_digest": semanticDigest,
			"domain_contract": discoveryDomainContractScope(domainContract),
			"query_digest":    trail.Response.QueryDigest, "jev_evidence_digest": trail.EvidenceDigest,
			"capability_status": trail.Response.Status, "investment_limit": map[string]any{"queries": 1, "model_calls": 0, "generation_attempts": 0},
			"excluded_scope": []string{"semantic correctness over all inputs", "generated runtime behavior", "permission authorization", "external network behavior"},
		},
		CoreDimensions: core, Dimensions: dimensions, StatusCounts: statusCounts,
		AggregateCompletenessScore: nil, FirstUnresolved: first, UnresolvedClaims: unresolved,
		NotClaimed: []string{"A JEV catalog match does not prove Gooo implements the requested behavior.",
			"Suggestions and follow-up questions are not semantic completeness evidence.",
			"This receipt records discovery only; it does not claim generation, execution or reverse observation."},
	}
	if generation != nil {
		receipt.Scope["generation"] = generation
		receipt.DecisionBasis = "Discovery and a saved generation artifact were replayed against the supplied source; generation coverage uses expected activity signatures from the separate domain contract. Independent runtime observations remain open."
		receipt.NotClaimed[2] = "Projection replay checks the saved construction and finite selection observations; native execution, independent cases and historical model calls need their own evidence."
	}
	return receipt
}

func capabilityCatalogMatchState(status jev.CapabilityQueryState) (string, int) {
	switch status {
	case jev.CapabilityQueryAvailable:
		return "PASS", 1
	case jev.CapabilityQueryDeferred:
		return "PROGRESS", 0
	default:
		return "UNKNOWN", 0
	}
}

func capabilityMatchReason(status jev.CapabilityQueryState) string {
	switch status {
	case jev.CapabilityQueryAvailable:
		return "JEV matched a descriptive catalog entry; the match does not establish that the requested behavior exists or runs."
	case jev.CapabilityQueryDeferred:
		return "JEV matched a capability request that requires an explicit external boundary."
	default:
		return "JEV did not resolve this question to a known capability catalog entry."
	}
}

func unknownCompletenessDimension(id, unit, reason string) completeness.CompletenessDimension {
	return completeness.CompletenessDimension{ID: id, Status: "UNKNOWN", Numerator: 0, Denominator: 1,
		Unit: unit, Reason: reason, Evidence: []string{"observation_not_produced_by_capability_discovery"}}
}

func boundaryObservationStatus(trail jev.CapabilityQueryTrail) string {
	if trail.Response.NonExecuting && trail.Response.NonAuthorizing {
		return "PASS"
	}
	return "FAIL_CLOSED"
}

func nextCapabilityReceiptOperation(id string) string {
	switch id {
	case "catalog_match":
		return "refine_question_or_add_a_source_bound_capability"
	case "declaration_coverage":
		return "provide_a_separate_gooo_domain_contract_with_expected_stable_declarations"
	case "execution_boundary_coverage":
		return "restore_the_non_executing_and_non_authorizing_discovery_contract"
	case "generation_coverage":
		return "generate_a_digest_bound_artifact_from_the_declaration"
	case "reverse_observation_coverage":
		return "execute_an_independent_runtime_observation_and_map_it_back_to_source"
	case "real_use_case_coverage":
		return "bind_independent_input_and_expected_output_cases"
	case "permission_boundary_coverage":
		return "record_a_source_bound_host_permission_observation"
	case "network_boundary_coverage":
		return "record_an_explicit_external_network_boundary_observation"
	default:
		return "produce_source_bound_evidence_for_this_dimension"
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func reportDiscoverUsage(jsonMode bool, stdout, stderr io.Writer) int {
	if jsonMode {
		return reportUsage(true, stdout, stderr, "discover", discoverUsage)
	}
	_, _ = fmt.Fprintln(stderr, discoverUsage)
	return exitUsage
}

func reportDiscoverFailure(jsonMode bool, stdout, stderr io.Writer, filename, code, message string) int {
	if jsonMode {
		return reportFailure(true, stdout, stderr, "discover", filename, code, message, syntax.Span{})
	}
	_, _ = fmt.Fprintf(stderr, "gooo: %s: %s: %s\n", filename, code, message)
	return exitFailure
}
