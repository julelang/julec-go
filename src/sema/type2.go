package sema

import (
	"math"
	"strconv"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
)

// This file reserved for type compatibility checking.

func get_fn_result_types(f *FnIns) []*TypeKind {
	switch {
	case f.Decl.Is_void() || f.Result == nil:
		return nil

	case f.Result.Tup() == nil:
		return []*TypeKind{f.Result}

	default:
		return f.Result.Tup().Types
	}
}

func trait_has_reference_receiver(t *Trait) bool {
	for _, f := range t.Methods {
		p := f.Params[0]
		if p.Is_ref() && p.Is_self() {
			return true
		}
	}
	return false
}

func float_assignable(kind string, d *Data) bool {
	value := strconv.FormatFloat(d.Constant.Read_f64(), 'e', -1, 64)
	return types.Check_bit_float(value, types.Bitsize_of(kind))
}

func sig_assignable(kind string, d *Data) bool {
	min := types.Min_of(kind)
	max := types.Max_of(kind)

	switch {
	case d.Constant.Is_f64():
		x := float64(d.Constant.Read_f64())
		i, frac := math.Modf(x)
		if frac != 0 {
			return false
		}
		return i >= min && i <= max

	case d.Constant.Is_u64():
		x := float64(d.Constant.Read_u64())
		if x <= max {
			return true
		}

	case d.Constant.Is_i64():
		x := float64(d.Constant.Read_i64())
		return min <= x && x <= max
	}

	return false
}

func unsig_assignable(kind string, d *Data) bool {
	max := types.Max_of(kind)

	switch {
	case d.Constant.Is_f64():
		x := d.Constant.Read_f64()
		if x < 0 {
			return false
		}

		i, frac := math.Modf(x)
		if frac != 0 {
			return false
		}
		return i <= max

	case d.Constant.Is_u64():
		x := float64(d.Constant.Read_u64())
		if x <= max {
			return true
		}

	case d.Constant.Is_i64():
		x := float64(d.Constant.Read_i64())
		return 0 <= x && x <= max
	}

	return false
}

func int_assignable(kind string, d *Data) bool {
	switch {
	case types.Is_sig_int(kind):
		return sig_assignable(kind, d)

	case types.Is_unsig_int(kind):
		return unsig_assignable(kind, d)

	default:
		return false
	}
}

type _TypeCompatibilityChecker struct {
	s           *_Sema // Used for error logging.
	dest        *TypeKind
	src         *TypeKind
	error_token lex.Token

	// References uses elem's type itself if true.
	deref bool
}

func (tcc *_TypeCompatibilityChecker) push_err(key string, args ...any) {
	tcc.s.push_err(tcc.error_token, key, args...)
}

func (tcc *_TypeCompatibilityChecker) check_trait() (ok bool) {
	if tcc.src.Is_nil() {
		return true
	}

	trt := tcc.dest.Trt()
	ref := false
	switch {
	case tcc.src.Ref() != nil:
		ref = true
		tcc.src = tcc.src.Ref().Elem
		if tcc.src.Strct() == nil {
			return false
		}
		fallthrough

	case tcc.src.Strct() != nil:
		s := tcc.src.Strct()
		if !s.Decl.Is_implements(trt) {
			return false
		}

		if trait_has_reference_receiver(trt) && !ref {
			tcc.push_err("trait_has_reference_parametered_function")
			return false
		}

		return true

	case tcc.src.Trt() != nil:
		return trt == tcc.src.Trt()

	default:
		return false
	}
}

func (tcc *_TypeCompatibilityChecker) check_ref() (ok bool) {
	if tcc.dest.To_str() == tcc.src.To_str() {
		return true
	} else if !tcc.deref {
		return false
	} else if tcc.src.Ref() == nil {
		tcc.dest = tcc.dest.Ref().Elem
		return tcc.check()
	}

	tcc.src = tcc.src.Ref().Elem
	return tcc.check()
}

