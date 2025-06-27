//go:build amd64 || arm64 || loong64 || mips64 || mips64le || ppc64 || ppc64le || riscv64 || s390x || wasm

package exprx

import "github.com/expr-lang/expr/ast"

type cmpPatcher struct{}

func (cmpPatcher) Visit(node *ast.Node) {}
