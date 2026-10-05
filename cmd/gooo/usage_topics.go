package main

import "maps"

var topicHelp = mergeHelpTopics(
	topicHelpGroup1,
	topicHelpGroup2,
	topicHelpGroup3,
	topicHelpGroup4,
)

func mergeHelpTopics(groups ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, group := range groups {
		maps.Copy(result, group)
	}
	return result
}
