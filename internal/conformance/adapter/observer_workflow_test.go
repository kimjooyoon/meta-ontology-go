package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func verifiedTestWorkflow(request Request) WorkflowBinding {
	head := strings.Repeat("b", 40)
	return WorkflowBinding{
		Status: WorkflowEvidenceVerified, Repository: "caller/repository",
		BaseSHA: strings.Repeat("a", 40), HeadSHA: head,
		EventRef: "event-002", CheckoutRef: "refs/pull/104/merge",
		Run: "run-002", Attempt: 1, ArtifactCount: 1,
		Jobs: successfulReceiptJobs(head),
	}
}

func attachVerifiedWorkflow(t *testing.T, observer *NoWriteObserver, request Request) {
	t.Helper()
	evidence, err := newVerifiedWorkflowEvidence(verifiedTestWorkflow(request))
	if err != nil {
		t.Fatal(err)
	}
	if err := observer.captureVerifiedWorkflow(evidence); err != nil {
		t.Fatal(err)
	}
}

func newBareObserver(t *testing.T, request Request) *NoWriteObserver {
	t.Helper()
	root := t.TempDir()
	tempRoot := filepath.Join(root, "tmp")
	if err := os.Mkdir(tempRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(root, "source.gooo")
	outputPath := filepath.Join(root, "output.go")
	if err := os.WriteFile(sourcePath, []byte("entity billing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outputPath, []byte("package billing\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	writeTemp(t, filepath.Join(tempRoot, "stable.tmp"), "stable")
	observer, err := NewNoWriteObserver(requestObservationBinding(request), ObserverPaths{
		SourcePath: sourcePath, OutputPath: outputPath, TempRoot: tempRoot,
	})
	if err != nil {
		t.Fatal(err)
	}
	return observer
}
