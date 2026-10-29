package bodyexecution

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const bridgeObservationFile = "gooo-native-field-local-bridge.json"

type CompositionBridgeWindow struct {
	OtherStart int `json:"other_start"`
	OtherEnd   int `json:"other_end"`
	OtherDrops int `json:"other_drops"`
	StartClock int `json:"start_clock"`
	Completion int `json:"completion"`
}

type CompositionBridgeTrace struct {
	Locals       []CompositionBridgeWindow `json:"locals"`
	Fields       []CompositionBridgeWindow `json:"fields"`
	LocalClocks  []int                     `json:"local_clocks"`
	FieldClocks  []int                     `json:"field_clocks"`
	Completions  int                       `json:"completions"`
	ElementReads []int                     `json:"element_reads,omitempty"`
}

type CompositionBridgeSummary struct {
	Schema             string `json:"schema"`
	Status             string `json:"status"`
	SHA256             string `json:"sha256"`
	MetadataSHA256     string `json:"metadata_sha256"`
	FieldWindows       int    `json:"field_windows"`
	LocalWindows       int    `json:"local_windows"`
	FieldUses          int    `json:"field_uses"`
	LocalUses          int    `json:"local_uses"`
	CrossReads         int    `json:"cross_window_reads"`
	Interrupted        int    `json:"interrupted_windows"`
	Completions        int    `json:"completions"`
	Scope              string `json:"scope"`
	ElementLinks       int    `json:"element_links,omitempty"`
	MissingElements    int    `json:"missing_element_links,omitempty"`
	UnresolvedElements int    `json:"unresolved_element_origins,omitempty"`
}

type nativeBridgeFile struct {
	Schema       string  `json:"schema"`
	Locals       [][]int `json:"locals"`
	Fields       [][]int `json:"fields"`
	LocalClocks  []int   `json:"local_clocks"`
	FieldClocks  []int   `json:"field_clocks"`
	Completions  []int   `json:"completions"`
	ElementReads []int   `json:"element_reads,omitempty"`
}

type compositionBridgeRun struct {
	first []byte
	file  nativeBridgeFile
}

func (run *compositionBridgeRun) reset(root string, r *CompositionRuntime) error {
	if len(r.FieldSites) == 0 {
		return nil
	}
	err := os.Remove(filepath.Join(root, bridgeObservationFile))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (run *compositionBridgeRun) observe(attempt int, root string, r *CompositionRuntime) error {
	if len(r.FieldSites) == 0 {
		return nil
	}
	file, err := os.Open(filepath.Join(root, bridgeObservationFile))
	if err != nil {
		return err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, (16<<20)+1))
	if err != nil || len(raw) > 16<<20 {
		return fmt.Errorf("native field/local bridge unavailable or exceeds16MiB: %v", err)
	}
	if attempt != 0 {
		if !bytes.Equal(run.first, raw) {
			return fmt.Errorf("native field/local bridge rows or shared clocks differ between runs")
		}
		return nil
	}
	run.first = raw
	return decode(raw, &run.file, 16<<20)
}
