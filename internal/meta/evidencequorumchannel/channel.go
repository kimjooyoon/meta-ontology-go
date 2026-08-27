package evidencequorumchannel

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumpolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumwire"
)

func ReadDependencies(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var result []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}
	sort.Strings(result)
	if len(result) == 0 {
		return nil, fmt.Errorf("dependency manifest is empty")
	}
	return result, nil
}

func ExecutableDigest(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return evidencequorumwire.DigestBytes(raw), nil
}

func Policy(path string, source []byte) (evidencequorumpolicy.Policy, error) {
	return evidencequorumpolicy.Parse(path, source)
}

func NewReceipt(policy evidencequorumpolicy.Policy, class, channel, head, sourcePath, sourceRaw, sourceSemantic string,
	executable string, dependencies []string, predicate string) evidencequorumwire.Receipt {
	return evidencequorumwire.Receipt{
		EvidenceClass:         class,
		Channel:               channel,
		HeadSHA:               head,
		SourcePath:            sourcePath,
		SubjectRawDigest:      sourceRaw,
		SubjectSemanticDigest: sourceSemantic,
		PolicySemanticDigest:  policy.SemanticDigest,
		ExecutableDigest:      executable,
		DependencyPaths:       append([]string(nil), dependencies...),
		DependencyDigest:      evidencequorumwire.DependencyDigest(dependencies),
		Producer:              policy.Claim.Producer,
		Consumer:              policy.Claim.Consumer,
		MetaOperation:         policy.Claim.MetaOperation,
		ProofChoice:           policy.Claim.ProofChoice,
		Predicate:             predicate,
	}
}

func Write(path string, receipt evidencequorumwire.Receipt) error {
	if path == "" {
		return fmt.Errorf("receipt output is required")
	}
	raw, err := json.Marshal(evidencequorumwire.Seal(receipt))
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
