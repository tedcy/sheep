package common

import (
	"testing"
	"fmt"
)

func Test_AddGet(t *testing.T) {
	a := NewAdder()
	b := "1"
	c := "" + "1"
	a.Add(b,1)
	a.Add(b,1)
	a.Add(b,1)
	a.Add(b,1)
	a.Add(b,1)
	a.Add(b,1)
	a.Add(b,1)
	a.Add("2",1)
	println(a.Get(c))
	fmt.Println(a.List())
	a.Clean()
	println(a.Get(c))
	fmt.Println(a.List())
}
