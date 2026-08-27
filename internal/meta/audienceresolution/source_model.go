package audienceresolution

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	policyIdentityPrefix     = "gooo://audience-resolution/policy/"
	resolutionIdentityPrefix = "gooo://audience-resolution/resolution/"
	claimStateIdentityPrefix = "gooo://audience-resolution/claim-state/"
	relationIdentityPrefix   = "gooo://audience-resolution/relation/"
)

type semanticSourceModel struct {
	SemanticDigest        string
	CanonicalIRDigest     string
	DeclarationCount      int
	Audiences             []AudienceContract
	PriorClaim            string
	ClaimStates           []string
	EvidenceClaimRelation string
}

type policyCoordinate struct {
	Ordinal    int
	Coordinate string
}

// deriveSemanticSource is deliberately the producer's only source authority.
// It crosses both compiler boundaries instead of counting declaration lines or
// trusting a contract JSON copy of the policy.
func deriveSemanticSource(filename string, source []byte) (semanticSourceModel, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() || file == nil {
		return semanticSourceModel{}, fmt.Errorf("source parse failed: %v", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return semanticSourceModel{}, fmt.Errorf("source semantic lowering failed: %w", err)
	}
	model := semanticSourceModel{
		SemanticDigest:    digestBytes([]byte(ir.SemanticCanonical())),
		CanonicalIRDigest: digestBytes([]byte(ir.SemanticCanonical())),
		DeclarationCount:  len(ir.Graph.Nodes()),
		ClaimStates:       []string{},
	}
	policies := map[string][]policyCoordinate{}
	resolutions := map[string]string{}
	for _, node := range ir.Graph.Nodes() {
		id := node.ID.String()
		switch {
		case strings.HasPrefix(id, policyIdentityPrefix):
			parts := strings.Split(strings.TrimPrefix(id, policyIdentityPrefix), "/")
			if len(parts) < 3 {
				return semanticSourceModel{}, fmt.Errorf("policy identity is incomplete: %s", id)
			}
			ordinal, err := strconv.Atoi(parts[1])
			if err != nil {
				return semanticSourceModel{}, fmt.Errorf("policy ordinal is invalid: %s", id)
			}
			policies[parts[0]] = append(policies[parts[0]], policyCoordinate{Ordinal: ordinal, Coordinate: strings.Join(parts[2:], "/")})
		case strings.HasPrefix(id, resolutionIdentityPrefix):
			parts := strings.Split(strings.TrimPrefix(id, resolutionIdentityPrefix), "/")
			if len(parts) < 2 {
				return semanticSourceModel{}, fmt.Errorf("resolution identity is incomplete: %s", id)
			}
			resolutions[parts[0]] = strings.Join(parts[1:], "/")
		case strings.HasPrefix(id, claimStateIdentityPrefix):
			state := strings.TrimPrefix(id, claimStateIdentityPrefix)
			if state != "OPEN" && state != "DISCHARGED" && state != "REFUTED" {
				return semanticSourceModel{}, fmt.Errorf("unknown formal claim state: %s", state)
			}
			model.ClaimStates = append(model.ClaimStates, state)
		case strings.HasPrefix(id, relationIdentityPrefix):
			if strings.TrimPrefix(id, relationIdentityPrefix) == "evidence-to-claim" {
				model.EvidenceClaimRelation = id
			}
		}
	}
	for _, audience := range []string{"USER", "TOOL_AUTHOR", "GOVERNOR"} {
		coordinates := policies[audience]
		sort.Slice(coordinates, func(i, j int) bool {
			if coordinates[i].Ordinal != coordinates[j].Ordinal {
				return coordinates[i].Ordinal < coordinates[j].Ordinal
			}
			return coordinates[i].Coordinate < coordinates[j].Coordinate
		})
		if len(coordinates) == 0 || resolutions[audience] == "" {
			return semanticSourceModel{}, fmt.Errorf("formal audience policy is incomplete for %s", audience)
		}
		values := make([]string, 0, len(coordinates))
		for _, coordinate := range coordinates {
			values = append(values, coordinate.Coordinate)
		}
		model.Audiences = append(model.Audiences, AudienceContract{Audience: audience, Resolution: resolutions[audience], Coordinates: values})
	}
	if len(model.ClaimStates) != 3 || model.EvidenceClaimRelation == "" {
		return semanticSourceModel{}, fmt.Errorf("formal claim state or evidence relation values are incomplete")
	}
	model.PriorClaim = "OPEN"
	return model, nil
}

func sourceAudience(model semanticSourceModel, audience string) AudienceContract {
	for _, value := range model.Audiences {
		if value.Audience == audience {
			return value
		}
	}
	return AudienceContract{Audience: audience}
}

func sourceCoordinates(model semanticSourceModel) []string {
	governor := sourceAudience(model, "GOVERNOR")
	return append([]string(nil), governor.Coordinates...)
}

func allCoordinates(model semanticSourceModel) []string {
	return sourceCoordinates(model)
}

func sourcePolicyValid(model semanticSourceModel) bool {
	if len(model.Audiences) != 3 {
		return false
	}
	previous := []string{}
	for _, audience := range model.Audiences {
		if len(audience.Coordinates) <= len(previous) {
			return false
		}
		for index, coordinate := range previous {
			if audience.Coordinates[index] != coordinate {
				return false
			}
		}
		previous = audience.Coordinates
	}
	return true
}

func sourceAudienceResolutionValid(model semanticSourceModel) bool {
	return sourceAudience(model, "USER").Resolution == "USER_VISIBLE_COORDINATES" &&
		sourceAudience(model, "TOOL_AUTHOR").Resolution == "TOOL_CONTRACT_COORDINATES" &&
		sourceAudience(model, "GOVERNOR").Resolution == "GOVERNOR_FULL_LEDGER"
}

func sourceReport(binding SourceBinding, model semanticSourceModel) SourceBinding {
	binding.SemanticDigest = model.SemanticDigest
	binding.DeclarationCount = model.DeclarationCount
	binding.Reconstructed = true
	return binding
}
