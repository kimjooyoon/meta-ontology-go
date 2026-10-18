package syntaxregistration

import (
	"context"
	"errors"
	"io/fs"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const firstDenominatorVersion = 22

var denominatorBoundaryIDs = [...]string{"syntax", "semantics", "binding", "use-cases", "toolchain", "release"}

// ObserveCurrentBaseline selects a version only after reading the complete
// versioned history and matching its latest syntax target to the live corpus.
func ObserveCurrentBaseline(repository fs.FS) (int, error) {
	const prefix = closureRoot + "evidence/denominator-v"
	paths, err := fs.Glob(repository, prefix+"*.json")
	if err != nil {
		return 0, failure("UNKNOWN", "observe-baseline", "REGISTRATION_HISTORY_UNREADABLE",
			"DIRECT_MISSING", "restore-denominator-history")
	}
	versions := make(map[int]bool, len(paths))
	latest := 0
	for _, path := range paths {
		part := strings.TrimSuffix(strings.TrimPrefix(path, prefix), ".json")
		version, err := strconv.Atoi(part)
		if err != nil || version < firstDenominatorVersion || version > 10000 ||
			strconv.Itoa(version) != part || versions[version] {
			return 0, failure("REFUTED", "observe-baseline:"+path, "REGISTRATION_HISTORY_NOT_CONTIGUOUS",
				"", "repair-denominator-history")
		}
		versions[version] = true
		if version > latest {
			latest = version
		}
	}
	if latest < firstDenominatorVersion {
		return 0, failure("UNKNOWN", "observe-baseline:"+denominatorPath(firstDenominatorVersion),
			"REGISTRATION_HISTORY_NOT_CONTIGUOUS",
			"DIRECT_MISSING", "restore-denominator-history")
	}
	priorSyntax := 0
	var priorLinks [len(denominatorBoundaryIDs)]int
	for version := firstDenominatorVersion; version <= latest; version++ {
		path := denominatorPath(version)
		if !versions[version] {
			return 0, failure("UNKNOWN", "observe-baseline:"+path, "REGISTRATION_HISTORY_NOT_CONTIGUOUS",
				"DIRECT_MISSING", "restore-denominator-history")
		}
		raw, err := fs.ReadFile(repository, path)
		if err != nil {
			return 0, failure("UNKNOWN", "observe-baseline:"+path, "REGISTRATION_HISTORY_UNREADABLE",
				"DIRECT_MISSING", "restore-denominator-history")
		}
		var value denominator
		if err := decodeStrict(raw, &value); err != nil || value.Schema != "gooo/vertical-slice-boundary-denominator/v1" ||
			value.Version != version || value.DenominatorID != denominatorID(version) ||
			len(value.Boundaries) != len(denominatorBoundaryIDs) {
			return 0, failure("REFUTED", "observe-baseline:"+path, "REGISTRATION_HISTORY_CONTRADICTION",
				"", "repair-denominator-history")
		}
		links := 0
		for index, boundary := range value.Boundaries {
			if boundary.ID != denominatorBoundaryIDs[index] || boundary.Schema == "" ||
				boundary.MetaOperation == "" || boundary.Target < 1 || boundary.LinkTarget < priorLinks[index] {
				return 0, failure("REFUTED", "observe-baseline:"+path, "REGISTRATION_HISTORY_CONTRADICTION",
					"", "repair-denominator-history")
			}
			priorLinks[index] = boundary.LinkTarget
			links += boundary.LinkTarget
		}
		if value.Boundaries[0].Target < priorSyntax || links == 0 {
			return 0, failure("REFUTED", "observe-baseline:"+path, "REGISTRATION_HISTORY_CONTRADICTION",
				"", "repair-denominator-history")
		}
		priorSyntax = value.Boundaries[0].Target
	}
	corpus, err := fs.ReadFile(repository, corpusPath)
	if err != nil {
		return 0, failure("UNKNOWN", "observe-baseline:"+corpusPath, "REGISTRATION_CORPUS_UNREADABLE",
			"DIRECT_MISSING", "restore-syntax-corpus")
	}
	_, _, capability, err := corpusTotals(corpus)
	if err != nil || priorSyntax != capability {
		return 0, failure("REFUTED", "observe-baseline", "REGISTRATION_BASELINE_DENOMINATOR_MISMATCH",
			"", "restore-bound-baseline")
	}
	if _, err := fs.Stat(repository, denominatorPath(latest+1)); err == nil {
		return 0, failure("UNKNOWN", "observe-baseline", "REGISTRATION_BASELINE_VERSION_STALE",
			"STALE", "refresh-case-intent")
	} else if !errors.Is(err, fs.ErrNotExist) {
		return 0, failure("UNKNOWN", "observe-baseline", "DENOMINATOR_AVAILABILITY_UNKNOWN",
			"DIRECT_MISSING", "restore-denominator-history")
	}
	return latest, nil
}

// ComposeCurrentDraft keeps filesystem history as an observed input. Gooo still
// owns the case shape and profile decision in ComposeDraft.
func ComposeCurrentDraft(ctx context.Context, repository fs.FS, intent CaseIntent) (Request, bodycodegen.FixedValue, error) {
	version, err := ObserveCurrentBaseline(repository)
	if err != nil {
		return Request{}, bodycodegen.FixedValue{}, err
	}
	if intent.BaseVersion < 0 || intent.BaseVersion > 10000 {
		return Request{}, bodycodegen.FixedValue{}, failure("REFUTED", "observe-baseline",
			"REGISTRATION_INPUT_INVALID", "", "correct-explicit-input")
	}
	if intent.BaseVersion != 0 && intent.BaseVersion != version {
		return Request{}, bodycodegen.FixedValue{}, failure("UNKNOWN", "observe-baseline",
			"REGISTRATION_BASELINE_VERSION_STALE", "STALE", "refresh-case-intent")
	}
	intent.BaseVersion = version
	return ComposeDraft(ctx, intent)
}
