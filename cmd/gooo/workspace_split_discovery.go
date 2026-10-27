package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

type workspaceConsumerDiscovery struct {
	Root        string                                 `json:"root"`
	Candidates  string                                 `json:"candidates"`
	Definitions []workspacesplit.DefinitionObservation `json:"definitions"`
}

func prepareDiscoveredSourceSplit(options workspaceSourceSplitOptions, reader SourceReader) (workspaceSourceSplitReceipt, error) {
	receipt := workspaceSourceSplitReceipt{Schema: "gooo/workspace-source-split-receipt/v1",
		Decision: "FAIL_CLOSED", ManifestPath: options.manifestPath}
	if _, ok := reader.(OSFileReader); !ok {
		return receipt, fmt.Errorf("consumer discovery requires the filesystem source reader")
	}
	discovery, consumers, err := discoverSplitConsumers(options)
	if err != nil {
		return receipt, err
	}
	declared := options
	declared.discoveryRoot, declared.consumerWorkspaces = "", consumers
	declared.manifestPath, err = canonicalSplitManifest(options.manifestPath)
	if err != nil {
		return receipt, err
	}
	proposal, err := prepareWorkspaceSourceSplit(declared, reader)
	if err != nil {
		return receipt, err
	}
	current, currentConsumers, err := discoverSplitConsumers(options)
	if err != nil || !reflect.DeepEqual(discovery, current) || !reflect.DeepEqual(consumers, currentConsumers) {
		return receipt, fmt.Errorf("workspace consumer scope changed during preparation: %v", err)
	}
	proposal.ManifestPath, proposal.Discovery = options.manifestPath, &discovery
	return proposal, nil
}

func discoverSplitConsumers(options workspaceSourceSplitOptions) (workspaceConsumerDiscovery, []string, error) {
	root, err := filepath.Abs(options.discoveryRoot)
	if err == nil {
		root, err = filepath.EvalSymlinks(root)
	}
	if err != nil {
		return workspaceConsumerDiscovery{}, nil, err
	}
	primary, err := canonicalSplitManifest(options.manifestPath)
	if err != nil {
		return workspaceConsumerDiscovery{}, nil, err
	}
	source, err := workspaceinput.SourcePath(options.source)
	if err != nil {
		return workspaceConsumerDiscovery{}, nil, err
	}
	relative, err := filepath.Rel(root, filepath.Join(filepath.Dir(primary), source))
	if err != nil {
		return workspaceConsumerDiscovery{}, nil, err
	}
	files, err := os.OpenRoot(root)
	if err != nil {
		return workspaceConsumerDiscovery{}, nil, err
	}
	defer files.Close()
	observed, err := workspacesplit.DiscoverFiles(files.FS(), filepath.ToSlash(relative))
	if err != nil {
		return workspaceConsumerDiscovery{}, nil, err
	}
	consumers, found := []string{}, false
	for _, name := range observed.Consumers {
		address := filepath.Join(root, filepath.FromSlash(name))
		if address == primary {
			found = true
		} else {
			consumers = append(consumers, address)
		}
	}
	if !found {
		return workspaceConsumerDiscovery{}, nil, fmt.Errorf("primary workspace is not a discovered JSON consumer in this scope")
	}
	return workspaceConsumerDiscovery{Root: root, Candidates: "recursive-json-files", Definitions: observed.Definitions}, consumers, nil
}

func canonicalSplitManifest(name string) (string, error) {
	absolute, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(absolute)), nil
}
