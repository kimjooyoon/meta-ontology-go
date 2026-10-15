package bodycodegen

//gooo:generated:start id="urn:gooo:body-surface:layout-facts" kind="entity"
type GoooRecord3d98480ff638481bab6aec0881a95f848aef4526fd94d47c179b206b32186b19 struct {
	GoooFieldff1783c3937d28f347d36104ed1a09a1e7e403325a8c7384b3b8aa74da8e7630 bool `json:"else_token"`
	GoooField3baec48a8e3619b2a79543ff0c4fd57bc93f019aee752cab81ac3c3561c2295c bool `json:"inserted_linebreak"`
	GoooField08d08094a8a48bc6f3dc51a4de245f376860e615451e4f8a5c1f0e4cfa8d2ec1 bool `json:"closing_brace"`
	GoooField0f7425c8dbeffa96e91d0b271f2f0cc7e398347b8332a67a838f7c840e22be46 bool `json:"whitespace_gap"`
}

//gooo:generated:end id="urn:gooo:body-surface:layout-facts" kind="entity"
//gooo:generated:start id="urn:gooo:body-surface:join-else" kind="activity"
func JoinElse(input0 bool, input1 bool, input2 bool, input3 bool) bool {

	return ClassifyElse(GoooRecord3d98480ff638481bab6aec0881a95f848aef4526fd94d47c179b206b32186b19{
		GoooFieldff1783c3937d28f347d36104ed1a09a1e7e403325a8c7384b3b8aa74da8e7630: input0,
		GoooField3baec48a8e3619b2a79543ff0c4fd57bc93f019aee752cab81ac3c3561c2295c: input1,
		GoooField08d08094a8a48bc6f3dc51a4de245f376860e615451e4f8a5c1f0e4cfa8d2ec1: input2,
		GoooField0f7425c8dbeffa96e91d0b271f2f0cc7e398347b8332a67a838f7c840e22be46: input3,
	})

}

//gooo:generated:end id="urn:gooo:body-surface:join-else" kind="activity"
//gooo:generated:start id="urn:gooo:body-surface:classify-else" kind="activity"
func ClassifyElse(input GoooRecord3d98480ff638481bab6aec0881a95f848aef4526fd94d47c179b206b32186b19) bool {

	return input.GoooFieldff1783c3937d28f347d36104ed1a09a1e7e403325a8c7384b3b8aa74da8e7630 && input.GoooField3baec48a8e3619b2a79543ff0c4fd57bc93f019aee752cab81ac3c3561c2295c && input.GoooField08d08094a8a48bc6f3dc51a4de245f376860e615451e4f8a5c1f0e4cfa8d2ec1 && input.GoooField0f7425c8dbeffa96e91d0b271f2f0cc7e398347b8332a67a838f7c840e22be46

}

//gooo:generated:end id="urn:gooo:body-surface:classify-else" kind="activity"
