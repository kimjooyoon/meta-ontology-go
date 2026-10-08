package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func smokeSourceGraph(binary string, input BuildInput) error {
	const source = "examples/text-operations/source.gooo.fixture"
	raw, err := commandOutput(input.Root, nil, binary, "body-context", "--value-flow", "--activity", "Classify",
		"--feature-version", jointdecision.RecordGraphSharedFeatureVersion, source)
	if err != nil {
		return err
	}
	name := input.Target.ID + "-source-graph-context.json"
	if err = os.WriteFile(filepath.Join(input.OutputDir, name), raw, 0644); err != nil {
		return err
	}
	sourceDigest, _, err := digestFile(filepath.Join(input.Root, source))
	if err != nil {
		return err
	}
	if err = validateSourceGraphSmoke(raw, sourceDigest); err != nil {
		return fmt.Errorf("TOOLCHAIN_RELEASE_SOURCE_GRAPH_SMOKE: %w", err)
	}
	return nil
}

func validateSourceGraphSmoke(raw []byte, sourceDigest string) error {
	var r struct {
		SourceSHA string `json:"original_source_sha256"`
		Calls     *int   `json:"model_predictions"`
		Tests     *int   `json:"candidate_tests"`
		Context   struct {
			Status  string `json:"status"`
			Feature string `json:"feature_version"`
			Text    string `json:"text"`
			SHA     string `json:"sha256"`
		} `json:"context"`
		Flow struct {
			Status string            `json:"status"`
			Nodes  []json.RawMessage `json:"nodes"`
		} `json:"value_flow"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}
	if sourceDigest == "" || r.SourceSHA != sourceDigest || r.Calls == nil || *r.Calls != 0 ||
		r.Tests == nil || *r.Tests != 0 || r.Context.Status != "ENCODED" || r.Flow.Status != "RESOLVED" ||
		r.Context.Feature != jointdecision.RecordGraphSharedFeatureVersion ||
		r.Context.SHA != fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(r.Context.Text))) {
		return fmt.Errorf("source-only graph export identity or counts differ")
	}
	graph, err := jointdecision.DecodeRecordGraphThree(r.Context.Text)
	if err != nil {
		return err
	}
	if len(graph.Nodes) != len(r.Flow.Nodes) {
		return fmt.Errorf("model graph and source flow node counts differ")
	}
	var features [jointdecision.ThreeFeatureDim]float32
	return jointdecision.FeaturesIntoRecordGraphThree(r.Context.Text, &features)
}
