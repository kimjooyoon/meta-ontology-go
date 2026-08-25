package evidence

const (
	directIndicator          = "storage.direct-entry"
	directObservedIndicator  = "storage.direct-entry-observed"
	directUnboundIndicator   = "storage.direct-entry-unclassified"
	bootstrapDirectIndicator = "storage.bootstrap-direct-entry"
	mixedIndicator           = "storage.mixed-kind"
	maxStoredDirectEntries   = 10
	maxStoredKinds           = 1
	topologyProof            = "axiomatic-foundation"
)

func Build(sha string, entries []Entry, objects, loss int, topology Topology) Report {
	unbound, lineDebt := 0, 0
	subjects := append([]subject(nil), topology.Subjects...)
	for _, entry := range entries {
		if entry.ObjectSHA == "" || entry.Backing == "" {
			unbound++
		}
		if entry.Language != "" && entry.Lines > 75 {
			lineDebt++
			subjects = append(subjects, subject{
				Indicator: "source.line-cap-debt", Logical: entry.Logical,
				Value: entry.Lines, Limit: 75, Consumer: "logical-source-splitter",
				Operation: "split-before-storage",
			})
		}
	}
	proof := topologyProof
	unclassifiedDirect := topology.ObservedDirect - topology.Direct - topology.ExemptDirect
	return evidence{
		Schema: "gooo.repository-projection-evidence.v1", SourceSHA: sha,
		TrackedFiles: len(entries), Objects: objects, Subjects: subjects,
		Indicators: []indicator{
			{ID: "projection.roundtrip-loss", Value: loss, Limit: 0, Blocking: true,
				Consumer: "repository-materializer", Operation: "restore-logical-tree", Proof: proof},
			{ID: "projection.unbound-entry", Value: unbound, Limit: 0, Blocking: true,
				Consumer: "repository-projector", Operation: "bind-content-object", Proof: proof},
			{ID: directObservedIndicator, Value: topology.ObservedDirect, Limit: -1, Blocking: false,
				Consumer: "repository-topology-classifier", Operation: "observe-direct-object-buckets", Proof: proof},
			{ID: directIndicator, Value: topology.Direct, Limit: 0, Blocking: true,
				Consumer: "radix-sharder", Operation: "split-object-bucket", Proof: proof},
			{ID: bootstrapDirectIndicator, Value: topology.ExemptDirect, Limit: 1, Blocking: true,
				Consumer: "github-actions", Operation: "preserve-workflow-discovery", Proof: proof},
			{ID: directUnboundIndicator, Value: unclassifiedDirect, Limit: 0, Blocking: true,
				Consumer: "repository-topology-classifier", Operation: "classify-direct-object-buckets", Proof: proof},
			{ID: mixedIndicator, Value: topology.Mixed, Limit: 0, Blocking: true,
				Consumer: "radix-sharder", Operation: "separate-branch-leaf", Proof: proof},
			{ID: "source.line-cap-debt", Value: lineDebt, Limit: 0, Blocking: false,
				Consumer: "logical-source-splitter", Operation: "split-before-storage", Proof: proof},
		},
	}
}
