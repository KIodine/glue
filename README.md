# Glue
Gluing two structures.

## Table Of Content
- [Introduction](#introduction)
- [Examples](#examples)
- [Glue options](#glue-options)
- [Tags](#tags)
- [Type Conversion](#type-conversion)
- [Performance](#performance)
- [Possible Improvements](#possible-improvements)
- [License](#license)

## Introduction
`Glue` is a simple tool that performs **shallow** copy between destination struct and source struct, from src field to dst field, the limitations are:
- Both fields are exported.
- Both fields have the same name or source have the field with name tagged on dst field\*.
- Both fields have (strictly) same type or have registered conversion function.

\*see tags for alias.

`Glue` is thread-safe although it uses some globally available cache internally.

One thing to be noted that `Glue` can pull fields from embedded struct, you may think `Glue` worked like this:
```go
type Emb struct {
    E int
}
type Foo struct {
    A int
    E int
}
type Bar struct {
    A int
    Emb
}

f := &Foo{}
b := &Bar{
    Emb: Emb{E: 1024},
    A:   4096,
}

err := glue.Glue(f, b)
// glue does something like this for you:
// f.A = b.A
// f.E = b.E (or f.E = b.Emb.E)
```

You may consider using `Glue` or other similar solution (like jinzhu/copier) when you have two near-identical struct that you cannot/have hard time change/changing the definition (legacy library or auto generated, ex: Protobuf), `Glue` can handle the filling process programmatically.

Or you can just hard-code the filling process ;), but if this kind of code happens all the time across the project, `Glue` and similar solutions are always available options.

`Glue` only accepts pointer to struct as parameter, if not it returns an `ErrNotPtrToStruct` error.

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

## Glue options
`Glue` have options as variadic parameter the currently available options are:
- `DoStrict`
  If a field does not have its counterpart, `Glue` returns `ErrUnsatisfiedField` error. The default mode is relaxed, `Glue` doens't complain if fields are not found.
- `DoFavorSource`
  `Glue` turns to "push" fields from source -- in other words, it is the source seeking counterpart in the destination.
  The default mode is favor destination, meaning the destination "pulls" fields from source.

Here is an example of using the `DoStrict` option:
```go
type Foo struct {
    A int
}
type Bar struct {
    B int
}
a := &Foo{}
b := &Bar{B: 1023}

opts := []glue.GlueOption{
    glue.DoStrict(),
}
err := glue.Glue(a, b, opts...)
fmt.Println(errors.Is(err, glue.ErrUnsatisfiedField))
// should be true

```

## Tags
A field can be tagged with a valid identifier as an alias, the effects are:
- Using favor destination(the default), it behaves as the field is the name we tags.
- Using favor source(use option `DoFavorSource`), the field being tagged exports as the name we tags.

In short, `Glue` sees the tag as the name of the field.

Unexported fields are always ignored even with tags, they cannot be set using reflect library by the way.

Currently `Glue` only accept one tag attribute, also the tag takes no effect when a field is in the source struct.

`Glue` panics if tag attribute is not `-`(ignore) or a valid golang identifier.

## Type conversion
You can register a global conversion function using `RegConv`, the function must have signature that takes type of source field and outputs a value that have the same type as the destination field.
`RegConv` fails if the `converter` passed in is not a function or a function having incompatible signature.

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

## Performance
Reflection stuffs are usually not quiet fast, especially involving embedded/anonymous fields, especially searching fields in embedded structure, it slows down the process significantly.
Searching fields inside embedded structure is about 4 time slower on my computer compare to accessing plain and expored fields, and cost almost 16 times more memory then the plain version during benchmarking.

Althought it is a slow process, it can be benefit from parallelized processing.

## Possible Improvements
- [ ] Optionally performs deep copy on reference types(slice, map, pointer to object).
- [ ] Get value from simple method that takes no parameter.

## License
This library is distributed under MIT license.
