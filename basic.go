package main

import (
	"basicallygo/context"
)

type t struct {
}

func (t t) Printline(s string) {
	println(s)
}

func main() {
	prog :=
		`30 PRINT 2 + ABS(-3)`

	term := t{}

	cont := context.New(term)

	cont.Set_input_buffer(prog)

	cont.Run()
}
