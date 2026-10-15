package bodycodegen

import (
	"go/ast"
	"go/token"
	"reflect"
)

func (r *valueOriginReader) block(block *ast.BlockStmt, env valueOriginEnv, controls []string) (valueOriginEnv, bool) {
	for _, statement := range block.List {
		var live bool
		env, live = r.statement(statement, env, controls)
		if !live {
			return env, false
		}
	}
	return env, true
}

func (r *valueOriginReader) statement(stmt ast.Stmt, env valueOriginEnv, controls []string) (valueOriginEnv, bool) {
	if stmt == nil || !r.step() {
		return env, true
	}
	switch value := stmt.(type) {
	case *ast.BlockStmt:
		return r.block(value, env, controls)
	case *ast.DeclStmt:
		return r.declaration(value, env, controls), true
	case *ast.AssignStmt:
		env = r.assign(value.Lhs, value.Rhs, value.Tok, env, controls)
	case *ast.IncDecStmt:
		env = r.assign([]ast.Expr{value.X}, []ast.Expr{value.X}, value.Tok, env, controls)
	case *ast.IfStmt:
		return r.branch(value, env, controls)
	case *ast.RangeStmt:
		return r.collectionLoop(value, env, controls)
	case *ast.ForStmt:
		return r.loop(value, env, controls)
	case *ast.ReturnStmt:
		for _, expr := range value.Results {
			id := r.node("return", "", expr, r.inputs(expr, env), controls)
			if id != "" {
				r.roots[id] = true
			}
		}
		return env, false
	case *ast.ExprStmt:
		r.inputs(value.X, env)
	case *ast.EmptyStmt:
	default:
		r.reasons["STATEMENT_FORM_UNRESOLVED"] = true
	}
	return env, true
}

func (r *valueOriginReader) assign(targets, values []ast.Expr, operation token.Token, env valueOriginEnv, controls []string) valueOriginEnv {
	out := copyValueOriginEnv(env)
	dependencies := make([][]string, len(values))
	for i, value := range values {
		dependencies[i] = r.inputs(value, env)
	}
	for index, target := range targets {
		name := valueOriginRoot(target)
		if name != nil && name.Name == "_" {
			continue
		}
		if name == nil || name.Obj == nil {
			r.reasons["ASSIGNMENT_TARGET_UNRESOLVED"] = true
			continue
		}
		var at ast.Node = target
		var inputs []string
		if len(values) == len(targets) {
			at = values[index]
			inputs = dependencies[index]
		} else if len(values) > 0 {
			r.reasons["MULTIVALUE_ASSIGNMENT_UNRESOLVED"] = true
			for _, dependency := range dependencies {
				inputs = joinValueOriginIDs(inputs, dependency)
			}
		}
		kind := "assignment"
		if operation == token.DEFINE {
			kind = "declaration"
		}
		if _, direct := target.(*ast.Ident); !direct {
			kind = "aggregate_assignment"
			inputs = joinValueOriginIDs(inputs, env[name.Obj], r.inputs(target, env))
			r.reasons["AGGREGATE_ALIASES_NOT_EXPANDED"] = true
		} else if operation != token.ASSIGN && operation != token.DEFINE {
			inputs = joinValueOriginIDs(inputs, env[name.Obj])
		}
		id := r.node(kind, name.Name, at, inputs, controls)
		if id != "" {
			out[name.Obj] = joinValueOriginIDs([]string{id})
		}
	}
	return out
}

func (r *valueOriginReader) loop(loop *ast.ForStmt, env valueOriginEnv, controls []string) (valueOriginEnv, bool) {
	env, live := r.statement(loop.Init, env, controls)
	if !live {
		return env, false
	}
	r.reasons["LOOP_EXIT_PATHS_NOT_EXPANDED"] = true
	entry, current := copyValueOriginEnv(env), copyValueOriginEnv(env)
	for range 16 {
		branchControls := controls
		if loop.Cond != nil {
			id := r.node("loop_condition", "", loop.Cond, r.inputs(loop.Cond, current), controls)
			if id != "" {
				r.roots[id] = true
			}
			branchControls = joinValueOriginIDs(controls, []string{id})
		}
		next, bodyLive := r.block(loop.Body, copyValueOriginEnv(current), branchControls)
		if bodyLive {
			next, _ = r.statement(loop.Post, next, branchControls)
		} else {
			next = copyValueOriginEnv(entry)
		}
		next = joinValueOriginEnvs(entry, next)
		if reflect.DeepEqual(current, next) {
			return next, true
		}
		current = next
	}
	r.reasons["LOOP_ORIGIN_LIMIT"] = true
	return current, true
}

func (r *valueOriginReader) collectionLoop(loop *ast.RangeStmt, env valueOriginEnv, controls []string) (valueOriginEnv, bool) {
	r.reasons["COLLECTION_ELEMENTS_JOINED"] = true
	r.reasons["LOOP_EXIT_PATHS_NOT_EXPANDED"] = true
	entry, current := copyValueOriginEnv(env), copyValueOriginEnv(env)
	for range 16 {
		id := r.node("collection", "", loop.X, r.inputs(loop.X, current), controls)
		nested := copyValueOriginEnv(current)
		for _, variable := range []ast.Expr{loop.Key, loop.Value} {
			if name, ok := variable.(*ast.Ident); ok && name.Name != "_" && name.Obj != nil {
				element := r.node("collection_element", name.Name, name, []string{id}, controls)
				nested[name.Obj] = []string{element}
			}
		}
		next, live := r.block(loop.Body, nested, joinValueOriginIDs(controls, []string{id}))
		if !live {
			next = copyValueOriginEnv(entry)
		}
		for _, variable := range []ast.Expr{loop.Key, loop.Value} {
			if name, ok := variable.(*ast.Ident); ok {
				delete(next, name.Obj)
			}
		}
		next = joinValueOriginEnvs(entry, next)
		if reflect.DeepEqual(current, next) {
			return next, true
		}
		current = next
	}
	r.reasons["LOOP_ORIGIN_LIMIT"] = true
	return current, true
}
