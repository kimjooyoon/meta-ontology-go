package toolchainrelease

import "testing"

func TestComputationReplayRequiresExactOriginalReceiptAndProgram(t *testing.T) {
	for _, p := range computationProfiles {
		t.Run(p.Name, func(t *testing.T) {
			f := newComputationFixture(t, p)
			saved, err := f.validate(false, boundedTextSmokeBinding{})
			if err != nil {
				t.Fatal(err)
			}
			f.Receipt["replayed_from_sha256"] = saved.Receipt
			f.Receipt["result"].(map[string]any)["replay"] = map[string]any{"model_calls": 0}
			if _, err := f.validate(true, saved); err != nil {
				t.Fatal(err)
			}
			for _, wrong := range []boundedTextSmokeBinding{{}, {Program: saved.Program, Receipt: "other"},
				{Program: "other", Receipt: saved.Receipt}} {
				if _, err := f.validate(true, wrong); err == nil {
					t.Fatal("unbound replay accepted")
				}
			}
			for _, m := range []computationMutation{{"replayed_from_sha256", nil, true},
				{"result.replay", nil, true}, {"result.replay.model_calls", nil, true}, {"result.replay.model_calls", 1, false}} {
				mutateComputationReceipt(t, f, m)
				if _, err := f.validate(true, saved); err == nil {
					t.Fatal("incomplete replay accepted")
				}
				f.Receipt["replayed_from_sha256"] = saved.Receipt
				f.Receipt["result"].(map[string]any)["replay"] = map[string]any{"model_calls": 0}
			}
		})
	}
}
