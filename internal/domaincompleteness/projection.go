package domaincompleteness

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

//go:generate sh -c "env GOOO_LAYA_URL= GOOO_LAYA_API_KEY= go run ../../cmd/gooo body-codegen --activity ClassifyDomainCompleteness ../../scripts/domain-completeness/profile.gooo > status_generated.go"
//go:generate sh -c "env GOOO_LAYA_URL= GOOO_LAYA_API_KEY= go run ../../cmd/gooo body-codegen --activity SelectDomainCompletenessOutcome ../../scripts/domain-completeness/profile.gooo > decision_generated.go"

// VerifyGeneratedProjection confirms that the checked-in Go bodies are exact
// deterministic lowerings of their activities in the Gooo profile.
func VerifyGeneratedProjection(profilePath string, profile []byte) error {
	root, err := moduleRoot(profilePath)
	if err != nil {
		return err
	}
	projections := []struct {
		activity string
		file     string
	}{
		{activity: "ClassifyDomainCompleteness", file: "status_generated.go"},
		{activity: "SelectDomainCompletenessOutcome", file: "decision_generated.go"},
	}
	for _, projection := range projections {
		generated, err := bodycodegen.Generate(profilePath, profile, projection.activity)
		if err != nil {
			return fmt.Errorf("generate Gooo completeness activity %s: %w", projection.activity, err)
		}
		projectionPath := filepath.Join(root, "internal", "domaincompleteness", projection.file)
		checkedIn, err := os.ReadFile(projectionPath)
		if err != nil {
			return fmt.Errorf("read generated Gooo completeness activity %s: %w", projection.activity, err)
		}
		if !bytes.Equal(checkedIn, []byte(generated.Source)) {
			return fmt.Errorf("generated Gooo completeness activity %s differs from its source body", projection.activity)
		}
	}
	return nil
}

func moduleRoot(sourcePath string) (string, error) {
	absolute, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", fmt.Errorf("resolve completeness profile path: %w", err)
	}
	for directory := filepath.Dir(absolute); ; directory = filepath.Dir(directory) {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("locate module root from completeness profile: %w", err)
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("locate module root from completeness profile %q", sourcePath)
		}
	}
}
