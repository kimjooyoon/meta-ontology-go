package bodypathstream

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodytiming"
)

type sourceAssemblyGenerator interface {
	GenerateSourceAssembly(context.Context, string, []byte, string) (bodycodegen.Result, error)
}

func generateRecordStream(ctx context.Context, model Generator, request Request, execute bool) (bodycodegen.Result, error) {
	if execute {
		return bodycodegen.Result{}, fmt.Errorf("record graph execution uses body-compose and positional typed cases")
	}
	if len(request.Document) != 0 || request.Options != (bodycodegen.TypedPathOptions{}) {
		return bodycodegen.Result{}, fmt.Errorf("record worker uses the source-owned field plan and budget")
	}
	g, ok := model.(sourceAssemblyGenerator)
	if !ok {
		return bodycodegen.Result{}, fmt.Errorf("generator does not support source record assembly")
	}
	phase := bodytiming.Start(ctx, "generation")
	result, err := g.GenerateSourceAssembly(ctx, "stream.gooo", []byte(request.Source), request.Activity)
	phase.End(err == nil)
	return result, err
}
