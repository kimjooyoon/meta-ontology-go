package bodyexecution

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type nativeLocalFile struct {
	Schema       string  `json:"schema"`
	Writes       [][]int `json:"writes"`
	Uses         [][]int `json:"uses"`
	Elements     [][]int `json:"elements"`
	Frames       []int   `json:"frames"`
	FrameDropped []int   `json:"frame_dropped"`
	WriteDropped []int   `json:"write_dropped"`
	UseDropped   []int   `json:"use_dropped"`
}

type compositionLocalRun struct {
	first []byte
	file  nativeLocalFile
}

func (run *compositionLocalRun) reset(root string, r *CompositionRuntime) error {
	if len(r.LocalSites) == 0 {
		return nil
	}
	err := os.Remove(filepath.Join(root, localObservationFile))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (run *compositionLocalRun) observe(attempt int, root string, r *CompositionRuntime) error {
	if len(r.LocalSites) == 0 {
		return nil
	}
	file, err := os.Open(filepath.Join(root, localObservationFile))
	if err != nil {
		return err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, (32<<20)+1))
	if err != nil {
		return err
	}
	if len(raw) > 32<<20 {
		return fmt.Errorf("native local observations exceed 32 MiB")
	}
	if attempt != 0 {
		if !bytes.Equal(run.first, raw) {
			return fmt.Errorf("compiled scalar local origins differ between runs")
		}
		return nil
	}
	run.first = raw
	return decode(raw, &run.file, 32<<20)
}

func (run *compositionLocalRun) bind(graph compositionGraph, traces []CompositionTrace, r *CompositionRuntime) error {
	if len(r.LocalSites) == 0 {
		return nil
	}
	f, cases := run.file, len(traces)
	if f.Schema != "gooo/native-local-observation/v1" || cases < 1 || cases > 128 ||
		f.Writes == nil || f.Uses == nil || f.Elements == nil || len(f.Writes) > 4096 || len(f.Uses) > 65536 || len(f.Elements) > 65536 ||
		len(f.Frames) != cases || len(f.FrameDropped) != cases || len(f.WriteDropped) != cases || len(f.UseDropped) != cases {
		return fmt.Errorf("native local observation protocol or bounds differ")
	}
	for c := range traces {
		traces[c].LocalFrames, traces[c].LocalFrameDropped = f.Frames[c], f.FrameDropped[c]
		traces[c].LocalWriteDropped, traces[c].LocalUseDropped = f.WriteDropped[c], f.UseDropped[c]
	}
	for _, group := range []struct {
		rows  [][]int
		width int
	}{{f.Writes, 10}, {f.Uses, 4}, {f.Elements, 6}} {
		last := -1
		for _, row := range group.rows {
			if len(row) != group.width || row[0] < 0 || row[0] >= cases || row[0] < last {
				return fmt.Errorf("local observation row width or case order differs")
			}
			last = row[0]
			t := &traces[row[0]]
			switch group.width {
			case 10:
				if row[9] != 0 && row[9] != 1 {
					return fmt.Errorf("local assignment completion must be 0 or 1")
				}
				t.LocalWrites = append(t.LocalWrites, CompositionLocalWrite{row[1], row[2], row[3], row[4], row[5], row[6], row[7], row[8], row[9] == 1})
			case 4:
				reason := ""
				if row[3] == -1 {
					reason = "NO_RECORDED_ASSIGNMENT"
				}
				if row[3] == -2 {
					reason = "WRITE_LIMIT"
				}
				t.LocalUses = append(t.LocalUses, CompositionLocalUse{row[1], row[2], row[3], reason})
			case 6:
				t.LocalElements = append(t.LocalElements, CompositionLocalElementWindow{row[1], row[2], row[3], row[4], row[5]})
			}
		}
	}
	summary, err := validateLocalObservations(traces, r.LocalSites, r.LocalUseSites, r.ReadSites, r.ConstructionSites)
	if err != nil {
		return err
	}
	summary.SHA256 = digest(run.first)
	r.LocalObservation = summary
	return nil
}