func (tcc *_TypeCompatibilityChecker) check_ptr() (ok bool) {
	if tcc.src.Is_nil() {
		return true
	}
	
	src := tcc.src.Ptr()
	if src == nil {
		return false
	}

	dest := tcc.dest.Ptr()

	if dest.Is_unsafe() {
		return true
	}

	sub := _TypeCompatibilityChecker{
		s:           tcc.s,
		error_token: tcc.error_token,
		src:         src.Elem,
		dest:        dest.Elem,
	}
	return sub.check()
}

func (tcc *_TypeCompatibilityChecker) check_slc() (ok bool) {
	if tcc.src.Is_nil() {
		return true
	}

	src := tcc.src.Slc()
	if src == nil {
		return false
	}
	dest := tcc.dest.Slc()

	sub := _TypeCompatibilityChecker{
		s:           tcc.s,
		error_token: tcc.error_token,
		src:         src.Elem,
		dest:        dest.Elem,
	}
	return sub.check()
}

func (tcc *_TypeCompatibilityChecker) check_arr() (ok bool) {
	src := tcc.src.Arr()
	if src == nil {
		return false
	}
	dest := tcc.dest.Arr()
	if dest.N != src.N {
		return false
	}
	
	sub := _TypeCompatibilityChecker{
		s:           tcc.s,
		error_token: tcc.error_token,
		src:         src.Elem,
		dest:        dest.Elem,
	}
	return sub.check()
}

func (tcc *_TypeCompatibilityChecker) check_map() (ok bool) {
	if tcc.src.Is_nil() {
		return true
	}
	
	src := tcc.src.Map()
	if src == nil {
		return false
	}

	dest := tcc.dest.Map()

	sub := _TypeCompatibilityChecker{
		s:           tcc.s,
		error_token: tcc.error_token,
		src:         src.Key,
		dest:        dest.Key,
	}
	if !sub.check() {
		return false
	}

	sub.src = src.Val
	sub.dest = dest.Val
	return sub.check()
}

func (tcc *_TypeCompatibilityChecker) check_struct() (ok bool) {
	src := tcc.src.Strct()
	if src == nil {
		return false
	}
	dest := tcc.dest.Strct()
	switch {
	case dest.Decl != src.Decl:
		return false

	case len(dest.Generics) == 0:
		return true
	}

	for i, dg := range dest.Generics {
		sg := src.Generics[i]
		sub := _TypeCompatibilityChecker{
			s:           tcc.s,
			error_token: tcc.error_token,
			src:         sg,
			dest:        dg,
		}
		if !sub.check() {
			return false
		}
	}
	return true
}

func (tcc *_TypeCompatibilityChecker) check_fn() (ok bool) {
	if tcc.src.Is_nil() {
		return true
	}

	src  := tcc.src.Fnc()
	if src == nil {
		return false
	}

	dest := tcc.dest.Fnc()
	if (src.Result != nil) != (dest.Result != nil) {
		return false
	}
	if len(src.Params) != len(dest.Params) {
		return false
	}

	for i := range src.Params {
		srcp := src.Params[i]
		destp := dest.Params[i]
		if (srcp != nil) && (destp != nil) {
			return false
		}

		sub := _TypeCompatibilityChecker{
			s:           tcc.s,
			error_token: tcc.error_token,
			src:         srcp.Kind,
			dest:        destp.Kind,
		}
		if !sub.check() {
			return false
		}
	}

	if src.Result != nil {
		sub := _TypeCompatibilityChecker{
			s:           tcc.s,
			error_token: tcc.error_token,
			src:         src.Result,
			dest:        dest.Result,
		}
		if !sub.check() {
			return false
		}
	}

	return true
}

func (tcc *_TypeCompatibilityChecker) check_enum() (ok bool) {
	r := tcc.src.Enm()
	if r == nil {
		return false
	}
	return tcc.dest.Enm() == r
}

