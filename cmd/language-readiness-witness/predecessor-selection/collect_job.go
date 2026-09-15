package main

func findProducerJob(values []workflowJob, name string) (workflowJob, int) {
	var result workflowJob
	count := 0
	for _, value := range values {
		if value.Name == name {
			result, count = value, count+1
		}
	}
	return result, count
}
