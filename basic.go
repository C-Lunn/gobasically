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
		`20 PRINT "world!"
10 PRINT "Hello,"
30 PRINT 3+3+4`

	term := t{}

	cont := context.New(term)

	cont.Set_input_buffer(prog)

	cont.Run()
}