func (tcc *_TypeCompatibilityChecker) check() (ok bool) {
	if tcc.src == nil {
		return false
	}

	if tcc.dest.Ref() != nil {
		return tcc.check_ref()
	}

	if tcc.src.Ref() != nil {
		tcc.src = tcc.src.Ref().Elem
	}

	switch {
	case tcc.dest.Trt() != nil:
		return tcc.check_trait()

	case tcc.dest.Ptr() != nil:
		return tcc.check_ptr()

	case tcc.dest.Slc() != nil:
		return tcc.check_slc()

	case tcc.dest.Arr() != nil:
		return tcc.check_arr()

	case tcc.dest.Map() != nil:
		return tcc.check_map()

	case tcc.dest.Enm() != nil:
		return tcc.check_enum()

	case tcc.dest.Strct() != nil:
		return tcc.check_struct()

	case tcc.dest.Fnc() != nil:
		return tcc.check_fn()

	case is_nil_compatible(tcc.dest):
		return tcc.src.Is_nil()

	default:
		return types.Types_are_compatible(tcc.dest.To_str(), tcc.src.To_str())
	}
}

// Checks value and type compatibility for assignment.
type _AssignTypeChecker struct {
	s           *_Sema // Used for error logging and type checking.
	dest        *TypeKind
	d           *Data
	error_token lex.Token
	deref       bool // Same as TypeCompatibilityChecker.deref field.
}

func (tcc *_AssignTypeChecker) push_err(key string, args ...any) {
	tcc.s.push_err(tcc.error_token, key, args...)
}

func (atc *_AssignTypeChecker) check_validity() bool {
	valid := true

	switch {
	case atc.d.Kind.Fnc() != nil:
		f := atc.d.Kind.Fnc()
		switch {
		case f.Is_builtin():
			atc.push_err("builtin_as_anonymous_fn")
			valid = false

		case f.Decl.Is_method():
			atc.push_err("method_as_anonymous_fn")
			valid = false

		case len(f.Decl.Generics) > 0:
			atc.push_err("genericed_fn_as_anonymous_fn")
			valid = false
		}

	case atc.d.Kind.Tup() != nil:
		atc.push_err("tuple_assign_to_single")
		valid = false
	}

	return valid
}

func (atc *_AssignTypeChecker) check_const() bool {
	if !atc.d.Is_const() || atc.dest.Prim() == nil ||
		atc.d.Kind.Prim() == nil || !types.Is_num(atc.d.Kind.Prim().kind) {
		return false
	}

	kind := atc.dest.Prim().kind
	switch {
	case types.Is_float(kind):
		if !float_assignable(kind, atc.d) {
			atc.push_err("overflow_limits")
			return false
		}

	case types.Is_int(kind):
		if !int_assignable(kind, atc.d) {
			atc.push_err("overflow_limits")
			return false
		}

	default:
		return false
	}

	return true
}

func (atc *_AssignTypeChecker) check() bool {
	switch {
	case atc.d == nil:
		// Skip Data is nil.
		return false

	case !atc.check_validity():
		// Data is invalid and error(s) logged about it.
		return false

	case atc.d.Variadiced:
		atc.push_err("incompatible_types", atc.dest.To_str(), atc.d.Kind.To_str()+"...")
		return false

	case atc.check_const():
		return true

	default:
		return atc.s.check_type_compatibility(atc.dest, atc.d.Kind, atc.error_token, atc.deref)
	}
}

type _DynamicTypeAnnotation struct {
	e           *_Eval
	f           *FnIns
	p           *ParamIns
	a           *Data
	error_token lex.Token
	k           **TypeKind
}

func (dta *_DynamicTypeAnnotation) push_generic(k *TypeKind, i int) {
	if k.Enm() != nil {
		dta.e.push_err(dta.error_token, "enum_not_supports_as_generic")
	}
	dta.f.Generics[i] = k
}

func (dta *_DynamicTypeAnnotation) annotate_prim(k *TypeKind) (ok bool) {
	kind := (*dta.k).To_str()
	for i, g := range dta.f.Decl.Generics {
		if kind != g.Ident {
			continue
		}

		t := dta.f.Generics[i]
		switch {
		case t == nil:
			dta.push_generic(k, i)

		case t.To_str() != k.To_str():
			// Generic already pushed but generic type and current kind
			// is different, so incompatible.
			return false
		}
		(*dta.k).kind = k
		return true
	}

	return false
}

