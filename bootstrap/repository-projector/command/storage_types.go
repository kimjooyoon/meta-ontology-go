package main

type config struct {
	root        string
	work        string
	expectedSHA string
	physical    string
}

type trackedFile struct {
	logical   string
	kind      string
	language  string
	mode      uint32
	lines     int
	data      []byte
	objectSHA string
	backing   string
}

type storedObject struct {
	id      string
	ext     string
	backing string
	data    []byte
}
