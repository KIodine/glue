package glue

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

// NOTE: `FieldByName` is slow but cacheable, yet the side effect of using mutex
// lock to protect the cache cancels out the benefit of caching it.

const (
	// struct tag
	glueTagKey = "glue"
	// tag attr
	attrIgnr = "-"
)

var (
	ErrGlue              = errors.New("GlueError") // the base error of package `glue`.
	ErrNotPtrToStruct    = fmt.Errorf("%w: one of the arguments is not pointer to struct", ErrGlue)
	ErrNotFunction       = fmt.Errorf("%w: the `converter` fed in is not a function", ErrGlue)
	ErrIncompatSignature = fmt.Errorf("%w: function signature incompatible", ErrGlue)
	ErrUnsatisfiedField  = fmt.Errorf("%w: unsatisfied field", ErrGlue)
)

type fieldAttr struct {
	Alias     string // The name a field used to pull/push from/to another struct.
	FieldMeta reflect.StructField
}
type typeAttr struct {
	ExportedNum int // the number of available/settable fields.
	FieldAttrs  []*fieldAttr
}

// The key of map of conversion functions.
type typeMapKey struct {
	Dst reflect.Type
	Src reflect.Type
}

var (
	cacheLock sync.Mutex
	attrCache = make(map[reflect.Type]*typeAttr, 32)
	convLock  sync.RWMutex
	typeMap   = make(map[typeMapKey]reflect.Value, 32)
)

// Glue copies fields from src to dst that have the same name and the same type.
// The major target of `Glue` is to satisfy the need of dst structure with best
// effort and does not require the two structures being the same "size"(have
// equally numbers of fields).
// `Glue` assumes that dst struct serves as a temporary storage of data and does
// not perform deepcopy on each field that is being copied from.
func Glue(dst, src interface{}, opts ...GlueOption) error {
	var (
		options glueOptions
		exist   bool
		vdst    = reflect.ValueOf(dst)
		vsrc    = reflect.ValueOf(src)
		// reflect stuffs
		dstStruct, srcStruct       reflect.Value
		dstType, srcType           reflect.Type
		alias                      string
		dstFieldMeta, srcFieldMeta reflect.StructField
		dstField, srcField         reflect.Value
		fAttrs                     *typeAttr
	)
	if !isValidPtrToStruct(&vdst) || !isValidPtrToStruct(&vsrc) {
		return ErrNotPtrToStruct
	}

	for _, opt := range opts {
		opt.apply(&options)
	}

	dstStruct = vdst.Elem()
	srcStruct = vsrc.Elem()
	dstType = dstStruct.Type()
	srcType = srcStruct.Type()

	if options.FavorSource {
		fAttrs = getTypeAttr(srcType)
	} else {
		fAttrs = getTypeAttr(dstType)
	}

	nFields := fAttrs.ExportedNum

	for i := 0; i < nFields; i++ {
		fa := fAttrs.FieldAttrs[i]
		alias = fa.Alias

		if options.FavorSource {
			srcFieldMeta = fa.FieldMeta
			dstFieldMeta, exist = dstType.FieldByName(alias)
		} else {
			dstFieldMeta = fa.FieldMeta
			srcFieldMeta, exist = srcType.FieldByName(alias)
		}

		if !exist {
			if options.Strict {
				return fmt.Errorf("%w: %#v", ErrUnsatisfiedField, alias)
			}
			continue
		}

		// test if types are strictly equal.
		var (
			fconv  reflect.Value
			doConv bool
		)
		if dstFieldMeta.Type != srcFieldMeta.Type {
			mk := typeMapKey{
				Dst: dstFieldMeta.Type,
				Src: srcFieldMeta.Type,
			}
			convLock.RLock()
			fconv, exist = typeMap[mk]
			convLock.RUnlock()
			doConv = exist
			if !exist {
				continue
			}
		}
		dstField = dstStruct.FieldByIndex(dstFieldMeta.Index)
		srcField = srcStruct.FieldByIndex(srcFieldMeta.Index)

		if !dstField.CanSet() || !srcField.CanSet() {
			continue
		}
		var v reflect.Value
		if !doConv {
			v = reflect.ValueOf(srcField.Interface())
		} else {
			ret := fconv.Call(
				[]reflect.Value{reflect.ValueOf(srcField.Interface())},
			)
			v = reflect.ValueOf(ret[0].Interface())
		}
		dstField.Set(v)

	}
	return nil
}

func isValidPtrToStruct(rv *reflect.Value) bool {
	if rv.Kind() != reflect.Ptr {
		return false
	}
	if rv.IsNil() {
		return false
	}
	ind := rv.Elem()
	//lint:ignore S1008 make sure there is one way to be correct.
	if ind.Kind() != reflect.Struct {
		return false
	}
	return true
}

