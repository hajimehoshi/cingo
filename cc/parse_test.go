// Copyright 2018 Hajime Hoshi.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc_test

import (
	"fmt"
	"strings"

	. "github.com/hajimehoshi/c2go/cc"
)

func prettyPrint(in string) {
	prog, err := Read("", strings.NewReader(in))
	if err != nil {
		panic(err)
	}
	WalkIndent(prog, func(x Syntax, indent int) {
		prettyPrintSyntax(x, indent)
	}, func(x Syntax, indent int) {})
}

func prettyPrintSyntax(x Syntax, indent int) {
	fmt.Printf("%s", strings.Repeat("  ", indent))

	switch x := x.(type) {
	default:
		panic(fmt.Sprintf("unexpected type %T", x))

	case *Prog:
		fmt.Printf("prog")

	case *Decl:
		fmt.Printf("decl")
		if x.Name != "" {
			fmt.Printf(": %s", x.Name)
		}

	case *Type:
		fmt.Printf("type: %s", x)

	case *Stmt:
		fmt.Printf("stmt: %s", x.Op)

	case *Expr:
		fmt.Printf("expr: %s", x.Op)
		if x.Op == Name || x.Op == Number || x.Op == Arrow || x.Op == Dot {
			fmt.Printf(", %s", x.Text)
		}

	case *Label:
		fmt.Printf("label")
	}

	fmt.Println()
}

func ExampleHelloWorld() {
	prettyPrint(`void printf(char*);

int main() {
  printf("Hello, World!");
  return 0;
}`)
	// Output:
	// prog
	//   decl: printf
	//     type: func( char*) void
	//       type: void
	//       decl
	//         type: char*
	//           type: char
	//   decl: main
	//     type: func() int
	//       type: int
	//     stmt: Block
	//       stmt: StmtExpr
	//         expr: Call
	//           expr: Name, printf
	//           expr: String
	//       stmt: Return
	//         expr: Number, 0
}
