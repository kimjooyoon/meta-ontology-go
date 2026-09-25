package main

import (
	"encoding/json"
	"io/fs"
	"os"
)

func decodeJSON(raw []byte, target any) error {
	return json.Unmarshal(raw, target)
}

func osDirFS(root string) fs.FS {
	return os.DirFS(root)
}
