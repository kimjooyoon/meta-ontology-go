package operationprovenance

func splitMutation(value string) []string {
	for index, char := range value {
		if char == ':' {
			return []string{value[:index], value[index+1:]}
		}
	}
	return nil
}
