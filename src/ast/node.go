// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package ast

import (
	"strings"

	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// Type of AST Node's data.
type NodeData = any

// AST Node.
type Node struct {
	Token lex.Token
	Data  any
}

// Reports whether node data is declaration.
func (n *Node) Is_decl() bool {
	switch n.Data.(type) {
	case *EnumDecl,
		*FnDecl,
		*StructDecl,
		*TraitDecl,
		*TypeAliasDecl,
		*FieldDecl,
		*UseDecl,
		*VarDecl,
		*TypeDecl:
		return true

	default:
		return false
	}
}

// Reports whether node data is comment or comment group.
func (n *Node) Is_comment() bool {
	switch n.Data.(type) {
	case *Comment, *CommentGroup:
		return true

	default:
		return false
	}
}

// Reports whether node data is impl.
func (n *Node) Is_impl() bool {
	switch n.Data.(type) {
	case *Impl:
		return true

	default:
		return false
	}
}

// Reports whether node data is use declaration.
func (n *Node) Is_use_decl() bool {
	switch n.Data.(type) {
	case *UseDecl:
		return true

	default:
		return false
	}
}

// Comment group.
type CommentGroup struct {
	Comments []*Comment
}

// Comment line.
type Comment struct {
	Token lex.Token
	Text  string
}

// Reports whether comment is directive.
func (c *Comment) Is_directive() bool {
	return strings.HasPrefix(c.Text, build.DIRECTIVE_PREFIX)
}

// Directive.
type Directive struct {
	Token lex.Token
	Tag   string
	Args  []string
}

// Kind type of type declarations.
type TypeDeclKind = any

// Type declaration.
// Also represents type expression.
//
// For primitive types:
//   - Represented by IdentType.
//   - Token's identity is data type.
//   - Primitive type kind is Ident.
//
// For function types:
//   - Function types represented by *FnDecl.
type TypeDecl struct {
	Token lex.Token
	Kind  TypeDeclKind
}

// Identifier type.
type IdentTypeDecl struct {
	Token      lex.Token
	Ident      string
	Cpp_linked bool
	Generics   []*TypeDecl
}

// Reports whether identifier is primitive type.
func (it *IdentTypeDecl) Is_prim() bool { return it.Token.Id == lex.ID_PRIM }

// Namespace chain type.
type NamespaceTypeDecl struct {
	Idents []lex.Token // Namespace chain with identifier tokens.
	Kind   *IdentTypeDecl  // Type of identifier.
}

type RefTypeDecl struct{ Elem *TypeDecl }      // Reference type.
type PtrTypeDecl struct{ Elem *TypeDecl }      // Pointer type.
type SlcTypeDecl struct{ Elem *TypeDecl }      // Slice type.
type TupleTypeDecl struct{ Types []*TypeDecl } // Tuple type.

// Reports whether pointer is unsafe pointer (*unsafe).
func (pt *PtrTypeDecl) Is_unsafe() bool { return pt.Elem == nil }

// Array type.
// Size epxression is nil for auto-sized array.
type ArrTypeDecl struct {
	Elem *TypeDecl
	Size *Expr
}

// Reports whether array is auto-sized.
func (at *ArrTypeDecl) Auto_sized() bool { return at.Size == nil }

// Map type.
type MapTypeDecl struct {
	Key *TypeDecl
	Val *TypeDecl
}

// Return type.
// Kind and Idents is nil for void type.
type RetTypeDecl struct {
	Kind   *TypeDecl
	Idents []lex.Token
}

// Type of Expr's data.
type ExprData = any

// Expression.
type Expr struct {
	Token lex.Token

	// Possible types:
	//  - *TupleExpr
	//  - *LitExpr
	//  - *Type
	//  - *IdentExpr
	//  - *UnaryExpr
	//  - *SubIdentExpr
	//  - *NsSelectionExpr
	//  - *VariadicExpr
	//  - *CastExpr
	//  - *FnCallExpr
	//  - *StructLit
	//  - *BraceLit
	//  - *SlicingExpr
	//  - *SliceExpr
	//  - *BinopExpr
	//  - *UnsafeExpr
	//  - *IndexingExpr
	//  - *FnDecl
	Kind ExprData
}

// Reports whether expression kind is function call.
func (e *Expr) Is_fn_call() bool {
	if e.Kind == nil {
		return false
	}
	switch e.Kind.(type) {
	case *FnCallExpr:
		return true
	default:
		return false
	}
}

// Tuple expression.
type TupleExpr struct {
	Expr []ExprData
}

// Literal expression.
type LitExpr struct {
	Token lex.Token
	Value string
}

// Unsafe expression.
type UnsafeExpr struct {
	Token lex.Token // Token of unsafe keyword.
	Expr  ExprData
}

// Reports whether literal is nil value.
func (le *LitExpr) Is_nil() bool { return le.Value == lex.KND_NIL }

// Identifier expression.
type IdentExpr struct {
	Token      lex.Token // Token of identifier.
	Ident      string
	Cpp_linked bool
}

// Reports whether identifier is self keyword.
func (ie *IdentExpr) Is_self() bool { return ie.Ident == lex.KND_SELF }

// Unary expression.
type UnaryExpr struct {
	Op   lex.Token
	Expr ExprData
}

// Variadiced expression.
type VariadicExpr struct {
	Token lex.Token
	Expr  ExprData
}

// Casting expression.
type CastExpr struct {
	Kind *TypeDecl
	Expr ExprData
}

// Namespace identifier selection expression.
type NsSelectionExpr struct {
	Ns    []lex.Token // Tokens of selected namespace identifier chain.
	Ident lex.Token   // Token of selected identifier.
}

// Object sub identifier selection expression.
type SubIdentExpr struct {
	Is_self bool      // True if root selector is "self" keyword.
	Expr    ExprData  // Selected object.
	Ident   lex.Token // TOken of selected identifier.
}

// Binary operation.
type BinopExpr struct {
	Left  ExprData
	Right ExprData
	Op    lex.Token
}

// Function call expression kind.
type FnCallExpr struct {
	Token      lex.Token
	Expr       *Expr
	Generics   []*TypeDecl
	Args       []*Expr
	Concurrent bool
}

// Field-Expression pair.
type FieldExprPair struct {
	Field lex.Token // Field identifier token.
	Expr  ExprData
}

// Reports whether pair targeted field.
func (fep *FieldExprPair) Is_targeted() bool { return fep.Field.Id != lex.ID_NA }

// Struct literal instiating expression.
type StructLit struct {
	Kind  *TypeDecl
	Exprs []ExprData // Possible types: *FieldExprPair, and other expressions.
}

// Anonymous brace instiating expression.
// Empty braces ( {} ).
type BraceLit struct {
	Token lex.Token
	Exprs []ExprData
}

// Reports literal is empty ( {} ).
func (bl *BraceLit) Is_empty() bool { return len(bl.Exprs) == 0 }

// Key-value pair expression.
type KeyValPair struct {
	Key   ExprData
	Val   ExprData
	Colon lex.Token
}

// Slice initiating expression.
// Also represents array initiating expression.
type SliceExpr struct {
	Token lex.Token
	Elems []ExprData
}

// Reports whether slice is empty.
func (se *SliceExpr) Is_empty() bool { return len(se.Elems) == 0 }

// Indexing expression.
type IndexingExpr struct {
	Token lex.Token
	Expr  ExprData // Value expression to indexing.
	Index ExprData // Index value expression.
}

// Slicing expression.
type SlicingExpr struct {
	Token lex.Token
	Expr  ExprData // Value expression to slicing.
	Start ExprData // Start index value expression.
	To    ExprData // To index value expression.
}

// GenericDecl type.
type GenericDecl struct {
	Token lex.Token
	Ident string
}

// Label statement.
type LabelSt struct {
	Token lex.Token
	Ident string
}

// Goto statement.
type GotoSt struct {
	Token lex.Token
	Label lex.Token
}

// Fall statement.
type FallSt struct {
	Token lex.Token
}

// Left expression of assign statement.
type AssignLeft struct {
	Token   lex.Token
	Mutable bool
	Ident   string
	Expr    *Expr
}

// Assign statement.
type AssignSt struct {
	Declarative bool
	Setter      lex.Token
	L           []*AssignLeft
	R           *Expr
}

// ScopeTree.
type ScopeTree struct {
	Parent   *ScopeTree // Nil if scope is root.
	Unsafety bool
	Deferred bool
	Stmts    []NodeData // Statements.
}

// ParamDecl.
type ParamDecl struct {
	Token    lex.Token
	Mutable  bool
	Variadic bool
	Kind     *TypeDecl
	Ident    string
}

// Reports whether parameter is self (receiver) parameter.
func (p *ParamDecl) Is_self() bool { return strings.HasSuffix(p.Ident, lex.KND_SELF) }

// Reports whether self (receiver) parameter is reference.
func (p *ParamDecl) Is_ref() bool { return p.Ident != "" && p.Ident[0] == '&' }

// Function declaration.
// Also represents anonymous function expression.
type FnDecl struct {
	Token        lex.Token
	Global       bool
	Unsafety     bool
	Public       bool
	Cpp_linked   bool
	Ident        string
	Directives   []*Directive
	Doc_comments *CommentGroup
	Scope        *ScopeTree
	Generics     []*GenericDecl
	Result       *RetTypeDecl
	Params       []*ParamDecl
}

// Variable declaration.
type VarDecl struct {
	Scope        *ScopeTree // nil for global scopes
	Token        lex.Token
	Ident        string
	Cpp_linked   bool
	Public       bool
	Mutable      bool
	Constant     bool
	Doc_comments *CommentGroup
	Kind         *TypeDecl // nil for auto-typed
	Expr         *Expr
}

// Return statement.
type RetSt struct {
	Token lex.Token
	Expr  *Expr
}

type IterKind = any // Type of Iter's kind.

// Iteration.
type Iter struct {
	Token lex.Token
	Kind  IterKind
	Scope *ScopeTree
}

// Reports whether iteration is infinity.
func (i *Iter) Is_inf() bool { return i.Kind == nil }

// While iteration kind.
type WhileKind struct {
	Expr       *Expr
	Next       NodeData // Nil if kind is while-next iteration.
	Next_token lex.Token
}

// Reports whether kind is while-next iteration.
func (wk *WhileKind) Is_while_next() bool { return wk.Next != nil }

// Range iteration kind.
type RangeKind struct {
	In_token lex.Token // Token of "in" keyword
	Expr     *Expr
	Key_a    *VarDecl // first key of range
	Key_b    *VarDecl // second key of range
}

// Break statement.
type BreakSt struct {
	Token lex.Token
	Label lex.Token
}

// Continue statement.
type ContSt struct {
	Token lex.Token
	Label lex.Token
}

// If condition.
type If struct {
	Token lex.Token
	Expr  *Expr
	Scope *ScopeTree
}

// Else condition.
type Else struct {
	Token lex.Token
	Scope *ScopeTree
}

// Condition chain.
type Conditional struct {
	Head    *If
	Tail    []*If
	Default *Else
}

// Type alias declration.
type TypeAliasDecl struct {
	Scope        *ScopeTree
	Public       bool
	Cpp_linked   bool
	Token        lex.Token
	Ident        string
	Kind         *TypeDecl
	Doc_comments *CommentGroup
}

// Case of match-case.
type Case struct {
	Token lex.Token
	Scope *ScopeTree

	// Holds expression.
	// Expressions holds *Type if If type matching.
	Exprs []*Expr
}

// Match-Case.
type MatchCase struct {
	Token      lex.Token
	Type_match bool
	Expr       *Expr
	Cases      []*Case
	Default    *Else
}

// Use declaration statement.
type UseDecl struct {
	Token      lex.Token
	Link_path  string // Use declaration path string.
	Full       bool   // Full implicit import.
	Selected   []lex.Token
	Cpp_linked bool // Cpp header use declaration.
	Std        bool // Standard package use declaration.
}

// Enum item.
type EnumItemDecl struct {
	Token lex.Token
	Ident string
	Expr  *Expr // Nil for auto expression.
}

// Reports whether item has auto expression.
func (ei *EnumItemDecl) Auto_expr() bool { return ei.Expr == nil }

// Enum declaration.
type EnumDecl struct {
	Token        lex.Token
	Public       bool
	Ident        string
	Kind         *TypeDecl
	Items        []*EnumItemDecl
	Doc_comments *CommentGroup
}

// Reports enum's type is default.
func (ed *EnumDecl) Default_typed() bool { return ed.Kind == nil }

// Field declaration.
type FieldDecl struct {
	Token   lex.Token
	Public  bool
	Mutable bool // Interior mutability.
	Ident   string
	Kind    *TypeDecl
}

// Structure declaration.
type StructDecl struct {
	Token        lex.Token
	Ident        string
	Fields       []*FieldDecl
	Public       bool
	Cpp_linked   bool
	Directives   []*Directive
	Doc_comments *CommentGroup
	Generics     []*GenericDecl
}

// Trait declaration.
type TraitDecl struct {
	Token        lex.Token
	Ident        string
	Public       bool
	Doc_comments *CommentGroup
	Methods      []*FnDecl
}

// Implementation.
type Impl struct {
	// This token available for these cases:
	//  - Implementation trait to structure, represents trait's token.
	Base lex.Token

	// This token available for these cases:
	//  - Implementation trait to structure, represents structure's token.
	//  - Implementation to structure, represents structure's token.
	Dest lex.Token

	// Given methods to implement.
	Methods []*FnDecl
}

// Reports whether implementation type is trait to structure.
func (ipl *Impl) Is_trait_impl() bool { return ipl.Base.Id != lex.ID_NA }

// Reports whether implementation type is append to destination structure.
func (ipl *Impl) Is_struct_impl() bool { return ipl.Base.Id == lex.ID_NA }
