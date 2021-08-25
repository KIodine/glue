package glue_test

import (
	"glue"
	"testing"
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
