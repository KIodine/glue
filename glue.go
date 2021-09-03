package glue

import (
	"errors"
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"
)

const glueTagKey = "glue"

var (
	ErrTypeIncompat = errors.New("types are not compatible")
)

/* PROPASAL:
- [ ] allow type conversion map
	``` prototype
	func RegConvertionMap(tDst, tSrc, cb interface) error {}
	```
- [ ] allow get from method? Only methods require no parameter and must have
	matching type.
	- [ ] allow push/pull fields: `glue:"pull=Alpha,push=Beta"`
	+ Reject the idea of `push` attr.
	- Allow override unexported field? -> No
- [ ] do deepcopy: `glue:"deep"`
- [X] cache tag analyze result?
- [X] allow ignore fields, ex:
	- ignore both: `glue:"pull=-,push=-"` or `glue:"-"`
		```
		type attrKey string
		const (
			attrPull attrKey = "pull"
			attrPush attrKey = "push"
			attrIgnr attrKey = "-"
			//attrDeep attrKey = "deep"
		)
		```
	- panic if: exported duplicated name, duplicated tag hint.
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
		srcTagName                 string
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

	/* for each field have the same name and same type, copy value. */
	for i := 0; i < dstNumFields; i++ {
		dstFieldMeta = dstType.Field(i)
		srcFieldName = dstFieldMeta.Name
		dstFieldName = srcFieldName
		/* only set public fields. This should be the same on the src. */
		/* backport of method `IsExported(1.17-)` */
		if !(dstFieldMeta.PkgPath == "") {
			continue
		}
		/* allow gluing src fields specified by tag, this takes priority. */
		/* X: allow strict src format "<struct>.<field>" ? */
		srcTagName, exist = dstFieldMeta.Tag.Lookup(glueTagKey)
		if exist {
			if srcTagName == "-" {
				continue
			}
			/* X: validate only once? how? */
			if !isValidIdentifier(srcTagName) {
				/* if the tag is not a valid identifier, it would never had a
				chance to be satisfied or to satisfy an untagged field. */
				panic(fmt.Errorf("%q is not a valid identifier", srcTagName))
			}
			srcFieldName = srcTagName
		}
		srcFieldMeta, exist = srcType.FieldByName(srcFieldName)
		/* no corresponding field on src. */
		if !exist {
			continue
		}
		/* test if types are strictly equal */
		if dstFieldMeta.Type != srcFieldMeta.Type {
			continue
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
		dstField.Set(reflect.ValueOf(srcField.Interface()))

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
