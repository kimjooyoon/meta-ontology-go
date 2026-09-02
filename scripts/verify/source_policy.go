package main

func checkSourcePolicyForRun(root, storageRoot string) error {
	if storageRoot == "" {
		storageRoot = root
	}
	if err := printSourceMetrics(root, storageRoot); err != nil {
		return err
	}
	return nil
}
