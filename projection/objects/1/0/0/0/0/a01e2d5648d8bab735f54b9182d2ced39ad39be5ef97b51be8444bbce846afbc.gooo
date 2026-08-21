package adapter

// ObservationBinding identifies the runner invocation observed by the trace.
type ObservationBinding struct {
	Fixture   string    `json:"fixture"`
	Operation Operation `json:"operation"`
	RunID     string    `json:"run_id"`
}

// ObserverPaths are the paths captured outside the producer response.
type ObserverPaths struct {
	SourcePath string `json:"source_path"`
	OutputPath string `json:"output_path"`
	TempRoot   string `json:"temp_root"`
}

// RejectionKind identifies a transaction that must leave observed files unchanged.
type RejectionKind string

const (
	RejectionCancelled RejectionKind = "cancelled"
	RejectionClosed    RejectionKind = "closed"
)

// LstatIdentity retains metadata needed to detect replacement with equal bytes.
type LstatIdentity struct {
	Exists          bool   `json:"exists"`
	Device          string `json:"device,omitempty"`
	Inode           string `json:"inode,omitempty"`
	Mode            string `json:"mode,omitempty"`
	Size            int64  `json:"size"`
	ModTimeUnixNano int64  `json:"mtime_unix_nano"`
}

// FileObservation is an observer-owned path snapshot.
type FileObservation struct {
	Path       string        `json:"path"`
	Kind       string        `json:"kind"`
	Exists     bool          `json:"exists"`
	ByteDigest string        `json:"byte_digest,omitempty"`
	Lstat      LstatIdentity `json:"lstat"`
}

// TempArtifactSnapshot is a canonical recursive snapshot rooted at TempRoot.
type TempArtifactSnapshot struct {
	Root    LstatIdentity     `json:"root"`
	Digest  string            `json:"digest"`
	Entries []FileObservation `json:"entries"`
}

// FilesystemState contains the source, output, and temporary artifact state.
type FilesystemState struct {
	Source FileObservation      `json:"source"`
	Output FileObservation      `json:"output"`
	Temp   TempArtifactSnapshot `json:"temp"`
}

// NoWriteObservation is returned by an independent NoWriteObserver.
// The private stamp prevents a response or decoded wire payload from becoming proof.
type NoWriteObservation struct {
	Binding  ObservationBinding `json:"binding"`
	Paths    ObserverPaths      `json:"paths"`
	Workflow WorkflowBinding    `json:"workflow"`
	Mutation MutationEvidence   `json:"mutation"`
	Reason   RejectionKind      `json:"rejection_reason,omitempty"`
	Before   FilesystemState    `json:"before"`
	After    FilesystemState    `json:"after"`
	stamp    *observerStamp
}
