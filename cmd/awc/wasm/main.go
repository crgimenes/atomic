package main

import (
	"fmt"
	"syscall/js"
)

func writeToScreen(this js.Value, args []js.Value) any {
	// print args in the console
	for i, arg := range args {
		fmt.Println("GO arg", i, "is", arg)
	}
	ret := js.ValueOf("ret from GO Hello, world!")
	return ret
}

func main() {
	fmt.Println("init from wasm")
	js.Global().Set("writeToScreen", js.FuncOf(writeToScreen))

	// call JS function from Go
	js.Global().Call("sendToServer", "JS sendToServer called from GO")

	c := make(chan struct{}, 0)
	<-c
}