func (dta *_DynamicTypeAnnotation) annotate_slc(k *TypeKind) (ok bool) {
	pslc := (*dta.k).Slc()
	if pslc == nil {
		return false
	}

	slc := k.Slc()
	dta.k = &pslc.Elem
	return dta.annotate_kind(slc.Elem)
}

func (dta *_DynamicTypeAnnotation) annotate_map(k *TypeKind) (ok bool) {
	pmap := (*dta.k).Map()
	if pmap == nil {
		return false
	}

	m := k.Map()
	check := func(k **TypeKind, ck *TypeKind) (ok bool) {
		old := dta.k
		dta.k = k
		ok = dta.annotate_kind(ck)
		dta.k = old
		return ok
	}
	return check(&pmap.Key, m.Key) && check(&pmap.Val, m.Val)
}

func (dta *_DynamicTypeAnnotation) annotate_fn(k *TypeKind) (ok bool) {
	pf := (*dta.k).Fnc()
	if pf == nil {
		return false
	}
	f := k.Fnc()
	switch {
	case len(pf.Params) != len(f.Params):
		return false

	case pf.Decl.Is_void() != f.Decl.Is_void():
		return false
	}

	ok = true
	old := dta.k
	for i, fp := range f.Params {
		pfp := pf.Params[i]
		dta.k = &pfp.Kind
		ok = dta.annotate_kind(fp.Kind) && ok
	}

	if !pf.Decl.Is_void() {
		dta.k = &pf.Result
		ok = dta.annotate_kind(f.Result) && ok
	}

	dta.k = old
	return ok
}

func (dta *_DynamicTypeAnnotation) annotate_ptr(k *TypeKind) (ok bool) {
	pptr := (*dta.k).Ptr()
	if pptr == nil {
		return false
	}

	ptr := k.Ptr()
	dta.k = &pptr.Elem
	return dta.annotate_kind(ptr.Elem)
}

func (dta *_DynamicTypeAnnotation) annotate_any(k *TypeKind) (ok bool) {
	kind := (*dta.k).To_str()
	for i, g := range dta.f.Decl.Generics {
		if kind != g.Ident {
			continue
		}

		t := dta.f.Generics[i]
		switch {
		case t == nil:
			dta.push_generic(k, i)

		case t.To_str() != k.To_str():
			// Generic already pushed but generic type and current kind
			// is different, so incompatible.
			return false
		}
		(*dta.k).kind = k
		return true
	}

	return false
}

func (dta *_DynamicTypeAnnotation) annotate_kind(k *TypeKind) (ok bool) {
	if dta.k == nil || *dta.k == nil {
		return
	}

	switch {
	case dta.annotate_any(k):
		return true

	case k.Prim() != nil:
		return dta.annotate_prim(k)

	case k.Slc() != nil:
		return dta.annotate_slc(k)

	case k.Map() != nil:
		return dta.annotate_map(k)

	case k.Fnc() != nil:
		return dta.annotate_fn(k)

	case k.Ptr() != nil:
		return dta.annotate_ptr(k)

	default:
		return false
	}
}

func (dta *_DynamicTypeAnnotation) annotate() (ok bool) {
	pk := dta.p.Kind
	dta.k = &pk

	return dta.annotate_kind(dta.a.Kind)
}

type _FnCallArgChecker struct {
	e                  *_Eval
	args               []*ast.Expr
	error_token        lex.Token
	f                  *FnIns
	dynamic_annotation bool
	arg_models         []ExprModel
}

func (fcac *_FnCallArgChecker) push_err_token(token lex.Token, key string, args ...any) {
	fcac.e.s.push_err(token, key, args...)
}

func (fcac *_FnCallArgChecker) push_err(key string, args ...any) {
	fcac.push_err_token(fcac.error_token, key, args...)
}

