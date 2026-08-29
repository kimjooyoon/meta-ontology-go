package main

import "github.com/kimjooyoon/meta-ontology-go/internal/verify"

func checkSourcePolicyForRun(root, storageRoot, from, to string) error {
	if storageRoot == "" {
		storageRoot = root
	}
	if err := printSourceMetrics(root, storageRoot); err != nil {
		return err
	}
	policy := verify.DefaultLinePolicy()
	if validRevision(from) && validRevision(to) && from != to {
		return verify.CheckProjectedSourcePolicyRevision(root, storageRoot, nil, policy, from, to)
	}
	return verify.CheckProjectedSourcePolicy(root, storageRoot, nil, policy)
}
