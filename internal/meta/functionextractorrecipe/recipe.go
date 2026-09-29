// Package recipe owns the generic extraction recipe authority.
package recipe

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

const manifestSchema = "gooo.function-extraction-recipes.v2"

type Manifest struct {
	Schema         string                `json:"schema"`
	Recipes        []ExtractionRecipe    `json:"recipes"`
	RetiredRecipes []RetiredRecipeRecord `json:"retired_recipes"`
}

type ExtractionRecipe struct {
	Subject   string         `json:"subject"`
	Operation string         `json:"operation"`
	Edits     []TextEdit     `json:"edits"`
	Creates   []FileCreation `json:"creates,omitempty"`
}

type TextEdit struct {
	Path string   `json:"path"`
	Old  []string `json:"old_lines"`
	New  []string `json:"new_lines"`
}

type FileCreation struct {
	Path  string   `json:"path"`
	Lines []string `json:"lines"`
}

// RetiredRecipeRecord keeps the exact recipe payload and the source identity
// against which it was retired. It is evidence only; retired recipes are never
// returned as executable extraction authority.
type RetiredRecipeRecord struct {
	Recipe       ExtractionRecipe `json:"recipe"`
	RecipeDigest string           `json:"recipe_digest"`
	SourcePath   string           `json:"source_path"`
	SourceDigest string           `json:"source_digest"`
	RetiredAt    string           `json:"retired_at_head"`
	ReasonCode   string           `json:"reason_code"`
	Reason       string           `json:"reason"`
}

//go:embed recipes.json
var embedded []byte

func Load() ([]ExtractionRecipe, error) {
	manifest, err := LoadManifest()
	if err != nil {
		return nil, err
	}
	return manifest.Recipes, nil
}

func LoadManifest() (Manifest, error) {
	var manifest Manifest
	if err := json.Unmarshal(embedded, &manifest); err != nil {
		return Manifest{}, err
	}
	if err := validateManifest(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func validateManifest(manifest Manifest) error {
	if manifest.Schema != manifestSchema {
		return fmt.Errorf("unsupported recipe manifest %q", manifest.Schema)
	}
	active, err := Index(manifest.Recipes)
	if err != nil {
		return err
	}
	retiredSubjects := make(map[string]struct{}, len(manifest.RetiredRecipes))
	for _, retired := range manifest.RetiredRecipes {
		subject := retired.Recipe.Subject
		if subject == "" || retired.SourcePath != subject {
			return fmt.Errorf("retired recipe has invalid source identity %q", subject)
		}
		if _, exists := active[subject]; exists {
			return fmt.Errorf("recipe %s is both active and retired", subject)
		}
		if _, exists := retiredSubjects[subject]; exists {
			return fmt.Errorf("duplicate retired recipe %s", subject)
		}
		retiredSubjects[subject] = struct{}{}
		if retired.Recipe.Operation == "" || (len(retired.Recipe.Edits) == 0 && len(retired.Recipe.Creates) == 0) {
			return fmt.Errorf("retired recipe %s has no operation or payload", subject)
		}
		if !isSHA256Digest(retired.SourceDigest) {
			return fmt.Errorf("retired recipe %s has invalid source digest", subject)
		}
		if !isSHA256Digest(retired.RecipeDigest) {
			return fmt.Errorf("retired recipe %s has invalid recipe digest", subject)
		}
		canonical, err := json.Marshal(retired.Recipe)
		if err != nil {
			return fmt.Errorf("marshal retired recipe %s: %w", subject, err)
		}
		digest := sha256.Sum256(canonical)
		if retired.RecipeDigest != "sha256:"+hex.EncodeToString(digest[:]) {
			return fmt.Errorf("retired recipe %s digest does not match archived payload", subject)
		}
		if !isCommitDigest(retired.RetiredAt) {
			return fmt.Errorf("retired recipe %s has invalid retirement head", subject)
		}
		if !isReasonCode(retired.ReasonCode) || strings.TrimSpace(retired.Reason) == "" {
			return fmt.Errorf("retired recipe %s lacks an explicit reason", subject)
		}
	}
	return nil
}

func isSHA256Digest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(decoded) == sha256.Size
}

func isCommitDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 20
}

func isReasonCode(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}

func Index(recipes []ExtractionRecipe) (map[string]ExtractionRecipe, error) {
	bySubject := make(map[string]ExtractionRecipe, len(recipes))
	for _, recipe := range recipes {
		if _, exists := bySubject[recipe.Subject]; exists {
			return nil, fmt.Errorf("duplicate recipe %s", recipe.Subject)
		}
		bySubject[recipe.Subject] = recipe
	}
	return bySubject, nil
}