func (fcac *_FnCallArgChecker) get_params() []*ParamIns {
	if len(fcac.f.Params) > 0 && fcac.f.Params[0].Decl != nil && fcac.f.Params[0].Decl.Is_self() {
		return fcac.f.Params[1:] // Remove receiver parameter.
	}
	return fcac.f.Params
}

func (fcac *_FnCallArgChecker) check_counts(params []*ParamIns) (ok bool) {
	n := len(params)
	if n > 0 && params[0].Decl.Is_self() {
		n--
	}

	diff := n - len(fcac.args)
	switch {
	case diff == 0:
		return true

	case n > 0 && params[n-1].Decl.Variadic:
		return true

	case diff < 0 || diff > len(params):
		fcac.push_err("argument_overflow")
		return false
	}

	idents := ""
	for ; diff > 0; diff-- {
		idents += ", " + params[n-diff].Decl.Ident
	}
	idents = idents[2:] // Remove first separator.
	fcac.push_err("missing_expr_for", idents)

	return false
}

func (fcac *_FnCallArgChecker) check_arg(p *ParamIns, arg *Data, error_token lex.Token) (ok bool) {
	if fcac.dynamic_annotation {
		dta := _DynamicTypeAnnotation{
			e:           fcac.e,
			f:           fcac.f,
			p:           p,
			a:           arg,
			error_token: error_token,
		}
		ok = dta.annotate()
		if !ok {
			fcac.push_err_token(error_token, "dynamic_type_annotation_failed")
			return false
		}
	}

	fcac.e.s.check_validity_for_init_expr(p.Decl.Mutable, p.Kind, arg, error_token)
	fcac.e.s.check_assign_type(p.Kind, arg, error_token, false)
	return true
}

func (fcac *_FnCallArgChecker) push(p *ParamIns, arg *ast.Expr) (ok bool) {
	var d *Data
	if !fcac.dynamic_annotation {
		old := fcac.e.prefix
		fcac.e.prefix = p.Kind
		d = fcac.e.eval_expr_kind(arg.Kind)
		fcac.e.prefix = old
	} else {
		d = fcac.e.eval_expr_kind(arg.Kind)
	}
	if d == nil {
		return false
	}
	fcac.arg_models = append(fcac.arg_models, d.Model)
	return fcac.check_arg(p, d, arg.Token)
}

func (fcac *_FnCallArgChecker) push_variadic(p *ParamIns, i int) (ok bool) {
	ok = true
	variadiced := false
	more := i+1 < len(fcac.args)
	model := &SliceExprModel{
		Elem_kind: p.Kind,
	}

	old := fcac.e.prefix
	fcac.e.prefix = p.Kind
	for ; i < len(fcac.args); i++ {
		arg := fcac.args[i]
		d := fcac.e.eval_expr_kind(arg.Kind)
		if d == nil {
			ok = false
			continue
		}

		if d.Variadiced {
			variadiced = true
			d.Variadiced = false // For ignore assignment checking error.
			switch d.Model.(type) {
			case *SliceExprModel:
				model = d.Model.(*SliceExprModel)
				model.Elem_kind = p.Kind

			default:
				model = nil
				fcac.arg_models = append(fcac.arg_models, d.Model)
			}
		} else {
			model.Elems = append(model.Elems, d.Model)
		}

		ok = fcac.check_arg(p, d, arg.Token) && ok
	}
	fcac.e.prefix = old

	if variadiced && more {
		fcac.push_err("more_args_with_variadiced")
	}

	if model != nil {
		fcac.arg_models = append(fcac.arg_models, model)
	}
	return ok
}

func (fcac *_FnCallArgChecker) check_args(params []*ParamIns) (ok bool) {
	ok = true
	i := 0
iter:
	for i < len(params) {
		p := params[i]
		switch {
		case p.Decl.Is_self():
			// Ignore self.

		case p.Decl.Variadic:
			ok = fcac.push_variadic(p, i) && ok
			break iter // Variadiced parameters always last.

		case len(fcac.args) <= i:
			ok = false
			break iter

		default:
			ok = fcac.push(p, fcac.args[i]) && ok
		}
		i++
	}

	return ok
}

