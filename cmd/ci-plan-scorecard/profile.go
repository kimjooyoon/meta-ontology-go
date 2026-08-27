package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/ciplanusecase"
)

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
