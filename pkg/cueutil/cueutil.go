package cueutil

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
)

// ASTToValue converst cue AST to a cue.Value using the given runtime.
func ASTToValue(r *cue.Runtime, ast ast.Node) (cue.Value, error) {
	// Format the cue code.
	str, err := ASTToSyntax(ast)
	if err != nil {
		return cue.Value{}, err
	}
	inst, err := r.Compile(".", str)
	if err != nil {
		fmt.Println(str)
		return cue.Value{}, fmt.Errorf("error converting node to value: %w", err)
	}

	return inst.Value(), nil
}

func ASTToSyntax(ast ast.Node) (string, error) {
	byt, err := format.Node(
		ast,
		format.TabIndent(false),
		format.UseSpaces(2),
	)
	if err != nil {
		return "", fmt.Errorf("error converting node to syntax: %w", err)
	}
	return string(byt), nil
}
