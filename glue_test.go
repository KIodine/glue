package glue_test

import (
	"glue"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlueBasic(t *testing.T) {
	type xFoo struct {
		Beta  int
		Alpha string
	}

	type xBar struct {
		Alpha string
		Beta  int
	}
	var err error

	foo := &xFoo{
		Alpha: "Alfa",
		Beta:  0,
	}
	bar := &xBar{
		Alpha: "b",
		Beta:  128,
	}
	err = glue.Glue(bar, foo)

	assert.Nil(t, err)
	assert.Equal(t, foo.Alpha, foo.Alpha)
}

func TestGlueTagBasic(t *testing.T) {
	type xBaz struct {
		A int `glue:"B"`
	}

	type xPaz struct {
		B int
	}
	var err error

	baz := &xBaz{A: -1}
	paz := &xPaz{B: 1997}
	err = glue.Glue(baz, paz)

	assert.Nil(t, err)
	assert.Equal(t, paz.B, baz.A)
	if baz.A != paz.B {
		t.Error("NE!")
	}
}

type bFoo struct {
	A int64
	B string
	C []byte
	D map[string]bool
	E chan int
}

type bBar struct {
	A int64
	B string
	C []byte
	D map[string]bool
	E chan int
	F uint32
}

func TestGlue(t *testing.T) {
	bb := &bBar{
		A: rand.Int63(),
		B: "bBarstring" + strconv.QuoteRune('A'+rand.Int31n(26)),
		C: []byte{'a', 'b', 'c', 'd'},
		D: make(map[string]bool),
		E: make(chan int, 24),
		F: rand.Uint32(),
	}
	bf := &bFoo{}
	glue.Glue(bf, bb)
	assert.Equal(t, bb.A, bf.A)
	assert.Equal(t, bb.B, bf.B)
	assert.Equal(t, bb.C, bf.C)
	assert.Equal(t, bb.D, bf.D)
	assert.Equal(t, bb.E, bf.E)
}

/* TODO: test embedded fields. */
func TestGlueEmbedded(t *testing.T) {
	type (
		EFoo struct {
			A int
		}
		eBar struct {
			EFoo
		}
		eBaz struct {
			EFoo
			B int
		}
	)
	var (
		ans_a int = 9977
	)
	ef := &EFoo{A: ans_a}
	eb := &eBar{}
	ez := &eBaz{EFoo: *ef, B: -1000}
	glue.Glue(eb, ez)
	/* NOTE: The embedded field is actually a field have the name as same as its
	type, field mixing is just a syntax sugar. */
	assert.Equal(t, ans_a, eb.EFoo.A)
}

/* --- benchmarks --- */
func BenchmarkGlueBasic(b *testing.B) {
	var ff = new(bFoo)
	bz := make([]*bBar, 128)
	for i := 0; i < len(bz); i++ {
		bz[i] = &bBar{
			A: rand.Int63(),
			B: "bBarstring" + strconv.QuoteRune('A'+rand.Int31n(26)),
			C: []byte{'a', 'b', 'c', 'd'},
			D: make(map[string]bool),
			E: make(chan int, 24),
			F: rand.Uint32(),
		}
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bf := bz[rand.Intn(len(bz))]
		glue.Glue(ff, bf)
	}
}

func BenchmarkGlueWithTag(b *testing.B) {
	type tFoo struct {
		A int64           `glue:"M"`
		B string          `glue:"N"`
		C []byte          `glue:"O"`
		D map[string]bool `glue:"P"`
		E chan int        `glue:"Q"`
	}
	type tBar struct {
		M int64
		N string
		O []byte
		P map[string]bool
		Q chan int
		r uint32
	}
	var ff = new(tFoo)
	bz := make([]*tBar, 128)
	for i := 0; i < len(bz); i++ {
		bz[i] = &tBar{
			M: rand.Int63(),
			N: "bBarstring" + strconv.QuoteRune('A'+rand.Int31n(26)),
			O: []byte{'a', 'b', 'c', 'd'},
			P: make(map[string]bool),
			Q: make(chan int, 24),
			r: rand.Uint32(),
		}
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bf := bz[rand.Intn(len(bz))]
		glue.Glue(ff, bf)
	}
}

/* TODO: test failing conditions:
- [X] Not valid parameter type.
	- [X] Not pointer.
	- [X] `nil` pointer.
	- [X] Different type.
- [X] Unexported destination.
- [X] Unmatched destination.
	- [X] Same name, different type.
- [ ] ?Fields cannot set.(How?)
*/

func TestInvalidParams(t *testing.T) {
	var (
		err error
		a   int              = 0
		b   string           = "1024"
		c   *struct{ A int } = nil
		d   *int             = nil
	)

	err = glue.Glue(nil, nil)
	assert.Equal(t, glue.ErrTypeIncompat, err)
	err = glue.Glue(c, d)
	assert.Equal(t, glue.ErrTypeIncompat, err)
	err = glue.Glue(1, 2)
	assert.Equal(t, glue.ErrTypeIncompat, err)
	err = glue.Glue(&a, &b)
	assert.Equal(t, glue.ErrTypeIncompat, err)
}

func TestUnexportedFields(t *testing.T) {
	// Test a field is not touched even it is tagged.
	type uFoo struct {
		a int `glue:"A"`
	}
	type uBar struct {
		A int
	}
	var a int = 1337

	uf := &uFoo{a: a}
	ub := &uBar{A: 16384}
	glue.Glue(uf, ub)
	assert.Equal(t, a, uf.a)
}

func TestUnmatchedFields(t *testing.T) {
	type (
		uFoo struct {
			A string `glue:"A"`
			B int    `glue:"B"`
		}
		uBar struct {
			A int
		}
	)
	var (
		ans_str      string = "example string"
		ans_a, ans_b        = -1, 9977
	)

	uf := &uFoo{A: ans_str, B: ans_a}
	ub := &uBar{A: ans_b}
	glue.Glue(uf, ub)

	assert.Equal(t, ans_str, uf.A)
	assert.Equal(t, ans_a, uf.B)
	assert.Equal(t, ans_b, ub.A)

}
