package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/ciplanusecase"
)

func run(args []string) error {
	settings, err := parse(args)
	if err != nil {
		return err
	}
	if settings.check {
		return checkExisting(settings.output)
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
	report := ciplanusecase.Evaluate(ciplanusecase.Input{Contract: contract, Reports: reports, Replays: replays, Goldens: goldens, Profile: profile, Source: source, GeneratedReplay: generatedReplay})
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(settings.output, append(raw, '\n'), 0o644); err != nil {
		return err
	}
	return ciplanusecase.Validate(report)
}

func checkExisting(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	report, err := ciplanusecase.DecodeReport(raw)
	if err != nil {
		return err
	}
	return ciplanusecase.Validate(report)
}
