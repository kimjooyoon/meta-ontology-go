package verify

func validRelationKind(kind string) bool {
	for _, candidate := range relationKinds {
		if candidate == kind {
			return true
		}
	}
	return false
}

func splitMutation(value string) []string {
	for index, char := range value {
		if char == ':' {
			return []string{value[:index], value[index+1:]}
		}
	}
	return nil
}
