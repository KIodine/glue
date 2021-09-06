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
// one possible solution is only parallelize the assigning part.

const (
	/* struct tag */
	glueTagKey = "glue"
	/* tag attr */
	attrIgnr = "-"
)

var (
	ErrNotPtrToStruct = errors.New("one of the arguments is not pointer to struct")
)

type (
	fieldAttr struct {
		PullFrom string
		Field    reflect.StructField
	}
	typeAttr struct {
		exportedNum int
		fieldArr    []*fieldAttr
	}
	/* X: better name? */
	typeMapKey struct {
		dst reflect.Type
		src reflect.Type
	}
)

var (
	cacheLock sync.Mutex
	typeCache = make(map[reflect.Type]*typeAttr, 32)
	// NOTE: if we're making `Glue` as method in the future, this will be a
	// private attribute.
	convLock sync.RWMutex
	typeMap  = make(map[typeMapKey]reflect.Value, 32)
)

/* PROPASAL:
- [ ] do deepcopy: `glue:"deep"`
- [ ] allow get from method? Only methods require no parameter and must have
	matching type.
- [ ] allow strict src name: `<struct>.<field>`?
*/

/* Glue copies fields from src to dst that have the same name and the same type.
The major target of `Glue` is to satisfy the need of dst structure with best
effort and does not require the two structures being the same "size"(have
equally numbers of fields).
`Glue` assumes that dst struct serves as a temporary storage of data and does
not perform deepcopy on each field that is being copied from. */
func Glue(dst, src interface{}) error {
	var (
		exist bool
		vdst  = reflect.ValueOf(dst)
		vsrc  = reflect.ValueOf(src)
		/* reflect stuffs */
		dstStruct, srcStruct       reflect.Value
		dstType, srcType           reflect.Type
		srcFieldName               string
		dstFieldMeta, srcFieldMeta reflect.StructField
		dstField, srcField         reflect.Value
		dstAttrs                   *typeAttr
	)
	if !isValidPtrToStruct(&vdst) || !isValidPtrToStruct(&vsrc) {
		return ErrNotPtrToStruct
	}
	dstStruct = vdst.Elem()
	srcStruct = vsrc.Elem()
	dstType = dstStruct.Type()
	srcType = srcStruct.Type()

	dstAttrs = getTypeAttr(dstType)
	dstNumFields := dstAttrs.exportedNum

	for i := 0; i < dstNumFields; i++ {
		dfa := dstAttrs.fieldArr[i]
		dstFieldMeta = dfa.Field
		srcFieldName = dfa.PullFrom

		srcFieldMeta, exist = srcType.FieldByName(srcFieldName)
		if !exist {
			continue
		}
		/* test if types are strictly equal */
		var (
			fconv  reflect.Value
			doConv bool
		)
		if dstFieldMeta.Type != srcFieldMeta.Type {
			mk := typeMapKey{
				dst: dstFieldMeta.Type,
				src: srcFieldMeta.Type,
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
	s = s[sz:] /* "step" forward. */
iter_rune:
	for {
		r, sz = utf8.DecodeRuneInString(s)
		if r == utf8.RuneError {
			if sz == 0 {
				/* string is consumed. */
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

// RegConversion creates a conversion mapping from src type to dst type.
// To create a mapping between two types, user can pass zero value of certain
// type as hint and a converter function that takes a value of src type and
// outputs dst type, this function checks the converter function have the
// correct function signature, if the converter is not a function or does not
// have the right signature, `RegConversion` returns false, on successful
// register, this function returns true.
func RegConversion(tDst, tSrc, converter interface{}) bool {
	typeDst := reflect.ValueOf(tDst).Type()
	typeSrc := reflect.ValueOf(tSrc).Type()
	vConvFunc := reflect.ValueOf(converter)
	if vConvFunc.Kind() != reflect.Func {
		return false
	}
	vfunc := reflect.FuncOf(
		[]reflect.Type{typeSrc},
		[]reflect.Type{typeDst},
		false,
	)
	if vConvFunc.Type() != vfunc {
		return false
	}

	convLock.Lock()
	defer convLock.Unlock()
	mk := typeMapKey{
		dst: typeDst,
		src: typeSrc,
	}
	typeMap[mk] = vConvFunc

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
	dstAttrs, exist = typeCache[t]
	if exist {
		return dstAttrs
	}
	dstAttrs = new(typeAttr)
	for i := 0; i < dstNumFields; i++ {
		fieldMeta := t.Field(i)
		// `attrMap` and `fieldArr` only records "available" fields:
		// 1) natively exported.
		// 2) not tagged as ignored.

		/* backport of method `IsExported(1.17-)` */
		if !(fieldMeta.PkgPath == "") {
			continue
		}

		rawAttrs, exist = fieldMeta.Tag.Lookup(glueTagKey)
		if !exist {
			/* early out, has no glue tag, pullname is field name. */
			fAttr = &fieldAttr{
				PullFrom: fieldMeta.Name,
				Field:    fieldMeta,
			}
			dstAttrs.exportedNum++
			dstAttrs.fieldArr = append(dstAttrs.fieldArr, fAttr)
			continue
		}
		if rawAttrs == attrIgnr {
			/* leave pull name blank */
			continue
		}
		if !isValidIdentifier(rawAttrs) {
			panic(fmt.Errorf("%q is not a valid identifier", rawAttrs))
		}
		fAttr = &fieldAttr{
			Field: fieldMeta,
		}

		// do parse below this line if allow multiple attributes.
		// rawAttrs = strings.Split()
		fAttr.PullFrom = rawAttrs
		// and do anylyze, if required so.

		dstAttrs.exportedNum++
		dstAttrs.fieldArr = append(dstAttrs.fieldArr, fAttr)
	}
	typeCache[t] = dstAttrs

	return dstAttrs
}
