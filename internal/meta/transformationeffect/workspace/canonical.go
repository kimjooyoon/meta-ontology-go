package workspace

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
)

const Schema = "gooo/transformation-content-patch/v1"

func hashBytes(payload []byte) string {
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func hashJSON(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return hashBytes(payload)
}

func seal(patch Patch) Patch {
	patch.ChangeDigest = hashJSON(patch.Changes)
	patch.PatchDigest = ""
	patch.PatchDigest = hashJSON(patch)
	return patch
}

func Validate(patch Patch) error {
	expected := patch
	expected.ChangeDigest, expected.PatchDigest = "", ""
	if patch.Schema != Schema || len(patch.HeadSHA) != 40 || !reflect.DeepEqual(seal(expected), patch) {
		return fmt.Errorf("patch envelope is not canonical")
	}
	for _, change := range patch.Changes {
		if change.Kind != "DELETE" {
			payload, err := base64.StdEncoding.DecodeString(change.AfterContentBase64)
			if err != nil || hashBytes(payload) != change.AfterSHA256 {
				return fmt.Errorf("patch content is unbound: %s", change.Path)
			}
		}
	}
	return nil
}
