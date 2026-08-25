package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const packagePartitionSchema = "gooo.go-package-partition-recipe.v1"

type packagePartitionRecipe struct {
	Schema        string             `json:"schema"`
	Subject       string             `json:"subject"`
	Moves         []packageMove      `json:"moves"`
	Creates       []packageCreate    `json:"creates"`
	Rewrites      []packageRewrite   `json:"rewrites"`
	Ranges        []packageRange     `json:"ranges"`
	ExpectedShape packageShape       `json:"expected_shape"`
}

type packageMove struct {
	Source, Destination, Package string
}

type packageCreate struct {
	Path, Content string
}

type packageRewrite struct {
	Path, Old, New string
	ExpectedCount int `json:"expected_count"`
}

type packageRange struct {
	Path, Start, End, Replacement string
}

type packageShape struct {
	BranchEntries int            `json:"branch_entries"`
	MaxEntries    int            `json:"max_entries"`
	Leaves        map[string]int `json:"leaves"`
}

func runPackagePartition(root, recipeName, expectedSHA, output string) error {
	if root == "" || recipeName == "" || expectedSHA == "" || output == "" {
		return fmt.Errorf("root, recipe, expected-sha, and output are required")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	authority, err := filepath.Abs(os.Getenv("GOOO_DISPOSABLE_ROOT"))
	if err != nil || os.Getenv("GOOO_DISPOSABLE_ROOT") == "" || authority != absolute {
		return fmt.Errorf("package partition requires the declared disposable root")
	}
	payload, err := os.ReadFile(recipeName)
	if err != nil {
		return err
	}
	var recipe packagePartitionRecipe
	if err := json.Unmarshal(payload, &recipe); err != nil {
		return err
	}
	if recipe.Schema != packagePartitionSchema || recipe.Subject == "" {
		return fmt.Errorf("unsupported package partition recipe")
	}
	writes := make(map[string]bool)
	if err := applyPackageRecipe(absolute, recipe, writes); err != nil {
		return err
	}
	if err := requirePackageShape(absolute, recipe); err != nil {
		return err
	}
	return writePackagePartitionReceipt(output, expectedSHA, recipe, writes)
}
