package glue_test

import (
	"glue"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
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
	var err error

	err = glue.RegConv(float64(0), int(0), Int2F64)
	assert.NoError(t, err)
	err = glue.RegConv(float32(0), int(0), Int2F32)
	assert.NoError(t, err)
	err = glue.RegConv(uint16(0), int(0), Int2U16)
	assert.NoError(t, err)

	err = glue.Glue(cf, cb)

	assert.NoError(t, err)
	assert.Equal(t, float64(0.0), cf.A)
	assert.Equal(t, float32(0.0), cf.B)
	assert.Equal(t, uint16(1024), cf.C)
}

func TestRegNonFunction(t *testing.T) {
	err := glue.RegConv(int(0), int(0), float64(0))
	assert.ErrorIs(t, err, glue.ErrNotFunction)
}

func TestRegIncompatSignature(t *testing.T) {
	Int2F32 := func(n int) float32 { return float32(n) }
	err := glue.RegConv(float32(0), int64(0), Int2F32)
	assert.ErrorIs(t, err, glue.ErrIncompatSignature)
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
	var err error

	err = glue.RegConv(float64(0), int(0), Int2F64)
	assert.NoError(b, err)
	err = glue.RegConv(float32(0), int(0), Int2F32)
	assert.NoError(b, err)
	err = glue.RegConv(uint16(0), int(0), Int2U16)
	assert.NoError(b, err)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = glue.Glue(cf, cb)

		cf.A = rand.Float64()
		cf.B = rand.Float32()
		cf.C = uint16(rand.Uint32())
	}
}

func BenchmarkConvParallel(b *testing.B) {
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

	cb := &cBar{
		A: 0,
		B: 0,
		C: 1024,
	}
	var err error

	err = glue.RegConv(float64(0), int(0), Int2F64)
	assert.NoError(b, err)
	err = glue.RegConv(float32(0), int(0), Int2F32)
	assert.NoError(b, err)
	err = glue.RegConv(uint16(0), int(0), Int2U16)
	assert.NoError(b, err)
	b.ResetTimer()

	var nJob = runtime.NumCPU()
	var wg sync.WaitGroup

	wg.Add(nJob)
	for j := 0; j < nJob; j++ {
		go func() {
			defer wg.Done()
			cf := &cFoo{
				A: -1.0,
				B: -1.0,
				C: 8192,
			}
			for i := 0; i < (b.N / nJob); i++ {
				_ = glue.Glue(cf, cb)

				cf.A = rand.Float64()
				cf.B = rand.Float32()
				cf.C = uint16(rand.Uint32())
			}
		}()
	}
	wg.Wait()
}

func TestMustRegConv(t *testing.T) {
	IntToDouble := func(n int) float64 {
		return float64(n)
	}
	assert.Panics(t, func() {
		_ = glue.MustRegConv(uint64(0), float64(0), IntToDouble)
	})
	assert.NotPanics(t, func() {
		_ = glue.MustRegConv(float64(0), int(0), IntToDouble)
	})
}

func TestDeregConv(t *testing.T) {
	type Foo struct {
		A int
	}
	type Bar struct {
		A string
	}
	strToInt := func(s string) int {
		n, _ := strconv.Atoi(s)
		return n
	}
	err := glue.RegConv(int(0), "", strToInt)
	assert.NoError(t, err)

	opts := []glue.GlueOption{
		glue.DoStrict(),
	}

	a := &Foo{A: -1}
	b := &Bar{A: "1023"}

	err = glue.Glue(a, b, opts...)
	assert.NoError(t, err)
	assert.Equal(t, 1023, a.A)

	glue.DeregConv(int(0), "")
	a.A = -1

	err = glue.Glue(a, b, opts...)
	assert.ErrorIs(t, err, glue.ErrUnsatisfiedField)
	assert.Equal(t, -1, a.A)

}