// isValidIdentifier checks if a string is a valid golang identifier.
func isValidIdentifier(s string) bool {
	var (
		r  rune
		sz int
	)
	if len(s) == 0 {
		return false
	}
	r, sz = utf8.DecodeRuneInString(s)
	if r == utf8.RuneError || !unicode.IsLetter(r) || r == rune('_') {
		return false
	}
	s = s[sz:] // "step" forward.
iter_rune:
	for {
		r, sz = utf8.DecodeRuneInString(s)
		if r == utf8.RuneError {
			if sz == 0 {
				// string is consumed.
				break iter_rune
			}
			return false
		}
		if !(unicode.IsLetter(r) || r == rune('_') ||
			unicode.IsDigit(r) || unicode.IsNumber(r)) {
			return false
		}
		s = s[sz:]
	}
	return true
}

// RegConv creates a conversion mapping from src type to dst type.
// To create a mapping between two types, user can pass zero value of certain
// type as hint and a converter function that takes a value of src type and
// outputs dst type, this function checks the converter function have the
// correct function signature, if the converter is not a function or does not
// have the right signature, `RegConv` returns corresponding error,
// on successful register, this function returns nil.
func RegConv(tDst, tSrc, converter interface{}) error {
	typeDst := reflect.ValueOf(tDst).Type()
	typeSrc := reflect.ValueOf(tSrc).Type()
	vConvFunc := reflect.ValueOf(converter)
	if vConvFunc.Kind() != reflect.Func {
		return ErrNotFunction
	}
	vfunc := reflect.FuncOf(
		[]reflect.Type{typeSrc},
		[]reflect.Type{typeDst},
		false,
	)
	if vConvFunc.Type() != vfunc {
		return ErrIncompatSignature
	}

	convLock.Lock()
	defer convLock.Unlock()
	mk := typeMapKey{
		Dst: typeDst,
		Src: typeSrc,
	}
	typeMap[mk] = vConvFunc

	return nil
}

// DeregConv deregisters the conversion mapping between two types.
func DeregConv(tDst, tSrc interface{}) {
	typeDst := reflect.ValueOf(tDst).Type()
	typeSrc := reflect.ValueOf(tSrc).Type()
	convLock.Lock()
	defer convLock.Unlock()
	mk := typeMapKey{
		Dst: typeDst,
		Src: typeSrc,
	}
	delete(typeMap, mk)
}

// MustRegConv is a shorthand allow user register conversion map on initialize,
// it panics if parameters does not meet the requirement of `RegConv`.
func MustRegConv(tDst, tSrc, converter interface{}) bool {
	err := RegConv(tDst, tSrc, converter)
	if err != nil {
		panic(err)
	}
	return true
}

// getTypeAttr returns cache of `*typeAttr`, it builds attribute if no cache
// can be acquired.
func getTypeAttr(t reflect.Type) *typeAttr {
	var (
		dstAttrs *typeAttr
		rawAttrs string
		fAttr    *fieldAttr
		exist    bool
	)
	dstNumFields := t.NumField()

	cacheLock.Lock()
	defer cacheLock.Unlock()
	dstAttrs, exist = attrCache[t]
	if exist {
		return dstAttrs
	}
	dstAttrs = new(typeAttr)
	for i := 0; i < dstNumFields; i++ {
		fieldMeta := t.Field(i)
		// `attrMap` and `fieldArr` only records "available" fields:
		// 1) natively exported.
		// 2) not tagged as ignored.

		// backport of method `IsExported(1.17-)`
		if !(fieldMeta.PkgPath == "") {
			continue
		}

		rawAttrs, exist = fieldMeta.Tag.Lookup(glueTagKey)
		if !exist {
			// early out, has no glue tag, pullname is field name.
			fAttr = &fieldAttr{
				Alias:     fieldMeta.Name,
				FieldMeta: fieldMeta,
			}
			dstAttrs.ExportedNum++
			dstAttrs.FieldAttrs = append(dstAttrs.FieldAttrs, fAttr)
			continue
		}
		if rawAttrs == attrIgnr {
			// ignore the field.
			continue
		}
		if !isValidIdentifier(rawAttrs) {
			panic(fmt.Errorf("%q is not a valid identifier", rawAttrs))
		}
		fAttr = &fieldAttr{
			FieldMeta: fieldMeta,
		}

		// do parse below this line if allow multiple attributes.
		// rawAttrs = strings.Split()
		// and do analyze, if required so.
		fAttr.Alias = rawAttrs

		dstAttrs.ExportedNum++
		dstAttrs.FieldAttrs = append(dstAttrs.FieldAttrs, fAttr)
	}
	attrCache[t] = dstAttrs

	return dstAttrs
}
