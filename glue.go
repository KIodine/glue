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

/* TODO:
- [ ] allow ignore fields.
- [ ] allow push/pull fields: `glue:"pull=Alpha,push=Beta"`
- [ ] allow get from method?
- [ ] cache tag analyze result? `map[reflect.Type]*GluCache`
*/

/* Glue tries to merge two structs by copying fields from dst to src that have
the same name and the same type. The major target of `Glue` is to satisfy the
need of dst structure with best effort and does not require the two structure
being "the same size"(have equally number of fields.). */
func Glue(dst, src interface{}) error {
	var (
		exist bool
		vdst  = reflect.ValueOf(dst)
		vsrc  = reflect.ValueOf(src)
		/* reflect stuffs */
		dstStruct, srcStruct       reflect.Value
		dstType, srcType           reflect.Type
		nameSrcField, nameDstField string
		dstFieldMeta, srcFieldMeta reflect.StructField
		dstField, srcField         reflect.Value
		nameByTag                  string
	)
	if !isValidPtrToStruct(&vdst) || !isValidPtrToStruct(&vsrc) {
		return ErrTypeIncompat
	}
	dstStruct = vdst.Elem()    // reflect.Indirect(vdst)
	dstType = dstStruct.Type() //reflect.TypeOf(dstStruct)
	srcStruct = vsrc.Elem()    //reflect.Indirect(vsrc)
	srcType = srcStruct.Type() //reflect.TypeOf(srcStruct)

	dstNumFields := dstType.NumField()

	/* for each field have the same name and same type, copy value. */
	for i := 0; i < dstNumFields; i++ {
		dstFieldMeta = dstType.Field(i)
		nameSrcField = dstFieldMeta.Name
		nameDstField = nameSrcField
		/* only set public fields. This should be the same on the src. */
		/* backport of method `IsExported(1.17~)` */
		if !(dstFieldMeta.PkgPath == "") {
			continue
		}
		/* allow gluing src fields specified by tag, this takes priority. */
		/* X: allow strict src format "<struct>.<field>" ? */
		nameByTag, exist = dstFieldMeta.Tag.Lookup(glueTagKey)
		if exist {
			if !isValidIdentifier(nameByTag) {
				/* if the tag is not a valid identifier, it would never had a
				chance to be satisfied or to satisfy an untagged field. */
				panic(fmt.Errorf("%q is not a valid identifier", nameByTag))
			}
			nameSrcField = nameByTag
		}
		srcFieldMeta, exist = srcType.FieldByName(nameSrcField)
		/* no corresponding field on src. */
		if !exist {
			continue
		}
		/* test if type equal */
		/*
			There are several unclear question/problems:
			1) There are type aliases. Type does strict compare, or this is
				exactly what we desire? -> Yes.
			2) Can type compare directly? even they are just interfaces? -> Yes.
		*/
		if dstFieldMeta.Type != srcFieldMeta.Type {
			continue
		}
		/*
			at this point we've guarenteed the field must exist:
			1) the dst field must exist
			2) we can get the field on src by the name from dst field
		*/
		dstField = dstStruct.FieldByName(nameDstField)
		srcField = srcStruct.FieldByName(nameSrcField)
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
	ind := rv.Elem() //reflect.Indirect(*rv)
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
	"Letter". */
	if r == utf8.RuneError || !unicode.IsLetter(r) {
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
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsNumber(r)) {
			return false
		}
		s = s[sz:]
	}
	return true
}