func (fcac *_FnCallArgChecker) check_dynamic_type_annotation() (ok bool) {
	for _, g := range fcac.f.Generics {
		if g == nil {
			fcac.push_err("dynamic_type_annotation_failed")
			return false
		}
	}
	return true
}

func (fcac *_FnCallArgChecker) check() (ok bool) {
	params := fcac.get_params()
	ok = fcac.check_counts(params)
	if !ok {
		return false
	}

	ok = fcac.check_args(params)
	if ok && fcac.dynamic_annotation {
		ok = fcac.check_dynamic_type_annotation()
	}

	return ok
}

type _StructLitChecker struct {
	e           *_Eval
	error_token lex.Token
	s           *StructIns
	args        []*StructArgExprModel
}

func (slc *_StructLitChecker) push_err(token lex.Token, key string, args ...any) {
	slc.e.push_err(token, key, args...)
}

func (slc *_StructLitChecker) push_match(f *FieldIns, d *Data, error_token lex.Token) {
	slc.args = append(slc.args, &StructArgExprModel{
		Field: f,
		Expr:  d.Model,
	})
	slc.e.s.check_validity_for_init_expr(f.Decl.Mutable, f.Kind, d, error_token)
	slc.e.s.check_assign_type(f.Kind, d, error_token, false)
}

func (slc *_StructLitChecker) check_pair(pair *ast.FieldExprPair, exprs []ast.ExprData) {
	// Check existing.
	f := slc.s.Find_field(pair.Field.Kind)
	if f == nil {
		slc.push_err(pair.Field, "ident_not_exist", pair.Field.Kind)
		return
	}

	// Check duplications.
dup_lookup:
	for _, expr := range exprs {
		switch expr.(type) {
		case *ast.FieldExprPair:
			dpair := expr.(*ast.FieldExprPair)
			switch {
			case pair == dpair:
				break dup_lookup

			case pair.Field.Kind == dpair.Field.Kind:
				slc.push_err(pair.Field, "already_has_expr", pair.Field.Kind)
				break dup_lookup
			}
		}
	}

	d := slc.e.eval_expr_kind(pair.Expr)
	if d == nil {
		return
	}
	slc.push_match(f, d, pair.Field)
}

func (slc *_StructLitChecker) ready_exprs(exprs []ast.ExprData) bool {
	ok := true
	for i, expr := range exprs {
		switch expr.(type) {
		case *ast.KeyValPair:
			pair := expr.(*ast.KeyValPair)
			switch pair.Key.(type) {
			case *ast.IdentExpr:
				// Ok

			default:
				slc.push_err(pair.Colon, "invalid_syntax")
				ok = false
				continue
			}

			exprs[i] = &ast.FieldExprPair{
				Field: pair.Key.(*ast.IdentExpr).Token,
				Expr:  pair.Val,
			}
		}
	}

	return ok
}

func (slc *_StructLitChecker) check(exprs []ast.ExprData) {
	if len(exprs) == 0 {
		return
	}

	if !slc.ready_exprs(exprs) {
		return
	}

	paired := false
	for i, expr := range exprs {
		switch expr.(type) {
		case *ast.FieldExprPair:
			pair := expr.(*ast.FieldExprPair)
			if i > 0 && !paired {
				slc.push_err(pair.Field, "invalid_syntax")
			}
			paired = true
			slc.check_pair(pair, exprs)

		case *ast.Expr:
			e := expr.(*ast.Expr)
			if paired {
				slc.push_err(e.Token, "argument_must_target_to_field")
			}
			if i >= len(slc.s.Fields) {
				slc.push_err(e.Token, "argument_overflow")
				continue
			}

			d := slc.e.eval_expr_kind(e.Kind)
			if d == nil {
				continue
			}

			field := slc.s.Fields[i]
			slc.push_match(field, d, e.Token)
		}
	}

	// Check missing arguments for fields.
	if !paired {
		n := len(slc.s.Fields)
		diff := n - len(exprs)
		switch {
		case diff <= 0:
			return
		}

		idents := ""
		for ; diff > 0; diff-- {
			idents += ", " + slc.s.Fields[n-diff].Decl.Ident
		}
		idents = idents[2:] // Remove first separator.
		slc.push_err(slc.error_token, "missing_expr_for", idents)
	}
}

