package glue

import (
	"errors"
	"reflect"
)

const glueTagKey = "glue"

var (
	ErrTypeIncompat = errors.New("types are not compatible")
)

/* Glue tries to merge two structs by copying fields from dst to src that have
the same name and the same type. */
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
		if !dstFieldMeta.IsExported() {
			continue
		}
		/* allow gluing src fields specified by tag, this takes priority. */
		/* X: allow strict src format "<struct>.<field>" ? */
		nameByTag, exist = dstFieldMeta.Tag.Lookup(glueTagKey)
		if exist {
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
