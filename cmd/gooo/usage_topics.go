package main

var topicHelp = mergeHelpTopics(
	topicHelpGroup1,
	topicHelpGroup2,
	topicHelpGroup3,
	topicHelpGroup4,
)

func mergeHelpTopics(groups ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, group := range groups {
		for topic, content := range group {
			result[topic] = content
		}
	}
	return result
}
