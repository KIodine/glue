package glue_test

import (
	"glue"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionStrict(t *testing.T) {
	type Alpha struct {
		A int
	}
	type Beta struct {
		B int
	}
	a := &Alpha{}
	b := &Beta{B: 1023}

	opts := []glue.GlueOption{
		glue.DoStrict(),
	}
	err := glue.Glue(a, b, opts...)
	assert.ErrorIs(t, err, glue.ErrGlue)
}

func TestOptionFavorSource(t *testing.T) {
	type Alpha struct {
		A int
	}
	type Beta struct {
		A int
	}
	a := &Alpha{A: -1}
	b := &Beta{A: 1023}

	opts := []glue.GlueOption{
		glue.DoFavorSource(),
	}
	err := glue.Glue(a, b, opts...)
	assert.NoError(t, err, glue.ErrGlue)
	assert.Equal(t, 1023, a.A)
}

func TestOptionFavorSrcTag(t *testing.T) {
	type Alpha struct {
		A int
	}
	type Beta struct {
		B int `glue:"A"`
	}
	a := &Alpha{A: -1}
	b := &Beta{B: 1023}

	opts := []glue.GlueOption{
		glue.DoFavorSource(),
	}
	err := glue.Glue(a, b, opts...)
	assert.NoError(t, err)
	assert.Equal(t, 1023, a.A)
}

func TestStructFavorSrc(t *testing.T) {
	type Alpha struct {
		A int
		B int
		C int
	}
	type Beta struct {
		A int
		B int
	}
	type Charlie struct {
		C int
		D int
	}
	a := &Alpha{A: -1}
	b := &Beta{A: 1023, B: 511}
	c := &Charlie{C: 2047, D: 4095}

	opts := []glue.GlueOption{
		glue.DoStrict(),
		glue.DoFavorSource(),
	}
	err := glue.Glue(a, b, opts...)
	assert.NoError(t, err)
	assert.Equal(t, 1023, a.A)
	assert.Equal(t, 511, a.B)

	err = glue.Glue(a, c, opts...)
	assert.ErrorIs(t, err, glue.ErrUnsatisfiedField)
}

func BenchmarkFavorSource(b *testing.B) {
	type Uniform struct {
		A int
		B int
		C int
		D int
		E int
		F int
	}
	type Victor struct {
		A int
		B int
		C int
		D int
		E int
		F int
	}
	u := &Uniform{}
	v := &Victor{
		A: 15,
		B: 31,
		C: 63,
		D: 127,
		E: 255,
		F: 511,
	}
	opts := []glue.GlueOption{
		glue.DoFavorSource(),
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		glue.Glue(u, v, opts...)
	}

}

func BenchmarkFavorSourceDive(b *testing.B) {
	type Embed struct {
		A int
		B int
		C int
		D int
		E int
		F int
	}
	type Uniform struct {
		Embed
	}
	type Victor struct {
		A int
		B int
		C int
		D int
		E int
		F int
	}
	u := &Uniform{}
	v := &Victor{
		A: 15,
		B: 31,
		C: 63,
		D: 127,
		E: 255,
		F: 511,
	}
	opts := []glue.GlueOption{
		glue.DoFavorSource(),
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		glue.Glue(u, v, opts...)
	}
}
