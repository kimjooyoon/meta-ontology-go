package languageconcept

import "io/fs"

func codeBindings(repository fs.FS, item Concept) (bool, []string) {
	if len(item.CodeBindings) == 0 {
		return false, []string{item.ID + ":no-code-binding"}
	}
	missing := make([]string, 0)
	for _, path := range item.CodeBindings {
		if !fs.ValidPath(path) {
			missing = append(missing, item.ID+":"+path)
			continue
		}
		if _, err := fs.Stat(repository, path); err != nil {
			missing = append(missing, item.ID+":"+path)
		}
	}
	return len(missing) == 0, missing
}

func validUseCases(cases []UseCase) bool {
	if len(cases) == 0 {
		return false
	}
	for _, item := range cases {
		if item.ID == "" || item.Trigger == "" || item.ExpectedOutcome == "" {
			return false
		}
	}
	return true
}
