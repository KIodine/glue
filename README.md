# Glue
## Introduction
`Glue` is a simple tool that performs shallow copy between destination struct and source struct, from src field to dst field, the limitation is:
- Both fields are exported.
- Both fields have the same name or source have the field with name tagged on dst field.
- Both fields have (strictly) same type or registered conversion function.

`Glue` only accepts pointer to struct as parameter, if not it returns an `ErrNotPtrToStruct` error.

A field can be tagged with a valid identifier to specify the name of field to pull from source struct or tagged as ignore using "-", in that way the field will not be filled.
Currently `Glue` only accept one tag attribute, also the tag takes no effect when a struct is the source struct.

User can register a global conversion function that takes type of source field and outputs a value that have the same type as the destination field.

## Examples
The most basic usage:
```go
type (
    Foo struct {
        A int
    }
    Bar struct {
        A int
    }
)

/* ... */
f := &Foo{A: -1}
b := &Bar{A: 1999}

glue.Glue(f, b)
if f.A != b.A {
    panic("Unexpected")
}
```

With tag:
```go
type (
    Foo struct {
        A int `glue:"B"`
    }
    Bar struct {
        B int
    }
)
f := &Foo{A: -1}
b := &Bar{B: 1024}
glue.Glue(f, b)
if f.A != b.B {
    panic("Unexpected")
}
```

Register conversion function:
```go
type (
    Foo struct {
        A int
    }
    Bar struct {
        A float64
    }
)
f64toInt := func(f float64) int {
    return int(f)
}
glue.RegConversion(int(0), float64(0), f64toInt)

f := &Foo{A: -1}
b := &Bar{A: 1024.0}
glue.Glue(f, b)

if f.A != 1024 {
    panic("unexpected")
}

```

## License
This library is distributed under MIT license.
