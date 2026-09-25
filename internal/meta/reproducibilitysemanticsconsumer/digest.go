package reproducibilitysemanticsconsumer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(raw)
}

func sealReceipt(receipt Receipt) Receipt {
	receipt.ReceiptDigest = ""
	receipt.ReceiptDigest = digestValue(receipt)
	return receipt
}

func sealJudgment(judgment Judgment) Judgment {
	judgment.JudgmentDigest = ""
	judgment.JudgmentDigest = digestValue(judgment)
	return judgment
}

func coordinate(numerator, denominator int) Coordinate {
	basisPoints := 0
	if denominator > 0 {
		basisPoints = numerator * 10_000 / denominator
	}
	return Coordinate{Numerator: numerator, Denominator: denominator, BasisPoints: basisPoints}
}