// Range checker and setter.
type _RangeChecker struct {
	sc   *_ScopeChecker
	rang *ast.RangeKind
	kind *RangeIter
	d    *Data
}

func (rc *_RangeChecker) build_var(decl *ast.VarDecl) *Var {
	v := build_var(decl)
	return v
}

func (rc *_RangeChecker) set_size_key() {
	if rc.rang.Key_a == nil || lex.Is_ignore_ident(rc.rang.Key_a.Ident) {
		return
	}

	rc.kind.Key_a = rc.build_var(rc.rang.Key_a)
	rc.kind.Key_a.Kind = &TypeSymbol{
		Kind: &TypeKind{
			kind: build_prim_type(types.TypeKind_INT),
		},
	}
}

func (rc *_RangeChecker) check_slice() {
	rc.set_size_key()
	if rc.rang.Key_b == nil || lex.Is_ignore_ident(rc.rang.Key_b.Ident) {
		return
	}

	slc := rc.d.Kind.Slc()
	rc.kind.Key_b = rc.build_var(rc.rang.Key_b)
	rc.kind.Key_b.Kind = &TypeSymbol{Kind: slc.Elem}
	rc.sc.s.check_validity_for_init_expr(rc.kind.Key_b.Mutable, rc.kind.Key_b.Kind.Kind, rc.d, rc.rang.In_token)
}

func (rc *_RangeChecker) check_array() {
	rc.set_size_key()
	if rc.rang.Key_b == nil || lex.Is_ignore_ident(rc.rang.Key_b.Ident) {
		return
	}

	arr := rc.d.Kind.Arr()
	rc.kind.Key_b = rc.build_var(rc.rang.Key_b)
	rc.kind.Key_b.Kind = &TypeSymbol{Kind: arr.Elem}
	rc.sc.s.check_validity_for_init_expr(rc.kind.Key_b.Mutable, rc.kind.Key_b.Kind.Kind, rc.d, rc.rang.In_token)
}

func (rc *_RangeChecker) check_map_key_a() {
	if rc.rang.Key_a == nil || lex.Is_ignore_ident(rc.rang.Key_a.Ident) {
		return
	}

	m := rc.d.Kind.Map()
	rc.kind.Key_a = rc.build_var(rc.rang.Key_a)
	rc.kind.Key_a.Kind = &TypeSymbol{Kind: m.Key}

	d := *rc.d
	d.Kind = m.Key
	rc.sc.s.check_validity_for_init_expr(rc.kind.Key_a.Mutable, rc.kind.Key_a.Kind.Kind, &d, rc.rang.In_token)
}

func (rc *_RangeChecker) check_map_key_b() {
	if rc.rang.Key_b == nil || lex.Is_ignore_ident(rc.rang.Key_b.Ident) {
		return
	}

	m := rc.d.Kind.Map()
	rc.kind.Key_b = rc.build_var(rc.rang.Key_b)
	rc.kind.Key_b.Kind = &TypeSymbol{Kind: m.Val}

	d := *rc.d
	d.Kind = m.Val
	rc.sc.s.check_validity_for_init_expr(rc.kind.Key_b.Mutable, rc.kind.Key_b.Kind.Kind, &d, rc.rang.In_token)
}

func (rc *_RangeChecker) check_map() {
	rc.check_map_key_a()
	rc.check_map_key_b()
}

