package glue_test

import (
	"glue"
	"math/rand"
	"runtime"
	"sync"
	"testing"
)

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
