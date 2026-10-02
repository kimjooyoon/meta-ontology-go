package receiptprojection

import "testing"

func TestProjectionCannotRewriteSourceOwnedIdentity(t *testing.T) {
	for name, change := range map[string]func(*Projection){"source-digest": func(p *Projection) { p.SourceSHA256 = "fake" }, "package": func(p *Projection) { p.Package = "changed" }, "field-id": func(p *Projection) { p.Entities[0].Fields[0].ID = "gooo://changed/id" }, "field-type": func(p *Projection) { p.Entities[0].Fields[0].GoType = "bool" }, "root": func(p *Projection) { p.Root = "Changed" }} {
		t.Run(name, func(t *testing.T) {
			p, err := Compile("receipt.gooo", fixture(t), "CompletenessReceipt")
			if err != nil {
				t.Fatal(err)
			}
			change(&p)
			if _, err := p.Go(); err == nil {
				t.Fatal("Go projection ignored source binding")
			}
			if _, err := p.JSONSchema(); err == nil {
				t.Fatal("JSON schema ignored source binding")
			}
		})
	}
	if _, err := (Projection{Package: "completeness", Root: "CompletenessReceipt"}).Go(); err == nil {
		t.Fatal("unbound synthetic projection accepted")
	}
}
