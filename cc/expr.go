// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc

import "fmt"

// An Expr is a parsed C expression.
type Expr struct {
	SyntaxInfo
	Op    ExprOp   // operator
	Left  *Expr    // left (or only) operand
	Right *Expr    // right operand
	List  []*Expr  // operand list, for Comma, Cond, Call
	Text  string   // name or literal, for Name, Number, Goto, Arrow, Dot
	Texts []string // list of literals, for String
	Type  *Type    // type operand, for SizeofType, Offsetof, Cast, CastInit, VaArg
	Init  *Init    // initializer, for CastInit
	Block []*Stmt  // for c2go

	// derived information
	XDecl *Decl
	XType *Type // expression type, derived
}

func (x *Expr) String() string {
	var p Printer
	p.hideComments = true
	p.printExpr(x, precLow)
	return p.String()
}

type ExprOp int

const (
	_          ExprOp = iota
	Add               // Left + Right
	AddEq             // Left += Right
	Addr              // &Left
	And               // Left & Right
	AndAnd            // Left && Right
	AndEq             // Left &= Right
	Arrow             // Left->Text
	Call              // Left(List)
	Cast              // (Type)Left
	CastInit          // (Type){Init}
	Comma             // x, y, z; List = {x, y, z}
	Cond              // x ? y : z; List = {x, y, z}
	Div               // Left / Right
	DivEq             // Left /= Right
	Dot               // Left.Name
	Eq                // Left = Right
	EqEq              // Left == Right
	Gt                // Left > Right
	GtEq              // Left >= Right
	Index             // Left[Right]
	Indir             // *Left
	Lsh               // Left << Right
	LshEq             // Left <<= Right
	Lt                // Left < Right
	LtEq              // Left <= Right
	Minus             // -Left
	Mod               // Left % Right
	ModEq             // Left %= Right
	Mul               // Left * Right
	MulEq             // Left *= Right
	Name              // Text (function, variable, or enum name)
	Not               // !Left
	NotEq             // Left != Right
	Number            // Text (numeric or chraracter constant)
	Offsetof          // offsetof(Type, Left)
	Or                // Left | Right
	OrEq              // Left |= Right
	OrOr              // Left || Right
	Paren             // (Left)
	Plus              //  +Left
	PostDec           // Left--
	PostInc           // Left++
	PreDec            // --Left
	PreInc            // ++Left
	Rsh               // Left >> Right
	RshEq             // Left >>= Right
	SizeofExpr        // sizeof(Left)
	SizeofType        // sizeof(Type)
	String            // Text (quoted string literal)
	Sub               // Left - Right
	SubEq             // Left -= Right
	Twid              // ~Left
	VaArg             // va_arg(Left, Type)
	Xor               // Left ^ Right
	XorEq             // Left ^= Right
)

var exprOpString = []string{
	Add:        "Add",
	AddEq:      "AddEq",
	Addr:       "Addr",
	And:        "And",
	AndAnd:     "AndAnd",
	AndEq:      "AndEq",
	Arrow:      "Arrow",
	Call:       "Call",
	Cast:       "Cast",
	CastInit:   "CastInit",
	Comma:      "Comma",
	Cond:       "Cond",
	Div:        "Div",
	DivEq:      "DivEq",
	Dot:        "Dot",
	Eq:         "Eq",
	EqEq:       "EqEq",
	Gt:         "Gt",
	GtEq:       "GtEq",
	Index:      "Index",
	Indir:      "Indir",
	Lsh:        "Lsh",
	LshEq:      "LshEq",
	Lt:         "Lt",
	LtEq:       "LtEq",
	Minus:      "Minus",
	Mod:        "Mod",
	ModEq:      "ModEq",
	Mul:        "Mul",
	MulEq:      "MulEq",
	Name:       "Name",
	Not:        "Not",
	NotEq:      "NotEq",
	Number:     "Number",
	Offsetof:   "Offsetof",
	Or:         "Or",
	OrEq:       "OrEq",
	OrOr:       "OrOr",
	Paren:      "Paren",
	Plus:       "Plus",
	PostDec:    "PostDec",
	PostInc:    "PostInc",
	PreDec:     "PreDec",
	PreInc:     "PreInc",
	Rsh:        "Rsh",
	RshEq:      "RshEq",
	SizeofExpr: "SizeofExpr",
	SizeofType: "SizeofType",
	String:     "String",
	Sub:        "Sub",
	SubEq:      "SubEq",
	Twid:       "Twid",
	VaArg:      "VaArg",
	Xor:        "Xor",
	XorEq:      "XorEq",
}

func (op ExprOp) String() string {
	if 0 <= int(op) && int(op) <= len(exprOpString) {
		return exprOpString[op]
	}
	return fmt.Sprintf("ExprOp(%d)", op)
}

// Prefix is an initializer prefix.
type Prefix struct {
	Span  Span
	Dot   string // .Dot =
	XDecl *Decl  // for .Dot
	Index *Expr  // [Index] =
}

// Init is an initializer expression.
type Init struct {
	SyntaxInfo
	Prefix []*Prefix // list of prefixes
	Expr   *Expr     // Expr
	Braced []*Init   // {Braced}

	XType *Type // derived type
}

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
