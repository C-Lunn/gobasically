package main

import "syscall/js"

func main() {
	done := make(chan struct{}, 0)
	global := js.Global()
	global.Set("echo", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		println(p[0].String())
		return nil
	}))

	<-done
}
