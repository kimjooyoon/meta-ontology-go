package bodycodegen

import (
	"go/ast"
	"go/token"
)

func (r *valueOriginReader) declaration(value *ast.DeclStmt, env valueOriginEnv, controls []string) valueOriginEnv {
	decl, ok := value.Decl.(*ast.GenDecl)
	if !ok || decl.Tok != token.VAR {
		r.reasons["DECLARATION_FORM_UNRESOLVED"] = true
		return env
	}
	for _, spec := range decl.Specs {
		variable, ok := spec.(*ast.ValueSpec)
		if !ok {
			r.reasons["DECLARATION_FORM_UNRESOLVED"] = true
			continue
		}
		var targets []ast.Expr
		for _, name := range variable.Names {
			targets = append(targets, name)
		}
		env = r.assign(targets, variable.Values, token.DEFINE, env, controls)
	}
	return env
}

func (r *valueOriginReader) branch(value *ast.IfStmt, env valueOriginEnv, controls []string) (valueOriginEnv, bool) {
	var live bool
	env, live = r.statement(value.Init, env, controls)
	if !live {
		return env, false
	}
	id := r.node("condition", "", value.Cond, r.inputs(value.Cond, env), controls)
	if id != "" {
		r.roots[id] = true
	}
	branchControls := joinValueOriginIDs(controls, []string{id})
	left, a := r.block(value.Body, copyValueOriginEnv(env), branchControls)
	right, b := r.statement(value.Else, copyValueOriginEnv(env), branchControls)
	if a != b {
		r.reasons["EARLY_EXIT_CONTROLS_NOT_EXPANDED"] = true
	}
	if !a {
		return right, b
	}
	if !b {
		return left, a
	}
	return joinValueOriginEnvs(left, right), true
}
