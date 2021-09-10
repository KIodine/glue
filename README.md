# Glue
## Introduction
`Glue` is a simple tool that performs **shallow** copy between destination struct and source struct, from src field to dst field, the limitation is:
- Both fields are exported.
- Both fields have the same name or source have the field with name tagged on dst field.
- Both fields have (strictly) same type or registered conversion function.

`Glue` only accepts pointer to struct as parameter, if not it returns an `ErrNotPtrToStruct` error.

A field can be tagged with a valid identifier to specify the name of field to pull from source struct or tagged as ignore using "-", in that way the field will not be filled.
Currently `Glue` only accept one tag attribute, also the tag takes no effect when a field is in the source struct.

You can register a global conversion function using `RegConv`, the function must have signature that takes type of source field and outputs a value that have the same type as the destination field.
`RegConv` fails if the `converter` passed in is not a function or a function having incompatible signature.

`Glue` is thread-safe though internal uses some globally available cache.

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
glue.RegConv(int(0), float64(0), f64toInt)

f := &Foo{A: -1}
b := &Bar{A: 1024.0}
glue.Glue(f, b)

if f.A != 1024 {
    panic("unexpected")
}

```
Register conversion function is thread-safe(it is protected by a RWMutex), however it may take a short time to take effect.

You may use `MustRegConv` during global initialization, it panics if check fails.
```go
var _ = glue.MustRegConv(int(0), float64(0), f64toInt) // ok
var _ = glue.MustRegConv(float64(0), int(0), f64toInt) // fail on startup
```

## Possible improvements
- [ ] Optionally performs deep copy on reference types(slice, map, pointer to object).

## License
This library is distributed under MIT license.
