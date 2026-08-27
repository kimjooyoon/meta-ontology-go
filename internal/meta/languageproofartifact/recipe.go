package languageproofartifact

func CanonicalRecipe() Recipe {
	return Recipe{Schema: RecipeSchema, Version: 1,
		ID: "gooo://recipe/language-proof-carrying-artifact/v1", Consumer: ConsumerID,
		Steps: []RecipeStep{
			{ID: "verify-source", Input: "source-bytes", MetaOperation: "recheck-source-digest", ProofChoice: "FOUNDATION"},
			{ID: "verify-operation", Input: "operation-receipt", MetaOperation: "recheck-operation-receipt", ProofChoice: "COHERENCE"},
			{ID: "verify-invariant", Input: "invariant-evidence", MetaOperation: "recheck-no-byte-authority", ProofChoice: "REGRESSION"},
			{ID: "grant-authority", Input: "consumer-verdict", MetaOperation: "grant-only-after-proof", ProofChoice: "COHERENCE"},
		},
	}
}
