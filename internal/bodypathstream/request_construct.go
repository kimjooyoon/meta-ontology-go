package bodypathstream

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodytiming"
)

func constructStreamResponse(ctx context.Context, model Generator, request Request, execute bool) (bodycodegen.Result, pathplan.Document, error) {
	phase := bodytiming.Start(ctx, "source_plan")
	source := []byte(request.Source)
	assembly, err := bodycodegen.SourceAssembly(ctx, "stream.gooo", source, request.Activity)
	if err != nil {
		phase.End(false)
		return bodycodegen.Result{}, pathplan.Document{}, err
	}
	if bodycodegen.IsRecordAssembly(assembly) {
		phase.End(true)
		response, err := generateRecordStream(ctx, model, request, execute)
		return response, pathplan.Document{}, err
	}
	document, err := bodycodegen.DecodeSourcePathDocument(ctx, "stream.gooo", source, request.Activity, request.Document)
	phase.End(err == nil)
	if err != nil {
		return bodycodegen.Result{}, document, fmt.Errorf("decode typed document: %w", err)
	}
	phase = bodytiming.Start(ctx, "generation")
	response, err := model.Generate(ctx, "stream.gooo", source, request.Activity, document, request.Options)
	phase.End(err == nil)
	return response, document, err
}
