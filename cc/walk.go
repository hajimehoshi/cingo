// Copyright 2018 Hajime Hoshi.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc

import (
	"fmt"
)

// Walk traverses the syntax x, calling before and after on entry to and exit from
// each Syntax encountered during the traversal. In case of cross-linked input,
// the traversal never visits a given Syntax more than once.
func Walk(x Syntax, before, after func(Syntax)) {
	seen := map[Syntax]struct{}{
		nil:           struct{}{},
		(*Decl)(nil):  struct{}{},
		(*Init)(nil):  struct{}{},
		(*Type)(nil):  struct{}{},
		(*Expr)(nil):  struct{}{},
		(*Stmt)(nil):  struct{}{},
		(*Label)(nil): struct{}{},
	}
	walk(x, func(s Syntax, indent int) { before(s) }, func(s Syntax, indent int) { after(s) }, seen, 0)
}

func WalkIndent(x Syntax, before, after func(Syntax, int)) {
	seen := map[Syntax]struct{}{
		nil:           struct{}{},
		(*Decl)(nil):  struct{}{},
		(*Init)(nil):  struct{}{},
		(*Type)(nil):  struct{}{},
		(*Expr)(nil):  struct{}{},
		(*Stmt)(nil):  struct{}{},
		(*Label)(nil): struct{}{},
	}
	walk(x, before, after, seen, 0)
}

func walk(x Syntax, before, after func(Syntax, int), seen map[Syntax]struct{}, indent int) {
	if _, ok := seen[x]; ok {
		return
	}
	seen[x] = struct{}{}
	before(x, indent)
	switch x := x.(type) {
	default:
		panic(fmt.Sprintf("walk: unexpected type %T", x))

	case *Prog:
		for _, d := range x.Decls {
			walk(d, before, after, seen, indent+1)
		}

	case *Decl:
		walk(x.Type, before, after, seen, indent+1)
		walk(x.Init, before, after, seen, indent+1)
		walk(x.Body, before, after, seen, indent+1)

	case *Init:
		for _, b := range x.Braced {
			walk(b, before, after, seen, indent+1)
		}
		walk(x.Expr, before, after, seen, indent+1)

	case *Type:
		walk(x.Base, before, after, seen, indent+1)
		for _, d := range x.Decls {
			walk(d, before, after, seen, indent+1)
		}
		walk(x.Width, before, after, seen, indent+1)

	case *Expr:
		walk(x.Left, before, after, seen, indent+1)
		walk(x.Right, before, after, seen, indent+1)
		for _, y := range x.List {
			walk(y, before, after, seen, indent+1)
		}
		walk(x.Type, before, after, seen, indent+1)
		walk(x.Init, before, after, seen, indent+1)
		for _, y := range x.Block {
			walk(y, before, after, seen, indent+1)
		}

	case *Stmt:
		walk(x.Pre, before, after, seen, indent+1)
		walk(x.Expr, before, after, seen, indent+1)
		walk(x.Post, before, after, seen, indent+1)
		walk(x.Decl, before, after, seen, indent+1)
		walk(x.Body, before, after, seen, indent+1)
		walk(x.Else, before, after, seen, indent+1)
		for _, y := range x.Block {
			walk(y, before, after, seen, indent+1)
		}
		for _, y := range x.Labels {
			walk(y, before, after, seen, indent+1)
		}

	case *Label:
		walk(x.Expr, before, after, seen, indent+1)
	}
	after(x, indent)
}

// Preorder calls f for each piece of syntax of x in a preorder traversal.
func Preorder(x Syntax, f func(Syntax)) {
	Walk(x, f, func(Syntax) {})
}

// Preorder calls f for each piece of syntax of x in a postorder traversal.
func Postorder(x Syntax, f func(Syntax)) {
	Walk(x, func(Syntax) {}, f)
}
