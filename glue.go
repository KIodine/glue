package glue

import (
	"errors"
	"reflect"
)

var (
	ErrTypeIncompat = errors.New("types are not compatible")
)

/* Glue tries to merge two structs by copying fields from dst to src that have
the same name and the same type. */
func Glue(dst, src interface{}) error {
	var (
		vdst = reflect.ValueOf(dst)
		vsrc = reflect.ValueOf(src)
		dstStruct,
		srcStruct reflect.Value
		srcType,
		dstType reflect.Type
	)
	if !isValidPtrToStruct(&vdst) || !isValidPtrToStruct(&vsrc) {
		return ErrTypeIncompat
	}
	dstStruct = vdst.Elem()    // reflect.Indirect(vdst)
	dstType = dstStruct.Type() //reflect.TypeOf(dstStruct)
	srcStruct = vsrc.Elem()    //reflect.Indirect(vsrc)
	srcType = srcStruct.Type() //reflect.TypeOf(srcStruct)
	//dstFields = dstStruct.NumField()
	dstNumFields := dstType.NumField()
	//dstType.Name()

	/* for each field have the same name and same type, copy value. */
	for i := 0; i < dstNumFields; i++ {
		dstFieldMeta := dstType.Field(i)
		name := dstFieldMeta.Name
		/* only set public fields. This should be the same on the src. */
		if !dstFieldMeta.IsExported() {
			continue
		}
		/* X: or allow compare tag `glue:"<name>"`? */
		srcFieldMeta, exist := srcType.FieldByName(name)
		/* no corresponding field on src. */
		if !exist {
			continue
		}
		/* test if type equal */
		/*
			There are several unclear question/problems:
			1) There are type aliases. Type does strict compare, or this is
				exactly what we desire?
			2) Can type compare directly? even they are just interfaces?
		*/
		if dstFieldMeta.Type != srcFieldMeta.Type {
			continue
		}
		/*
			at this point we've guarenteed the field must exist:
			1) the dst field must exist
			2) we can get the field on src by the name from dst field
		*/
		dstField := dstStruct.FieldByName(name)
		srcField := srcStruct.FieldByName(name)
		/* require both side can set. */
		if !dstField.CanSet() || !srcField.CanSet() {
			continue
		}
		/* does this copy struct field recursivly/deep copy? */
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
	ind := reflect.Indirect(*rv)
	/* don't care it's zero. */
	//lint:ignore S1008 I know what I'm doing.
	if ind.Kind() != reflect.Struct {
		return false
	}
	return true
}
