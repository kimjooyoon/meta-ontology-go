package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type RegistryBinding struct {
	Operation        string `json:"operation"`
	Executor         string `json:"executor"`
	ProofChoice      string `json:"proof_choice"`
	ExecutorEntityID string `json:"executor_entity_id"`
	ProofEntityID    string `json:"proof_entity_id"`
}

type ExecutorCoverage struct {
	Operation  string `json:"operation"`
	Executor   string `json:"executor"`
	EntityID   string `json:"entity_id"`
	EntityName string `json:"entity_name"`
	Covered    bool   `json:"covered"`
}

func registrySnapshot() []RegistryBinding {
	result := make([]RegistryBinding, 0)
	for _, binding := range generation.DefaultRegistry() {
		executor := binding.Executor
		result = append(result, RegistryBinding{
			Operation: string(binding.Operation), Executor: executor,
			ProofChoice:      string(binding.ProofChoice),
			ExecutorEntityID: "executor://" + executor,
			ProofEntityID:    proofEntityID(binding.ProofChoice),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Operation != result[j].Operation {
			return result[i].Operation < result[j].Operation
		}
		return result[i].Executor < result[j].Executor
	})
	return result
}

func proofEntityID(choice generation.ProofChoice) string {
	switch choice {
	case generation.ProofFoundation:
		return "proof://foundation"
	case generation.ProofCoherence:
		return "proof://coherence"
	case generation.ProofRegress:
		return "proof://regression"
	default:
		return "proof://unknown"
	}
}

func registryDigest(registry []RegistryBinding) string {
	data, _ := json.Marshal(registry)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
