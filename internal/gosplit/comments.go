package gosplit

import (
	"go/ast"
	"go/token"
	"sort"
)

func commentsForPart(fset *token.FileSet, file, part *ast.File) []*ast.CommentGroup {
	comments := ast.NewCommentMap(fset, file, file.Comments).Filter(part).Comments()
	seen := make(map[*ast.CommentGroup]bool, len(comments))
	appendComment := func(group *ast.CommentGroup) {
		if group != nil && !seen[group] {
			comments = append(comments, group)
			seen[group] = true
		}
	}
	for _, group := range comments {
		seen[group] = true
	}
	for _, declaration := range part.Decls {
		partGroup, ok := declaration.(*ast.GenDecl)
		if !ok || partGroup.Tok != token.IMPORT {
			continue
		}
		appendComment(partGroup.Doc)
		for _, raw := range partGroup.Specs {
			spec, ok := raw.(*ast.ImportSpec)
			if !ok {
				continue
			}
			appendComment(spec.Doc)
			appendComment(spec.Comment)
		}
		for _, declaration := range file.Decls {
			sourceGroup, ok := declaration.(*ast.GenDecl)
			if !ok || sourceGroup.Tok != token.IMPORT || sourceGroup.Pos() != partGroup.Pos() {
				continue
			}
			if sourceGroup.Doc != nil {
				appendComment(sourceGroup.Doc)
			}
		}
	}
	for _, group := range file.Comments {
		if group.End() < file.Package && !seen[group] {
			comments = append(comments, group)
		}
	}
	sort.Slice(comments, func(i, j int) bool { return comments[i].Pos() < comments[j].Pos() })
	return comments
}
