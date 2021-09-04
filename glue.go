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
	/* X: better name? */
	typeMapKey struct {
		dst reflect.Type
		src reflect.Type
	}
)

var (
	cacheLock sync.Mutex
	typeCache = make(map[reflect.Type]fieldCache, 32)
	convLock  sync.RWMutex
	typeMap   = make(map[typeMapKey]reflect.Value, 32)
)

/* PROPASAL:
- [ ] allow auto convertion if type mapping is registered.
- [ ] do deepcopy: `glue:"deep"`
- [ ] allow get from method? Only methods require no parameter and must have
	matching type.
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
		//srcTagName                 string
	)
	if !isValidPtrToStruct(&vdst) || !isValidPtrToStruct(&vsrc) {
		return ErrTypeIncompat
	}
	dstStruct = vdst.Elem()
	srcStruct = vsrc.Elem()
	dstType = dstStruct.Type()
	srcType = srcStruct.Type()

	dstNumFields := dstType.NumField()
	/* X: if do cache tag parse results, do it here and analyze all fields at
	once, protected by mutex lock. `parseTag(v *reflect.Value) fieldCache`
	glue tags have no effect on the src side. */
	cacheLock.Lock()
	var (
		dstAttrs fieldCache
		rawAttrs string
		fAttr    *fieldAttr
	)
	dstAttrs, exist = typeCache[dstType]
	if !exist {
		// TODO: separate as helper routine.
		dstAttrs = make(fieldCache, 8)
		for i := 0; i < dstNumFields; i++ {
			fAttr = &fieldAttr{}
			dstFieldMeta = dstType.Field(i)
			dstAttrs[dstFieldMeta.Name] = fAttr

			/* backport of method `IsExported(1.17-)` */
			if !(dstFieldMeta.PkgPath == "") {
				/* leave pull name blank */
				continue
			}

			rawAttrs, exist = dstFieldMeta.Tag.Lookup(glueTagKey)
			if !exist {
				/* has no glue tag */
				fAttr.PullName = dstFieldMeta.Name
				continue
			}
			if rawAttrs == attrIgnr {
				/* leave pull name blank */
				continue
			}
			if !isValidIdentifier(rawAttrs) {
				cacheLock.Unlock()
				panic(fmt.Errorf("%q is not a valid identifier", rawAttrs))
			}
			fAttr.PullName = rawAttrs
			/* do below this line parse if allow multiple attributes. */
			//rawAttrs = strings.Split()
		}
		typeCache[dstType] = dstAttrs
	}
	cacheLock.Unlock()

	/* for each field have the same name and same type, copy value. */
	for i := 0; i < dstNumFields; i++ {
		dstFieldMeta = dstType.Field(i)
		dstFieldName = dstFieldMeta.Name
		srcFieldName = dstAttrs[dstFieldName].PullName
		/* only set public fields. This should be the same on the src. */

		if srcFieldName == "" {
			continue
		}
		/* allow gluing src fields specified by tag, this takes priority. */
		/* X: allow strict src format "<struct>.<field>" ? */
		srcFieldMeta, exist = srcType.FieldByName(srcFieldName)
		/* no corresponding field on src. */
		if !exist {
			continue
		}
		/* test if types are strictly equal */
		var fconv reflect.Value
		if dstFieldMeta.Type != srcFieldMeta.Type {
			// TODO: try convert src to dst.
			mk := typeMapKey{
				dst: dstFieldMeta.Type,
				src: srcFieldMeta.Type,
			}
			convLock.RLock()
			fconv, exist = typeMap[mk]
			convLock.RUnlock()
			if !exist {
				continue
			}
		}
		/*
			at this point we've guarenteed the field must exist:
			1) the dst field must exist
			2) we can get the field on src by the name from dst field
		*/
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
		if reflect.ValueOf(fconv).IsZero() {
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

func getAttrCache(t reflect.Type) fieldCache {
	dstNumFields := t.NumField()
	/* X: if do cache tag parse results, do it here and analyze all fields at
	once, protected by mutex lock. `parseTag(v *reflect.Value) fieldCache`
	glue tags have no effect on the src side. */
	cacheLock.Lock()
	defer cacheLock.Unlock()
	var (
		dstAttrs fieldCache
		rawAttrs string
		fAttr    *fieldAttr
		exist    bool
	)
	dstAttrs, exist = typeCache[t]
	if !exist {
		// TODO: separate as helper routine.
		dstAttrs = make(fieldCache, 8)
		for i := 0; i < dstNumFields; i++ {
			fAttr = &fieldAttr{}
			dstFieldMeta := t.Field(i)
			dstAttrs[dstFieldMeta.Name] = fAttr

			/* backport of method `IsExported(1.17-)` */
			if !(dstFieldMeta.PkgPath == "") {
				/* leave pull name blank */
				continue
			}

			rawAttrs, exist = dstFieldMeta.Tag.Lookup(glueTagKey)
			if !exist {
				/* has no glue tag */
				fAttr.PullName = dstFieldMeta.Name
				continue
			}
			if rawAttrs == attrIgnr {
				/* leave pull name blank */
				continue
			}
			if !isValidIdentifier(rawAttrs) {
				panic(fmt.Errorf("%q is not a valid identifier", rawAttrs))
			}
			fAttr.PullName = rawAttrs
			/* do below this line parse if allow multiple attributes. */
			//rawAttrs = strings.Split()
		}
		typeCache[t] = dstAttrs
	}

	return dstAttrs
}
