package metriccounterfactual

import artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"

const StateSchema = "gooo/metric-counterfactual-state/v1"

type RootPolicy struct {
	CountsApplicability   string `json:"counts_applicability"`
	TopologyApplicability string `json:"topology_applicability"`
	TopologyReason        string `json:"topology_reason"`
	ReadmeRequirement     string `json:"readme_requirement"`
}

type FileMetric struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Lines    int    `json:"lines"`
	Bytes    int    `json:"bytes"`
	Digest   string `json:"digest"`
}

type DirectoryMetric struct {
	Path             string `json:"path"`
	DirectFolders    int    `json:"direct_folders"`
	DirectFiles      int    `json:"direct_files"`
	RecursiveFolders int    `json:"recursive_folders"`
	RecursiveFiles   int    `json:"recursive_files"`
	GoFiles          int    `json:"go_files"`
	GoooFiles        int    `json:"gooo_files"`
	GoLines          int    `json:"go_lines"`
	GoooLines        int    `json:"gooo_lines"`
}

type Totals struct {
	DirectFolders    int `json:"direct_folders"`
	DirectFiles      int `json:"direct_files"`
	RecursiveFolders int `json:"recursive_folders"`
	RecursiveFiles   int `json:"recursive_files"`
	GoFiles          int `json:"go_files"`
	GoooFiles        int `json:"gooo_files"`
	GoLines          int `json:"go_lines"`
	GoooLines        int `json:"gooo_lines"`
}

type State struct {
	Schema      string            `json:"schema"`
	Files       []FileMetric      `json:"files"`
	Directories []DirectoryMetric `json:"directories"`
	Totals      Totals            `json:"totals"`
	RootPolicy  RootPolicy        `json:"root_policy"`
	Digest      string            `json:"digest"`
}

func SealState(value State) (State, error) {
	value.Digest = ""
	digest, err := artifact.Digest(value)
	value.Digest = digest
	return value, err
}

func ValidState(value State) bool {
	digest := value.Digest
	sealed, err := SealState(value)
	return err == nil && digest == sealed.Digest
}
