package bodycodegen

import (
	"context"
	"fmt"
	"go/parser"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type recordAssemblyPlan struct {
	body             preparedBody
	original         preparedBody
	spec             *assemblyspec.Spec
	choices          []RecordValueChoice
	sites            []recordValueSite
	dependencySHA256 string
}

type recordValueSite struct {
	start, end int
	kind       string
	choice     RecordValueChoice
}

func IsRecordAssembly(spec *assemblyspec.Spec) bool { return spec != nil && len(spec.ValueCases) != 0 }

// ValidateSourceAssembly checks field alternatives and typed cases before any
// optional model is loaded. Source IR search validates its grammar and bodies;
// choice-based Integer assembly keeps its existing document path.
func ValidateSourceAssembly(ctx context.Context, filename string, source []byte, activity string) error {
	spec, err := SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return err
	}
	if IsSourceIRBodyFill(spec) {
		// Reuse the bounded deterministic scorer to check every assignment and
		// case before a composition loads any optional model.
		_, err = GenerateWithSourceIRBodyFill(ctx, filename, source, activity, spec, "", "", IRBodyFillOptions{})
		return err
	}
	if IsSourceIRSearch(spec) {
		return ValidateSourceIRSearch(ctx, filename, source, activity, spec)
	}
	if IsRecordAssembly(spec) {
		_, err = prepareRecordAssembly(ctx, filename, source, activity)
		return err
	}
	_, err = DecodeSourcePathDocument(ctx, filename, source, activity, nil)
	return err
}

func prepareRecordAssembly(ctx context.Context, filename string, source []byte, activity string) (recordAssemblyPlan, error) {
	var plan recordAssemblyPlan
	if ctx == nil {
		return plan, fmt.Errorf("record assembly requires a context")
	}
	if err := ctx.Err(); err != nil {
		return plan, err
	}
	var err error
	plan.spec, err = SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return plan, err
	}
	if !IsRecordAssembly(plan.spec) || plan.spec.Seed != "" {
		return plan, fmt.Errorf("record assembly requires value_case and unseeded bounded selection")
	}
	plan.original, err = prepareActivityBody(filename, source, activity)
	if err != nil {
		return plan, err
	}
	planning, err := sourceAssemblyPlanningSource(ctx, filename, source, activity)
	if err != nil {
		return plan, err
	}
	plan.body, err = prepareActivityBody(filename, planning, activity)
	if err != nil {
		return plan, err
	}
	if recordTypeByName(plan.body.records, plan.body.outputType) == nil {
		return plan, fmt.Errorf("field assembly requires a declared record result")
	}
	if err = plan.bindSites(); err != nil {
		return plan, err
	}
	if err = validateRecordAssemblyCases(ctx, plan.body.base.source, activity, plan.body.records, plan.spec.ValueCases); err != nil {
		return plan, err
	}
	if err = plan.validateAlternatives(ctx); err != nil {
		return plan, err
	}
	if plan.spec.Baseline != "" {
		mask, err := recordPickedMask(plan.spec)
		if err != nil {
			return plan, err
		}
		candidate, err := plan.candidate(mask)
		if err != nil || string(candidate.source) != string(plan.original.base.source) {
			return plan, fmt.Errorf("record computes differs from baseline and picked fields")
		}
	}
	return plan, ctx.Err()
}

func (p *recordAssemblyPlan) bindSites() error {
	sites, err := recordAssemblySites(p.body)
	if err != nil {
		return err
	}
	used := make(map[int]bool)
	for _, choice := range p.spec.Choices {
		var matching []recordValueSite
		for _, site := range sites {
			if site.kind == choice.Kind {
				matching = append(matching, site)
			}
		}
		if choice.Occurrence >= len(matching) || used[matching[choice.Occurrence].start] {
			return fmt.Errorf("field choice needs a unique %s occurrence", choice.Kind)
		}
		site := matching[choice.Occurrence]
		used[site.start] = true
		site.choice.ID, site.choice.Occurrence, site.choice.Second = choice.ID, choice.Occurrence, choice.Alternative
		site.choice.Intent = choice.Intent
		if _, err := parser.ParseExpr(choice.Alternative); err != nil {
			return fmt.Errorf("field alternative: %w", err)
		}
		p.sites, p.choices = append(p.sites, site), append(p.choices, site.choice)
	}
	return p.validateSiteSpans()
}

func (p recordAssemblyPlan) validateSiteSpans() error {
	for i, a := range p.sites {
		for _, b := range p.sites[:i] {
			if a.start < b.end && b.start < a.end {
				return fmt.Errorf("field choices must have disjoint expression spans")
			}
		}
	}
	return nil
}

func (p recordAssemblyPlan) candidateBody(mask uint16) (string, error) {
	if int(mask) >= 1<<len(p.sites) {
		return "", fmt.Errorf("record choice mask exceeds declared alternatives")
	}
	body := p.body.body
	type replacement struct {
		start, end int
		text       string
	}
	var changes []replacement
	for i, site := range p.sites {
		if mask&(1<<i) != 0 {
			changes = append(changes, replacement{site.start, site.end, "(" + site.choice.Second + ")"})
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].start > changes[j].start })
	for _, c := range changes {
		body = body[:c.start] + c.text + body[c.end:]
	}
	return body, nil
}

func (p recordAssemblyPlan) candidate(mask uint16) (generatedRoute, error) {
	body, err := p.candidateBody(mask)
	if err != nil {
		return generatedRoute{}, err
	}
	return p.body.generateBody(body, preserveRoute)
}

func (p *recordAssemblyPlan) validateAlternatives(ctx context.Context) error {
	// Disjoint sites and static pure calls make the baseline plus individual
	// alternatives cover every callable dependency, including alternative-only calls.
	var fingerprint strings.Builder
	fingerprint.WriteString(digest(p.body.base.source))
	for i := range p.sites {
		if err := ctx.Err(); err != nil {
			return err
		}
		candidate, err := p.candidate(1 << i)
		if err != nil {
			return fmt.Errorf("field choice %q: %w", p.choices[i].ID, err)
		}
		fingerprint.WriteString(digest(candidate.source))
	}
	p.dependencySHA256 = digest([]byte(fingerprint.String()))
	return nil
}

func recordPickedMask(spec *assemblyspec.Spec) (uint16, error) {
	var mask uint16
	for i, pick := range spec.Picked {
		if pick.Label == "value_second" {
			mask |= 1 << i
		} else if pick.Label != "value_first" {
			return 0, fmt.Errorf("invalid record field pick")
		}
	}
	return mask, nil
}

func recordTypeByName(records []RecordType, name string) *RecordType {
	for i := range records {
		if records[i].Name == name || records[i].GoName == name {
			return &records[i]
		}
	}
	return nil
}
