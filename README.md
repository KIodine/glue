# Glue: merging fields shared between two structs
## Introduction
`Glue` is a simple tool that merges fields that:
- Is exported.
- Have the same name or source is tagged(see below).
- Have (strictly) same type

between source and destination structs.

`Glue` also accepts a tag in destination struct that specifies the name of source field, though it will be slightly slower.

## Example
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

## License
This library is distributed under MIT license
