package main

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

type packageInputOrigins struct {
	Construction *constructionInputEvidence
	Native       *bodyexecution.NativeOriginInput
}

func readPackageExecuteInput(ctx context.Context, reader SourceReader, flags map[string]string, entry packageruntime.EntrySpec) ([]byte, packageInputOrigins, error) {
	origins := packageInputOrigins{}
	path := flags["--cases"]
	for _, key := range []string{"--inputs", "--construction-receipt", "--origin-receipt"} {
		if flags[key] != "" {
			path = flags[key]
		}
	}
	raw, err := readPackageInputBytes(reader, path, flags["--construction-receipt"] != "" || flags["--origin-receipt"] != "")
	if err != nil {
		return nil, origins, err
	}
	if flags["--construction-receipt"] != "" {
		raw, origins.Construction, err = packageConstructionInputs(ctx, raw, entry)
	}
	if flags["--origin-receipt"] != "" {
		query, queryErr := packageOriginQuery(flags)
		if queryErr != nil {
			return nil, origins, queryErr
		}
		raw, origins.Native, err = packageOriginInputsWithShape(raw, entry, query, flags["--origin-shape"])
	}
	return raw, origins, err
}
