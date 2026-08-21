package main

type metricValue struct {
	ID    string `json:"metric_id"`
	Value int    `json:"value"`
}

type metaBinding struct {
	Kind            string `json:"kind"`
	Operation       string `json:"operation"`
	Route           string `json:"route"`
	IndicatorCount  int    `json:"indicator_count"`
	IndicatorDigest string `json:"indicator_digest"`
}

type subjectWitness struct {
	Space        string        `json:"space"`
	Path         string        `json:"path"`
	SubjectKind  string        `json:"subject_kind"`
	Language     string        `json:"language,omitempty"`
	Metrics      []metricValue `json:"metrics"`
	Meta         metaBinding   `json:"meta"`
	WitnessDigest string       `json:"witness_digest"`
}

func fileWitness(file fileMetric, binding metaBinding) subjectWitness {
	witness := subjectWitness{
		Space: "LOGICAL_FILE", Path: file.Path, SubjectKind: "FILE",
		Language: file.Language, Metrics: []metricValue{{ID: "lines", Value: file.Lines}},
		Meta: binding,
	}
	return sealWitness(witness)
}

func directoryWitness(space string, directory directoryMetric, binding metaBinding) subjectWitness {
	witness := subjectWitness{
		Space: space, Path: directory.Path, SubjectKind: directory.SubjectKind,
		Metrics: []metricValue{
			{ID: "direct_folders", Value: directory.DirectFolders},
			{ID: "direct_files", Value: directory.DirectFiles},
			{ID: "recursive_folders", Value: directory.RecursiveFolders},
			{ID: "recursive_files", Value: directory.RecursiveFiles},
			{ID: "go_files", Value: directory.GoFiles},
			{ID: "gooo_files", Value: directory.GoooFiles},
			{ID: "go_lines", Value: directory.GoLines},
			{ID: "gooo_lines", Value: directory.GoooLines},
		}, Meta: binding,
	}
	return sealWitness(witness)
}

func sealWitness(witness subjectWitness) subjectWitness {
	witness.WitnessDigest = ""
	witness.WitnessDigest = digestJSON(witness)
	return witness
}

func witnessKey(witness subjectWitness) string {
	return witness.Space + "\x00" + witness.Path
}
