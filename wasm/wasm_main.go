package main

import (
	"basicallygo/context"
	"basicallygo/terminal"
	"fmt"
	"syscall/js"
)

func main() {
	done := make(chan struct{}, 0)
	global := js.Global()
	term := &terminal.Terminal{}
	cont := context.New(term)
	interrupt := make(chan bool, 1)

	global.Set("basic_set_term_printline", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		function := p[0]
		term.Printline = func(s string) {
			println("invoking", function.String())
			function.Invoke(js.ValueOf(s))
		}
		return nil
	}))

	global.Set("basic_accept_line", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		cont.Accept_line(p[0].String())
		return nil
	}))

	global.Set("basic_run", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		run_done := make(chan bool)
		go cont.Run(interrupt, run_done)
		return nil
	}))

	global.Set("basic_interrupt", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		interrupt <- true
		return nil
	}))

	global.Set("basic_list", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		println("In list")
		fmt.Printf("%+v\n", term)
		cont.List()
		return nil
	}))

	<-done
}
