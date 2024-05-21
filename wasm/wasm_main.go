package main

import (
	"basicallygo/context"
	"basicallygo/terminal"
	"strings"
	"syscall/js"
)

func basic_run(cont *context.Context, int chan bool) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]

			go func() {
				run_done := make(chan bool)
				go cont.Run(int, run_done)
				<-run_done
				resolve.Invoke()
			}()

			return nil
		})

		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

func main() {
	done := make(chan struct{}, 0)
	global := js.Global()
	term := &terminal.Terminal{}
	cont := context.New(term)
	interrupt := make(chan bool, 1)

	global.Set("basic_run", basic_run(cont, interrupt))

	global.Set("basic_set_term_printline", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		function := p[0]
		term.Printline = func(s string) {
			function.Invoke(js.ValueOf(s))
		}
		return js.ValueOf(true)
	}))

	global.Set("basic_accept_line", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		ok := cont.Accept_line(p[0].String())
		return js.ValueOf(ok)
	}))

	global.Set("basic_interrupt", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		interrupt <- true
		return nil
	}))

	global.Set("basic_list", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		cont.List()
		return nil
	}))

	global.Set("basic_set_get_keys_down", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		function := p[0]
		cont.Get_keys_down = func() []string {
			keys := function.Invoke().String()
			// Split into slice
			s := strings.Split(keys, ",")
			return s
		}
		return js.ValueOf(true)
	}))
	<-done
}
