package workfrontier

import (
	"encoding/json"
)

type contractFixture struct {
	Name              string          `json:"name"`
	Input             json.RawMessage `json:"input"`
	DecodeError       bool            `json:"decode_error"`
	PermutationTest   bool            `json:"permutation_test"`
	GreedyNonmaximum  bool            `json:"greedy_nonmaximum"`
	RequiredConflicts [][]string      `json:"required_conflicts"`
	Expected          expectedResult  `json:"expected"`
}
type expectedResult struct {
	Status      string   `json:"status"`
	Quality     string   `json:"quality"`
	SelectedIDs []string `json:"selected_ids"`
	WorkIDs     []string `json:"work_ids"`
}
type oracleInput struct {
	M               int               `json:"m"`
	CPUCapacity     int               `json:"cpu_capacity"`
	RegisteredPaths []string          `json:"registered_paths"`
	Pressures       *[]oraclePressure `json:"pressures"`
}
type oraclePressure struct {
	ID           string        `json:"id"`
	WorkID       string        `json:"work_id"`
	Priority     int           `json:"priority"`
	CPU          int           `json:"cpu"`
	Prerequisite string        `json:"prerequisite"`
	Claims       []oracleClaim `json:"claims"`
}
type oracleClaim struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
}
type oracleResult struct {
	Status      string
	SelectedIDs []string
	WorkIDs     []string
	MaximumSize int
}
