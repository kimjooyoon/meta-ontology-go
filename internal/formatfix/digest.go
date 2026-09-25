package formatfix

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func planDigest(plan Plan) string {
	plan.PlanDigest = ""
	raw, _ := json.Marshal(plan)
	return digestBytes(raw)
}

func seal(plan Plan) Plan {
	plan.PlanDigest = planDigest(plan)
	return plan
}
