package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/ciplanusecase"
	"github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"
)

type options struct {
	contract   string
	source     string
	generatedA string
	generatedB string
	reports    string
	replays    string
	golden     string
	profile    string
	output     string
	check      bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	settings, err := parse(args)
	if err != nil {
		return err
	}
	if settings.check {
		raw, err := os.ReadFile(settings.output)
		if err != nil {
			return err
		}
		report, err := ciplanusecase.DecodeReport(raw)
		if err != nil {
			return err
		}
		return ciplanusecase.Validate(report)
	}
	contractRaw, err := os.ReadFile(settings.contract)
	if err != nil {
		return err
	}
	contract, err := ciplanusecase.DecodeContract(contractRaw)
	if err != nil {
		return err
	}
	reports, err := readReports(settings.reports, contract)
	if err != nil {
		return err
	}
	replays, err := readReports(settings.replays, contract)
	if err != nil {
		return err
	}
	goldens, err := readGoldens(settings.golden, contract)
	if err != nil {
		return err
	}
	profile, err := readProfile(settings.profile)
	if err != nil {
		return err
	}
	source, err := sourceProfile(filepath.Dir(settings.source))
	if err != nil {
		return err
	}
	generatedReplay, err := equalGeneratedGo(settings.generatedA, settings.generatedB)
	if err != nil {
		return err
	}
	report := ciplanusecase.Evaluate(ciplanusecase.Input{
		Contract: contract, Reports: reports, Replays: replays, Goldens: goldens,
		Profile: profile, Source: source, GeneratedReplay: generatedReplay,
	})
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(settings.output, raw, 0o644); err != nil {
		return err
	}
	return ciplanusecase.Validate(report)
}

func parse(args []string) (options, error) {
	settings := options{}
	flags := flag.NewFlagSet("ci-plan-scorecard", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&settings.contract, "contract", "", "fixed contract")
	flags.StringVar(&settings.source, "source", "", "Gooo source")
	flags.StringVar(&settings.generatedA, "generated-a", "", "first generated directory")
	flags.StringVar(&settings.generatedB, "generated-b", "", "second generated directory")
	flags.StringVar(&settings.reports, "reports", "", "invocation reports directory")
	flags.StringVar(&settings.replays, "replays", "", "replayed reports directory")
	flags.StringVar(&settings.golden, "golden", "", "golden plans directory")
	flags.StringVar(&settings.profile, "profile", "", "resource profile")
	flags.StringVar(&settings.output, "output", "", "scorecard output")
	flags.BoolVar(&settings.check, "check", false, "validate an existing scorecard")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 || settings.output == "" {
		return options{}, fmt.Errorf("usage: ci-plan-scorecard --output <report.json> [--check | evidence flags]")
	}
	if !settings.check && (settings.contract == "" || settings.source == "" || settings.generatedA == "" || settings.generatedB == "" || settings.reports == "" || settings.replays == "" || settings.golden == "" || settings.profile == "") {
		return options{}, fmt.Errorf("all evidence flags are required")
	}
	return settings, nil
}

func readReports(directory string, contract ciplanusecase.Contract) (map[string]metainvocation.Report, error) {
	reports := make(map[string]metainvocation.Report, len(contract.Cases))
	for _, spec := range contract.Cases {
		raw, err := os.ReadFile(filepath.Join(directory, spec.ID+".json"))
		if err != nil {
			return nil, err
		}
		report := metainvocation.Report{}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&report); err != nil {
			return nil, fmt.Errorf("decode %s report: %w", spec.ID, err)
		}
		if err := metainvocation.Validate(report); err != nil {
			return nil, fmt.Errorf("validate %s report: %w", spec.ID, err)
		}
		reports[spec.ID] = report
	}
	return reports, nil
}

func readGoldens(directory string, contract ciplanusecase.Contract) (map[string]ciplanusecase.GoldenPlan, error) {
	goldens := map[string]ciplanusecase.GoldenPlan{}
	for _, spec := range contract.Cases {
		if spec.ExpectedDecision != metainvocation.DecisionPass {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(directory, spec.ID+".json"))
		if err != nil {
			return nil, err
		}
		golden, err := ciplanusecase.DecodeGolden(raw)
		if err != nil {
			return nil, err
		}
		goldens[spec.ID] = golden
	}
	return goldens, nil
}

func readProfile(path string) (ciplanusecase.Profile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ciplanusecase.Profile{}, err
	}
	profile := ciplanusecase.Profile{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&profile); err != nil {
		return ciplanusecase.Profile{}, err
	}
	if profile.Schema != "gooo/ci-plan-resource-profile/v1" {
		return ciplanusecase.Profile{}, fmt.Errorf("resource profile schema is invalid")
	}
	return profile, nil
}

func sourceProfile(root string) (ciplanusecase.SourceProfile, error) {
	profile := ciplanusecase.SourceProfile{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		extension := filepath.Ext(path)
		if extension != ".gooo" && extension != ".go" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lines := bytes.Count(raw, []byte{'\n'})
		if len(raw) != 0 && raw[len(raw)-1] != '\n' {
			lines++
		}
		if extension == ".gooo" {
			profile.GoooFiles++
			profile.GoooLines += lines
		} else {
			profile.GoFiles++
			profile.GoLines += lines
		}
		return nil
	})
	return profile, err
}

func equalGeneratedGo(first, second string) (bool, error) {
	left, err := generatedGoFiles(first)
	if err != nil {
		return false, err
	}
	right, err := generatedGoFiles(second)
	if err != nil {
		return false, err
	}
	if len(left) == 0 || len(left) != len(right) {
		return false, nil
	}
	for path, raw := range left {
		if !bytes.Equal(raw, right[path]) {
			return false, nil
		}
	}
	return true, nil
}

func generatedGoFiles(root string) (map[string][]byte, error) {
	files := map[string][]byte{}
	paths := []string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".go") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil, err
		}
		files[filepath.ToSlash(relative)] = raw
	}
	return files, nil
}
