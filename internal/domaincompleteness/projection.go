package domaincompleteness

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

//go:generate sh -c "env GOOO_LAYA_URL= GOOO_LAYA_API_KEY= go run ../../cmd/gooo body-codegen --activity ClassifyDomainCompleteness ../../scripts/domain-completeness/profile.gooo > status_generated.go"

// VerifyGeneratedProjection confirms that the checked-in Go body is the exact
// deterministic lowering of the classifier activity in the Gooo profile.
func VerifyGeneratedProjection(profilePath string, profile []byte) error {
	generated, err := bodycodegen.Generate(profilePath, profile, "ClassifyDomainCompleteness")
	if err != nil {
		return fmt.Errorf("generate Gooo completeness classifier: %w", err)
	}
	root, err := moduleRoot(profilePath)
	if err != nil {
		return err
	}
	projectionPath := filepath.Join(root, "internal", "domaincompleteness", "status_generated.go")
	projection, err := os.ReadFile(projectionPath)
	if err != nil {
		return fmt.Errorf("read generated Gooo completeness classifier: %w", err)
	}
	if !bytes.Equal(projection, []byte(generated.Source)) {
		return fmt.Errorf("generated Gooo completeness classifier differs from the source activity")
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
