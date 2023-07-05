// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sema

import (
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// Return type.
type RetType struct {
	Kind   *TypeSymbol
	Idents []lex.Token
}

// Parameter.
type Param struct {
	Token    lex.Token
	Mutable  bool
	Variadic bool
	Kind     *TypeSymbol
	Ident    string
}

func (p *Param) instance() *ParamIns {
	return &ParamIns{
		Decl: p,
		Kind: nil,
	}
}

// Reports whether parameter is self (receiver) parameter.
func (p *Param) Is_self() bool { return p.Ident == "&self" || p.Ident == "self" }

// Reports whether self (receiver) parameter is reference.
func (p *Param) Is_ref() bool { return p.Ident != "" && p.Ident[0] == '&' }

// Function.
type Fn struct {
	sema *_Sema

	Token      lex.Token
	Global     bool
	Unsafety   bool
	Public     bool
	Cpp_linked bool
	Ident      string
	Directives []*ast.Directive
	Doc        string
	Scope      *ast.ScopeTree
	Generics   []*ast.GenericDecl
	Result     *RetType
	Params     []*Param
	Owner      *Struct

	// Function instances for each unique type combination of function call.
	// Nil if function is never used.
	Instances []*FnIns
}

// Reports whether return type is void.
func (f *Fn) Is_void() bool { return f.Result == nil }

// Reports whether function is method.
func (f *Fn) Is_method() bool { return f.Owner != nil }

// Reports whether function is entry point.
func (f *Fn) Is_entry_point() bool { return f.Ident == build.ENTRY_POINT }

// Reports whether function is anonymous function.
func (f *Fn) Is_anon() bool { return lex.Is_anon_ident(f.Ident) }

// Reports whether function has return variable(s).
func (f *Fn) Any_var() bool { return f.Result != nil && len(f.Result.Idents) > 0 }

// Reports whether any parameter uses generic types.
func (f *Fn) Parameters_uses_generics() bool {
	if len(f.Generics) == 0 {
		return false
	}

	for _, p := range f.Params {
		if p.Is_self() {
			continue
		}

		pk := p.Kind.Kind.To_str()
		for _, g := range f.Generics {
			if strings.Contains(pk, g.Ident) {
				return true
			}
		}
	}

	return false
}

// Reports whether result type uses generic types.
func (f *Fn) Result_uses_generics() bool {
	if f.Is_void() {
		return false
	} else if f.Result.Kind == nil || f.Result.Kind.Kind == nil {
		return false
	}

	rk := f.Result.Kind.Kind.To_str()
	for _, g := range f.Generics {
		if strings.Contains(rk, g.Ident) {
			return true
		}
	}

	return false
}

// Force to new instance.
func (f *Fn) instance_force() *FnIns {
	ins := &FnIns{
		Decl:  f,
		Scope: &Scope{},
	}

	ins.Params = make([]*ParamIns, len(f.Params))
	for i, p := range f.Params {
		ins.Params[i] = p.instance()
	}

	if ins.Decl.Result != nil {
		ins.Result = ins.Decl.Result.Kind.Kind
	}

	return ins
}

func (f *Fn) instance() *FnIns {
	// Returns already created instance for just one unique combination.
	if len(f.Generics) == 0 && len(f.Instances) == 1 {
		return f.Instances[0]
	}

	return f.instance_force()
}

func (f *Fn) append_instance(ins *FnIns) (bool, int) {
	if len(f.Generics) == 0 {
		// Skip already created instance for just one unique combination.
		if len(f.Instances) == 1 {
			return false, 0
		}

		f.Instances = append(f.Instances, ins)
		return true, -1
	}

	if len(f.Instances) == 0 {
		f.Instances = append(f.Instances, ins)
		return true, -1
	}

	for j, ains := range f.Instances {
		for i, ag := range ains.Generics {
			if ag.To_str() == ins.Generics[i].To_str() {
				// Instance is exist.
				return false, j
			}
		}
	}

	f.Instances = append(f.Instances, ins)
	return true, -1
}

// Parameter instance.
type ParamIns struct {
	Decl *Param
	Kind *TypeKind
}

// Implement: Kind
// Returns ParamIns's type kind as string.
func (p ParamIns) To_str() string {
	s := ""
	if p.Decl.Mutable {
		s += lex.KND_MUT + " "
	}

	if p.Decl.Is_self() {
		if p.Decl.Is_ref() {
			s += "&"
		}
		s += "self"
		return s
	}

	if p.Decl.Variadic {
		s += lex.KND_TRIPLE_DOT
	}
	if p.Kind != nil {
		s += p.Kind.To_str()
	}
	return s
}

// Function instance.
type FnIns struct {
	Owner    *StructIns
	Decl     *Fn
	Generics []*TypeKind
	Params   []*ParamIns
	Result   *TypeKind
	Scope    *Scope
	caller   _BuiltinCaller
	anon     bool
}

// Reports whether instance is built-in.
func (f *FnIns) Is_builtin() bool { return f.caller != nil }

// Reports whether instance is anonymous function.
func (f *FnIns) Is_anon() bool { return f.anon || f.Decl != nil && f.Decl.Is_anon() }

// Implement: Kind
// Returns Fn's type kind as string.
func (f FnIns) To_str() string {
	s := ""
	if f.Decl.Unsafety {
		s += "unsafe "
	}
	s += "fn"

	if len(f.Generics) > 0 {
		s += "["
		for i, t := range f.Generics {
			s += t.To_str()
			if i+1 < len(f.Generics) {
				s += ","
			}
		}
		s += "]"
	} else if len(f.Decl.Generics) > 0 { // Use Decl's generic if not parsed yet.
		s += "["
		for i, g := range f.Decl.Generics {
			s += g.Ident
			if i+1 < len(f.Decl.Generics) {
				s += ","
			}
		}
		s += "]"
	}

	s += "("
	n := len(f.Params)
	if n > 0 {
		for _, p := range f.Params {
			s += p.To_str()
			s += ","
		}
		s = s[:len(s)-1] // Remove comma.
	}
	s += ")"
	if !f.Decl.Is_void() {
		s += ":"
		s += f.Result.To_str()
	}
	return s
}
