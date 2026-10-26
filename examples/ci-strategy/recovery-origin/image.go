package recoveryorigin

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func readImage(raw []byte, role string, out *R_ReceiptOrigin) error {
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return err
	}
	name := "pushed-image.json"
	if role == "readback" {
		name = "pulled-image.json"
	}
	var selected *zip.File
	for _, file := range reader.File {
		if file.Name == name {
			if selected != nil {
				return fmt.Errorf("duplicate artifact member %s", name)
			}
			selected = file
		}
	}
	if selected == nil {
		return fmt.Errorf("missing artifact member %s", name)
	}
	return decodeImage(selected, out)
}

func decodeImage(file *zip.File, out *R_ReceiptOrigin) error {
	r, err := file.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	raw, err := io.ReadAll(io.LimitReader(r, (1<<20)+1))
	if err != nil {
		return err
	}
	if len(raw) > 1<<20 {
		return fmt.Errorf("artifact member %s exceeds one MiB", file.Name)
	}
	var images []struct {
		ID          string
		RepoDigests []string
	}
	if err := json.Unmarshal(raw, &images); err != nil {
		return err
	}
	for _, image := range images {
		out.F_image_ids = append(out.F_image_ids, image.ID)
		out.F_subjects = append(out.F_subjects, image.RepoDigests...)
	}
	return nil
}
