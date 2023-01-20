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

	// pega o canvas no html
	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	// pega o contexto 2d do canvas
	ctx := canvas.Call("getContext", "2d")

	// ajusta o tamanho da imagem dentro do canvas para 800x600
	//canvas.Set("width", 800)
	//canvas.Set("height", 600)

	// desenha um retangulo vermelho
	ctx.Set("fillStyle", "red")
	ctx.Call("fillRect", 10, 10, 100, 100)

	// desenha pixel branco na posicao 10, 10 do canvas
	ctx.Set("fillStyle", "white")
	ctx.Call("fillRect", 10, 10, 1, 1)

	c := make(chan struct{}, 0)
	<-c
}

/*

ws := js.Global().Get("WebSocket").New("ws://localhost:8080/ws")

ws.Call("addEventListener", "open", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    fmt.Println("open")

    ws.Call("send", js.TypedArrayOf([]byte{123}))
    return nil
}))
*/
