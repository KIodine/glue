package glue

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

const (
	/* struct tag */
	glueTagKey = "glue"
	/* tag attr */
	attrIgnr = "-"
)

var (
	ErrTypeIncompat = errors.New("types are not compatible")
)

type (
	fieldAttr struct {
		PullName string
	}
	fieldCache map[string]*fieldAttr
	typeAttr   struct {
		exportedNum int
		attrMap     fieldCache
		fieldArr    []reflect.StructField
		// X: how to cache `FieldByName`?
	}
	/* X: better name? */
	typeMapKey struct {
		dst reflect.Type
		src reflect.Type
	}
)

// X: better name?
var (
	cacheLock sync.Mutex
	typeCache = make(map[reflect.Type]*typeAttr, 32)
	convLock  sync.RWMutex
	typeMap   = make(map[typeMapKey]reflect.Value, 32)
)

/* PROPASAL:
- [ ] allow auto convertion if type mapping is registered.
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
		srcFieldName, dstFieldName string
		dstFieldMeta, srcFieldMeta reflect.StructField
		dstField, srcField         reflect.Value
		dstAttrs                   *typeAttr
	)
	if !isValidPtrToStruct(&vdst) || !isValidPtrToStruct(&vsrc) {
		return ErrTypeIncompat
	}
	dstStruct = vdst.Elem()
	srcStruct = vsrc.Elem()
	dstType = dstStruct.Type()
	srcType = srcStruct.Type()

	dstAttrs = getTypeAttr(dstType)
	// X: also cache srcType?
	dstNumFields := dstAttrs.exportedNum //dstType.NumField()

	/* for each field have the same name and same type, copy value. */
	for i := 0; i < dstNumFields; i++ {
		/* golang reflect performs linear scan, repeated call results in
		O(n^2) time complexity. */
		// X: costly, but can cache(in `getAttrCache`)?
		dstFieldMeta = dstAttrs.fieldArr[i] //dstType.Field(i)
		dstFieldName = dstFieldMeta.Name
		srcFieldName = dstAttrs.attrMap[dstFieldName].PullName

		// part 1: test if
		// 1) src field exists
		// 2) two sides have same type or have registered conversion function.

		/* only set public fields. This should be the same on the src.
		the name is blank if dst is unexported or tagged as ignore. */
		if srcFieldName == "" {
			continue
		}
		/* allow gluing src fields specified by tag, this takes priority. */
		/* X: costly, but can cache? yes, but aware that this function dives
		into embedded members and does scan with breadth-first search. */
		// X: cache using {reflect.Type, string} composite key?
		srcFieldMeta, exist = srcType.FieldByName(srcFieldName)
		/* no corresponding field on src. */
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

		// part 2: test two fields can set and do conversion if required so.

		/*
			at this point we've guarenteed the field must exist:
			1) the dst field must exist
			2) we can get the field on src by the name from dst field
		*/
		/* `FieldByName` is costly, can do cache? */
		dstField = dstStruct.FieldByName(dstFieldName)
		srcField = srcStruct.FieldByName(srcFieldName)
		/* require both side can set.(probably just need to test one side) */
		/* Q: `CanSet` means `can mutate`? is there such thing a immutable field
		in struct? */
		if !dstField.CanSet() || !srcField.CanSet() {
			continue
		}
		/* does this copy struct field recursivly/deep copy? -> just shallow copy */
		/* X: maybe a `copyRecursive` */
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
	/* don't care it's zero. */
	//lint:ignore S1008 I know what I'm doing.
	if ind.Kind() != reflect.Struct {
		return false
	}
	return true
}

func isValidIdentifier(s string) bool {
	var (
		r  rune
		sz int
	)
	if len(s) == 0 {
		return false
	}
	r, sz = utf8.DecodeRuneInString(s)
	/* golang language spec requires the first unicode character must be a
	"Letter" or '_'. */
	if r == utf8.RuneError || !unicode.IsLetter(r) || r == rune('_') {
		/* either the string is empty and encountered an invalid utf8 encode are
		regarded as error. */
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

// NOTE: can have struct of types as key.
// TODO: give it a better name
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
	mk := typeMapKey{
		dst: typeDst,
		src: typeSrc,
	}
	typeMap[mk] = vConvFunc
	convLock.Unlock()
	return true
}

// getTypeAttr returns cache of `*typeAttr`, it does scan if no cache can
// be acquired.
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
	dstAttrs = new(typeAttr) //make(fieldCache, 8)
	dstAttrs.attrMap = make(fieldCache, 8)
	for i := 0; i < dstNumFields; i++ {
		dstFieldMeta := t.Field(i)
		// `attrMap` and `fieldArr` only records "available" fields:
		// 1) natively exported.
		// 2) not tagged as ignored.

		/* backport of method `IsExported(1.17-)` */
		if !(dstFieldMeta.PkgPath == "") {
			/* leave pull name blank */
			continue
		}

		rawAttrs, exist = dstFieldMeta.Tag.Lookup(glueTagKey)
		if !exist {
			/* early out, has no glue tag, pullname is field name. */
			fAttr = &fieldAttr{
				PullName: dstFieldMeta.Name,
			}
			dstAttrs.attrMap[dstFieldMeta.Name] = fAttr
			dstAttrs.exportedNum++
			dstAttrs.fieldArr = append(dstAttrs.fieldArr, dstFieldMeta)
			continue
		}
		if rawAttrs == attrIgnr {
			/* leave pull name blank */
			continue
		}
		if !isValidIdentifier(rawAttrs) {
			panic(fmt.Errorf("%q is not a valid identifier", rawAttrs))
		}
		fAttr = &fieldAttr{}
		// only record exported/not-ignored struct fields.

		/* do below this line parse if allow multiple attributes. */
		// rawAttrs = strings.Split()
		fAttr.PullName = rawAttrs
		// and do anylyze, if required so.

		dstAttrs.attrMap[dstFieldMeta.Name] = fAttr
		dstAttrs.exportedNum++
		dstAttrs.fieldArr = append(dstAttrs.fieldArr, dstFieldMeta)
	}
	typeCache[t] = dstAttrs

	return dstAttrs
}
