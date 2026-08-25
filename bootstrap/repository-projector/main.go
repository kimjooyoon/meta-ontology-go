package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
)

func main() {
	settings := config{}
	flag.StringVar(&settings.root, "root", ".", "exact Git checkout to project")
	flag.StringVar(&settings.work, "work", "", "new directory outside the checkout")
	flag.StringVar(&settings.expectedSHA, "expected-sha", "", "required exact HEAD")
	flag.StringVar(&settings.physical, "physical-root", "", "stored repository to restore")
	flag.Parse()
	if err := run(settings); err != nil {
		log.Fatal(err)
	}
}

func run(settings config) error {
	if settings.physical != "" {
		return restorePhysical(settings)
	}
	root, err := filepath.Abs(settings.root)
	if err != nil {
		return err
	}
	work, err := filepath.Abs(settings.work)
	if err != nil {
		return err
	}
	if err := verifyGitIdentity(root, settings.expectedSHA); err != nil {
		return err
	}
	if err := prepareWork(root, work); err != nil {
		return err
	}
	paths, err := trackedPaths(root)
	if err != nil {
		return err
	}
	files, err := readTracked(root, paths)
	if err != nil {
		return err
	}
	objects := buildObjects(files)
	if err := assignBacking(objects); err != nil {
		return err
	}
	model := buildManifest(settings.expectedSHA, files, objects)
	stored, err := writeStore(work, model, objects, files)
	if err != nil {
		return err
	}
	loss, err := materialize(stored, filepath.Join(work, "materialized"), model)
	if err != nil {
		return err
	}
	topology, err := topologyFailures(stored)
	if err != nil {
		return err
	}
	report := buildEvidence(settings.expectedSHA, model, len(objects), loss, topology)
	if err := writeEvidence(work, report); err != nil {
		return err
	}
	fmt.Printf("repository-projector: tracked=%d objects=%d loss=%d observed_direct=%d direct=%d exempt_direct=%d mixed=%d\n",
		len(files), len(objects), loss, topology.ObservedDirect, topology.Direct, topology.ExemptDirect, topology.Mixed)
	return requireBlockingZero(report)
}