func (rc *_RangeChecker) check_str() {
	rc.set_size_key()
	if rc.rang.Key_b == nil || lex.Is_ignore_ident(rc.rang.Key_b.Ident) {
		return
	}

	rc.kind.Key_b = rc.build_var(rc.rang.Key_b)
	rc.kind.Key_b.Kind = &TypeSymbol{
		Kind: &TypeKind{
			kind: build_prim_type(types.TypeKind_U8),
		},
	}
}

func (rc *_RangeChecker) check() bool {
	if rc.d.Kind.Ref() != nil {
		rc.d.Model = &ExplicitDerefExprModel{
			Expr: rc.d.Model,
		}
		rc.d.Kind = rc.d.Kind.Ref().Elem
	}

	switch {
	case rc.d.Kind.Slc() != nil:
		rc.check_slice()
		return true

	case rc.d.Kind.Arr() != nil:
		rc.check_array()
		return true

	case rc.d.Kind.Map() != nil:
		rc.check_map()
		return true
	}

	prim := rc.d.Kind.Prim()
	if prim != nil && prim.Is_str() {
		rc.check_str()
		return true
	}

	rc.sc.s.push_err(rc.rang.In_token, "iter_range_require_enumerable_expr")
	return false
}

// Return type checker for return statements.
type _RetTypeChecker struct {
	sc          *_ScopeChecker
	f           *FnIns
	types       []*TypeKind // Return types.
	exprs       []*Data     // Return expressions.
	vars        []*Var      // Return variables.
	error_token lex.Token
}

func (rtc *_RetTypeChecker) prepare_types() {
	rtc.types = get_fn_result_types(rtc.f)
}

func (rtc *_RetTypeChecker) prepare_exprs(d *Data) {
	if d == nil {
		return
	}

	if d.Kind.Tup() == nil {
		rtc.exprs = append(rtc.exprs, d)
		return
	}

	rtc.exprs = get_datas_from_tuple_data(d)
}

func (rtc *_RetTypeChecker) ret_vars() {
	if rtc.f.Decl.Is_void() {
		return
	}
	rtc.vars = make([]*Var, len(rtc.f.Decl.Result.Idents))

	root_scope := rtc.sc.get_root()

	j := 0
	for i, ident := range rtc.f.Decl.Result.Idents {
		if !lex.Is_ignore_ident(ident.Kind) {
			rtc.vars[i] = root_scope.table.Vars[j]
			j++
		} else {
			rtc.vars[i] = &Var{
				Ident: lex.IGNORE_IDENT,
				Kind:  &TypeSymbol{Kind: rtc.types[i]},
			}
		}
	}

	// Not exist any real variable.
	if j == 0 {
		rtc.vars = nil
	}
}

func (rtc *_RetTypeChecker) check_exprs() {
	for i, d := range rtc.exprs {
		if i >= len(rtc.types) {
			break
		}

		t := rtc.types[i]
		if !d.Mutable && is_mut(d.Kind) {
			rtc.sc.s.push_err(rtc.error_token, "ret_with_mut_typed_non_mut")
			return
		}

		ac := _AssignTypeChecker{
			s:           rtc.sc.s,
			dest:        t,
			d:           d,
			error_token: rtc.error_token,
		}
		ac.check()
	}
}

func (rtc *_RetTypeChecker) check(d *Data) bool {
	rtc.prepare_types()
	rtc.prepare_exprs(d)
	rtc.ret_vars()

	n := len(rtc.exprs)
	if n == 0 && !rtc.f.Decl.Is_void() {
		if !rtc.f.Decl.Any_var() {
			rtc.sc.s.push_err(rtc.error_token, "require_ret_expr")
			return false
		}
		return true
	}

	if n > 0 && rtc.f.Decl.Is_void() {
		rtc.sc.s.push_err(rtc.error_token, "void_function_ret_expr")
		return false
	}

	if n > len(rtc.types) {
		rtc.sc.s.push_err(rtc.error_token, "overflow_ret")
	} else if n < len(rtc.types) {
		rtc.sc.s.push_err(rtc.error_token, "missing_multi_ret")
	}

	rtc.check_exprs()
	return true
}
