//go:build !(amd64 || arm64 || loong64 || mips64 || mips64le || ppc64 || ppc64le || riscv64 || s390x || wasm)

package exprx

import (
	"math"
	"reflect"

	"github.com/expr-lang/expr/ast"
)

type cmpPatcher struct{}

func (cmpPatcher) Visit(node *ast.Node) {
	n, ok := (*node).(*ast.BinaryNode)
	if !ok {
		return
	}

	if !isIntKind(n.Left.Type().Kind()) || !isIntKind(n.Right.Type().Kind()) {
		return
	}

	if !(n.Left.Type().OverflowInt(math.MaxInt32) || !n.Right.Type().OverflowInt(math.MaxInt32)) {
		return
	}

	var binaryNode *ast.BinaryNode
	var cmpValue int
	switch op := n.Operator; op {
	case ">":
		cmpValue = 1
	case "<":
		cmpValue = -1
	case "==":
		cmpValue = 0
	case ">=", "<=":
		binaryNode = &ast.BinaryNode{
			Operator: op,
			Right:    &ast.IntegerNode{Value: 0},
		}
	default:
		return
	}

	if binaryNode == nil {
		binaryNode = &ast.BinaryNode{
			Operator: "==",
			Right:    &ast.IntegerNode{Value: cmpValue},
		}
	}

	binaryNode.Left = &ast.CallNode{
		Callee:    &ast.IdentifierNode{Value: "cmp"},
		Arguments: []ast.Node{n.Left, n.Right},
	}
	ast.Patch(node, binaryNode)

	(*node).SetType(reflect.TypeFor[int64]())
}

func isIntKind(rkind reflect.Kind) bool {
	switch rkind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	}
	return false
}
