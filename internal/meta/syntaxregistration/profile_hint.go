package syntaxregistration

import (
	"context"
	"encoding/json"
	"io/fs"
	"slices"
	"strings"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const ProfileHintSchema = "gooo/syntax-registration-profile-hint/v4"

var hintProfiles = registrationHintProfiles()

func registrationHintProfiles() []string {
	profiles := []string{"ordinary", "text-fields", "implicit-ports", "ordered-collections"}
	for _, profile := range syntax.SupportedEntityFieldsProfiles() {
		profiles = append(profiles, profile.ID)
	}
	return profiles
}

type ProfileReplay struct {
	Profile     string   `json:"profile"`
	Decision    string   `json:"decision"`
	Diagnostics []string `json:"diagnostics"`
}

// ProfileHintReport is advisory. Only an explicit, later CaseIntent can select
// a profile, and candidate generation replays that source independently.
type ProfileHintReport struct {
	Schema             string          `json:"schema"`
	Authority          string          `json:"authority"`
	Path               string          `json:"path"`
	SourceDigest       string          `json:"source_digest,omitempty"`
	RequiredCapability string          `json:"required_capability,omitempty"`
	Status             string          `json:"status"`
	SuggestedProfile   string          `json:"suggested_profile,omitempty"`
	CompatibleProfiles []string        `json:"compatible_profiles"`
	MatchCount         int             `json:"match_count"`
	Reason             string          `json:"reason,omitempty"`
	Profiles           []ProfileReplay `json:"profiles"`
	Exercise           ProfileExercise `json:"source_exercise"`
	ActivityID         string          `json:"activity_id,omitempty"`
	ActivitySource     string          `json:"activity_source_sha256,omitempty"`
	ProjectionDigest   string          `json:"projection_sha256,omitempty"`
}

type profileObservation struct {
	Path               string                 `json:"path"`
	SourceDigest       string                 `json:"source_digest"`
	RequiredCapability string                 `json:"required_capability"`
	ExerciseKnown      bool                   `json:"exercise_known"`
	ExerciseCount      int64                  `json:"exercise_count"`
	Profiles           []profileCompatibility `json:"profiles"`
}

type profileCompatibility struct {
	Profile  string `json:"profile"`
	Passed   bool   `json:"passed"`
	Eligible bool   `json:"eligible"`
}

type profileHintValue struct {
	Path               string `json:"path"`
	SourceDigest       string `json:"source_digest"`
	RequiredCapability string `json:"required_capability"`
	Profile            string `json:"profile"`
	Status             string `json:"status"`
	MatchCount         int    `json:"match_count"`
	ExerciseStatus     string `json:"exercise_status"`
}

type profileReplayResult struct {
	profiles   []ProfileReplay
	compatible []string
	passed     []bool
	unknown    bool
}

// InspectProfileHint pins one source before trying the exact parser profiles.
// Missing or invalid source is a diagnostic result, not a registration failure.
func InspectProfileHint(ctx context.Context, repository fs.FS, path string) ProfileHintReport {
	return InspectProfileHintForCapability(ctx, repository, path, "")
}

// InspectProfileHintForCapability narrows the advisory choice by an explicit
// caller requirement while retaining every pinned replay observation.
func InspectProfileHintForCapability(ctx context.Context, repository fs.FS,
	path, requirement string) ProfileHintReport {
	return inspectProfileHintSourceForCapability(ctx, repository, path, requirement, contractSource)
}

func inspectProfileHintSource(ctx context.Context, repository fs.FS, path string, source []byte) ProfileHintReport {
	return inspectProfileHintSourceForCapability(ctx, repository, path, "", source)
}

func readProfileHintSource(repository fs.FS, path string) ([]byte, string) {
	if !fs.ValidPath(path) || path == "." || !strings.HasSuffix(path, ".gooo") {
		return nil, "INVALID_SOURCE_PATH"
	}
	raw, err := fs.ReadFile(repository, path)
	if err != nil {
		return nil, "SOURCE_UNREADABLE"
	}
	if len(raw) > 1<<20 {
		return nil, "SOURCE_TOO_LARGE"
	}
	return raw, ""
}

func inspectProfileHintSourceForCapability(ctx context.Context, repository fs.FS, path, requirement string,
	source []byte) ProfileHintReport {
	report := ProfileHintReport{Schema: ProfileHintSchema, Authority: "ADVISORY_ONLY",
		Path: path, RequiredCapability: requirement, Status: "UNKNOWN",
		Exercise: ProfileExercise{Status: "UNKNOWN"},
		Profiles: []ProfileReplay{}, CompatibleProfiles: []string{}}
	if !validProfileRequirement(requirement) {
		report.Reason = "INVALID_REQUIRED_CAPABILITY"
		report.Exercise.Reason = report.Reason
		return report
	}
	raw, reason := readProfileHintSource(repository, path)
	if reason != "" {
		report.Reason = reason
		report.Exercise.Reason = reason
		return report
	}
	report.SourceDigest = digest(raw)
	report.Exercise = observeProfileExercise(path, raw, requirement)
	replayed := replayProfileHintSource(path, raw, requirement)
	report.Profiles, report.CompatibleProfiles = replayed.profiles, replayed.compatible
	if replayed.unknown {
		report.Reason = "PROFILE_OBSERVATION_UNKNOWN"
		return report
	}
	observation := profileObservation{Path: path, SourceDigest: report.SourceDigest,
		RequiredCapability: requirement, ExerciseKnown: report.Exercise.Known,
		ExerciseCount: int64(report.Exercise.MatchCount), Profiles: make([]profileCompatibility, len(replayed.profiles))}
	for i, profile := range replayed.profiles {
		observation.Profiles[i] = profileCompatibility{Profile: profile.Profile, Passed: replayed.passed[i],
			Eligible: profileSatisfiesRequirement(profile.Profile, requirement)}
	}
	input, _ := json.Marshal([]any{observation})
	value, err := bodycodegen.EvaluateFixedValue(ctx, "internal/meta/syntaxregistration/contract.gooo",
		source, "SuggestRegistrationProfile", input)
	if err != nil {
		report.Reason = "HINT_EVALUATION_UNAVAILABLE"
		return report
	}
	report.ActivityID, report.ActivitySource, report.ProjectionDigest =
		value.ActivityID, value.SourceSHA256, value.ProjectionSHA256
	if !validProfileHint(value, source, path, report.SourceDigest, &report) {
		report.Reason = "HINT_CONTRACT_MISMATCH"
		return report
	}
	return report
}

// replayProfileHintSource owns parser compatibility observations. The caller
// keeps source exercise and the Gooo decision separate from this host replay.
func replayProfileHintSource(path string, raw []byte, requirement string) profileReplayResult {
	pinned := fstest.MapFS{path: &fstest.MapFile{Data: raw}}
	result := profileReplayResult{profiles: []ProfileReplay{}, compatible: []string{}, passed: make([]bool, len(hintProfiles))}
	for index, profile := range hintProfiles {
		definition := languagesyntax.CaseDefinition{Path: path, Kind: languagesyntax.KindValid}
		switch profile {
		case "ordinary":
		case "text-fields":
			definition.EntityFields = true
		case "implicit-ports":
			definition.ImplicitActivityPorts = true
		case "ordered-collections":
			definition.OrderedCollections = true
		default:
			definition.EntityFieldsProfile = profile
		}
		replayResult := languagesyntax.ReplayCase(pinned, definition)
		result.profiles = append(result.profiles, ProfileReplay{Profile: profile,
			Decision: replayResult.ObservedDecision, Diagnostics: replayResult.Diagnostics})
		if replayResult.ObservedDecision == replay.DecisionUnknown {
			result.unknown = true
			return result
		}
		result.passed[index] = replayResult.ObservedDecision == replay.DecisionPass
	}
	for index, profile := range hintProfiles {
		if result.passed[index] && profileSatisfiesRequirement(profile, requirement) {
			result.compatible = append(result.compatible, profile)
		}
	}
	return result
}

func validProfileHint(value bodycodegen.FixedValue, source []byte, path, sourceDigest string,
	report *ProfileHintReport) bool {
	if value.ActivityID != "gooo://syntax-registration/suggest-profile" ||
		value.SourceSHA256 != digest(source) || value.ProjectionSHA256 == "" {
		return false
	}
	var hint profileHintValue
	if err := decodeStrict(value.Value, &hint); err != nil ||
		hint.Path != path || hint.SourceDigest != sourceDigest ||
		hint.RequiredCapability != report.RequiredCapability ||
		hint.MatchCount != len(report.CompatibleProfiles) ||
		hint.ExerciseStatus != expectedExerciseStatus(report.RequiredCapability, report.Exercise) {
		return false
	}
	if hint.Status == "NO_MATCH" && hint.Profile == "" && len(report.CompatibleProfiles) == 0 {
		report.Status = hint.Status
		report.Exercise.Status = hint.ExerciseStatus
		return true
	}
	wantStatus := "AMBIGUOUS"
	if hint.MatchCount == 1 {
		wantStatus = "HINT"
	} else if report.RequiredCapability != "" {
		wantStatus = "MINIMAL_HINT"
	}
	if hint.Status != wantStatus || hint.MatchCount == 0 ||
		!slices.Contains(report.CompatibleProfiles, hint.Profile) ||
		(report.RequiredCapability != "" && hint.Profile != report.CompatibleProfiles[0]) {
		return false
	}
	report.Status, report.SuggestedProfile, report.MatchCount =
		hint.Status, hint.Profile, hint.MatchCount
	report.Exercise.Status = hint.ExerciseStatus
	return true
}
