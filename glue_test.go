package glue_test

import (
	"glue"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type xFoo struct {
	Beta  int
	Alpha string
}

type xBar struct {
	Alpha string
	Beta  int
}

func TestGlueBasic(t *testing.T) {
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
	if err != nil {
		t.Error(err)
	}
	if bar.Alpha != foo.Alpha {
		t.Error("NE!", bar, foo)
	}
	t.Log(bar, foo)
}

type xBaz struct {
	A int `glue:"B"`
}

type xPaz struct {
	B int
}

func TestGlueTagBasic(t *testing.T) {
	var err error
	baz := &xBaz{A: -1}
	paz := &xPaz{B: 1997}
	err = glue.Glue(baz, paz)
	if err != nil {
		t.Error(err)
	}
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
	assert.Equal(t, bf.A, bb.A)
	assert.Equal(t, bf.B, bb.B)
	assert.Equal(t, bf.C, bb.C)
	assert.Equal(t, bf.D, bb.D)
	assert.Equal(t, bf.E, bb.E)
}

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

/* TODO: benchmark gluing using tag. */
