package glue_test

import (
	"glue"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvBasic(t *testing.T) {
	Int2F64 := func(n int) float64 {
		return float64(n)
	}
	Int2F32 := func(n int) float32 {
		return float32(n)
	}
	Int2U16 := func(n int) uint16 {
		return uint16(n)
	}
	type cFoo struct {
		A float64
		B float32
		C uint16
	}
	type cBar struct {
		A int
		B int
		C int
	}
	cf := &cFoo{
		A: -1.0,
		B: -1.0,
		C: 8192,
	}
	cb := &cBar{
		A: 0,
		B: 0,
		C: 1024,
	}
	var ok bool
	ok = glue.RegConversion(float64(0), int(0), Int2F64)
	assert.True(t, ok)
	ok = glue.RegConversion(float32(0), int(0), Int2F32)
	assert.True(t, ok)
	ok = glue.RegConversion(uint16(0), int(0), Int2U16)
	assert.True(t, ok)

	err := glue.Glue(cf, cb)

	assert.NoError(t, err)
	assert.Equal(t, float64(0.0), cf.A)
	assert.Equal(t, float32(0.0), cf.B)
	assert.Equal(t, uint16(1024), cf.C)

	// reset, randomize.
	cf.A = rand.Float64()
	cf.B = rand.Float32()
	cf.C = uint16(rand.Uint32())

}

func BenchmarkConv(b *testing.B) {
	Int2F64 := func(n int) float64 {
		return float64(n)
	}
	Int2F32 := func(n int) float32 {
		return float32(n)
	}
	Int2U16 := func(n int) uint16 {
		return uint16(n)
	}
	type cFoo struct {
		A float64
		B float32
		C uint16
	}
	type cBar struct {
		A int
		B int
		C int
	}
	cf := &cFoo{
		A: -1.0,
		B: -1.0,
		C: 8192,
	}
	cb := &cBar{
		A: 0,
		B: 0,
		C: 1024,
	}
	var ok bool
	ok = glue.RegConversion(float64(0), int(0), Int2F64)
	assert.True(b, ok)
	ok = glue.RegConversion(float32(0), int(0), Int2F32)
	assert.True(b, ok)
	ok = glue.RegConversion(uint16(0), int(0), Int2U16)
	assert.True(b, ok)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = glue.Glue(cf, cb)
		/*
			assert.NoError(b, err)
			assert.Equal(b, float64(0.0), cf.A)
			assert.Equal(b, float32(0.0), cf.B)
			assert.Equal(b, uint16(1024), cf.C)
		*/
		// reset, randomize.
		cf.A = rand.Float64()
		cf.B = rand.Float32()
		cf.C = uint16(rand.Uint32())
	}
}
