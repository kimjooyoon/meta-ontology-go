package guardedpromotion

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestPromotionJSONRequiresOneJSONFile(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	file, err := writer.Create("promotion.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte(`{"decision":"PASS"}`)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := promotionJSON(buffer.Bytes())
	if err != nil || string(data) != `{"decision":"PASS"}` {
		t.Fatalf("data=%s err=%v", data, err)
	}
}

func TestPromotionJSONRejectsEmptyArchive(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := promotionJSON(buffer.Bytes()); err == nil {
		t.Fatal("empty archive was accepted")
	}
}
