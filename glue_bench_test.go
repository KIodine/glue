package glue_test

import (
	"glue"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

func BenchmarkDive(b *testing.B) {
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
	u := &Uniform{
		Embed: Embed{
			A: 15,
			B: 31,
			C: 63,
			D: 127,
			E: 255,
			F: 511,
		},
	}
	v := &Victor{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		glue.Glue(v, u)
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

func BenchmarkGlueWithTagParallel(b *testing.B) {
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
	const nJob = 4
	var bar sync.WaitGroup

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

	bar.Add(nJob)
	for j := 0; j < nJob; j++ {
		go func() {
			defer bar.Done()
			ff := new(tFoo)
			for i := 0; i < (b.N / nJob); i++ {
				bf := bz[rand.Intn(len(bz))]
				glue.Glue(ff, bf)
			}
		}()
	}
	bar.Wait()
}

func BenchmarkBigStruct(b *testing.B) {
	type (
		Foo struct {
			Alfa     int
			Bravo    int
			Charlie  int
			Delta    int
			Echo     int
			Foxtrot  int
			Golf     int
			Hotel    int
			India    int
			Juliett  int
			Kilo     int
			Lima     int
			Mike     int
			November int
			Oscar    int
			Papa     int
			Quebec   int
			Romeo    int
			Sierra   int
			Tango    int
			Uniform  int
			Victor   int
			Whiskey  int
			Xray     int
			Yankee   int
			Zulu     int
		}
		Bar struct {
			A int `glue:"Alfa"`
			B int `glue:"Bravo"`
			C int `glue:"Charlie"`
			D int `glue:"Delta"`
			E int `glue:"Echo"`
			F int `glue:"Foxtrot"`
			G int `glue:"Golf"`
			H int `glue:"Hotel"`
			I int `glue:"india"`
			J int `glue:"Juliett"`
			K int `glue:"Kilo"`
			L int `glue:"Lima"`
			M int `glue:"Mike"`
			N int `glue:"November"`
			O int `glue:"Oscar"`
			P int `glue:"Papa"`
			Q int `glue:"Quebec"`
			R int `glue:"Romeo"`
			S int `glue:"Sierra"`
			T int `glue:"Tango"`
			U int `glue:"Uniform"`
			V int `glue:"Victor"`
			W int `glue:"Whiskey"`
			X int `glue:"Xray"`
			Y int `glue:"Yankee"`
			Z int `glue:"Zulu"`
		}
	)
	fs := make([]*Foo, 128)
	for i := 0; i < len(fs); i++ {
		fs[i] = &Foo{
			Alfa: rand.Int(), Bravo: rand.Int(), Charlie: rand.Int(),
			Delta: rand.Int(), Echo: rand.Int(), Foxtrot: rand.Int(),
			Golf: rand.Int(), Hotel: rand.Int(), India: rand.Int(),
			Juliett: rand.Int(), Kilo: rand.Int(), Lima: rand.Int(),
			Mike: rand.Int(), November: rand.Int(), Oscar: rand.Int(),
			Papa: rand.Int(), Quebec: rand.Int(), Romeo: rand.Int(),
			Sierra: rand.Int(), Tango: rand.Int(), Uniform: rand.Int(),
			Victor: rand.Int(), Whiskey: rand.Int(), Xray: rand.Int(),
			Yankee: rand.Int(), Zulu: rand.Int(),
		}
	}
	b.ResetTimer()

	br := &Bar{}
	for i := 0; i < b.N; i++ {
		f := fs[rand.Intn(len(fs))]
		glue.Glue(br, f)
	}
}

func BenchmarkBigStructParallel(b *testing.B) {
	type (
		Foo struct {
			Alfa     int
			Bravo    int
			Charlie  int
			Delta    int
			Echo     int
			Foxtrot  int
			Golf     int
			Hotel    int
			India    int
			Juliett  int
			Kilo     int
			Lima     int
			Mike     int
			November int
			Oscar    int
			Papa     int
			Quebec   int
			Romeo    int
			Sierra   int
			Tango    int
			Uniform  int
			Victor   int
			Whiskey  int
			Xray     int
			Yankee   int
			Zulu     int
		}
		Bar struct {
			A int `glue:"Alfa"`
			B int `glue:"Bravo"`
			C int `glue:"Charlie"`
			D int `glue:"Delta"`
			E int `glue:"Echo"`
			F int `glue:"Foxtrot"`
			G int `glue:"Golf"`
			H int `glue:"Hotel"`
			I int `glue:"india"`
			J int `glue:"Juliett"`
			K int `glue:"Kilo"`
			L int `glue:"Lima"`
			M int `glue:"Mike"`
			N int `glue:"November"`
			O int `glue:"Oscar"`
			P int `glue:"Papa"`
			Q int `glue:"Quebec"`
			R int `glue:"Romeo"`
			S int `glue:"Sierra"`
			T int `glue:"Tango"`
			U int `glue:"Uniform"`
			V int `glue:"Victor"`
			W int `glue:"Whiskey"`
			X int `glue:"Xray"`
			Y int `glue:"Yankee"`
			Z int `glue:"Zulu"`
		}
	)
	nJob := runtime.NumCPU()
	fs := make([]*Foo, 128)
	for i := 0; i < len(fs); i++ {
		fs[i] = &Foo{
			Alfa: rand.Int(), Bravo: rand.Int(), Charlie: rand.Int(),
			Delta: rand.Int(), Echo: rand.Int(), Foxtrot: rand.Int(),
			Golf: rand.Int(), Hotel: rand.Int(), India: rand.Int(),
			Juliett: rand.Int(), Kilo: rand.Int(), Lima: rand.Int(),
			Mike: rand.Int(), November: rand.Int(), Oscar: rand.Int(),
			Papa: rand.Int(), Quebec: rand.Int(), Romeo: rand.Int(),
			Sierra: rand.Int(), Tango: rand.Int(), Uniform: rand.Int(),
			Victor: rand.Int(), Whiskey: rand.Int(), Xray: rand.Int(),
			Yankee: rand.Int(), Zulu: rand.Int(),
		}
	}
	bar := new(sync.WaitGroup)
	bar.Add(nJob)
	b.ResetTimer()

	for j := 0; j < nJob; j++ {
		go func() {
			defer bar.Done()
			br := &Bar{}
			for i := 0; i < (b.N / nJob); i++ {
				f := fs[rand.Intn(len(fs))]
				glue.Glue(br, f)
			}
		}()
	}
	bar.Wait()
}

func BenchmarkIncompatStruct(b *testing.B) {
	type Foo struct {
		A int
		B string
		C map[int]string
		D []int
		E float64
	}
	type Bar struct {
		A string
		B []int
		C float64
		D int
		E map[int]string
	}
	u := &Foo{}
	v := &Bar{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		glue.Glue(u, v)
	}
}

func BenchmarkTimeConversion(b *testing.B) {
	type Foo struct {
		A time.Time
	}
	type Bar struct {
		A int64
	}
	unixToTime := func(t int64) time.Time {
		return time.Unix(t, 0)
	}
	err := glue.RegConv(time.Time{}, int64(0), unixToTime)
	assert.NoError(b, err)

	now := time.Now().UTC()

	u := &Foo{}
	v := &Bar{
		A: now.Unix(),
	}

	for i := 0; i < b.N; i++ {
		glue.Glue(u, v)
	}
}
