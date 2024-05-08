package main

import (
	"syscall/js"
)

type Rect struct {
	x      int
	y      int
	width  int
	height int
}

type Card struct {
	rect  Rect
	title string
}

var (
	width    float64
	height   float64
	document js.Value
	canvasEl js.Value
	ctx      js.Value
	draw     js.Func
	exCard   *Card
)

func main() {
	done := make(chan struct{}, 0)
	document = js.Global().Get("document")
	canvasEl = document.Call("querySelector", "canvas")
	ctx = canvasEl.Call("getContext", "2d")
	exCard = &Card{
		Rect{
			32,
			32,
			100,
			80,
		},
		"example",
	}

	draw = js.FuncOf(func(this js.Value, args []js.Value) any {
		// Pull window size to handle resize
		curBodyW := document.Get("body").Get("clientWidth").Float()
		curBodyH := document.Get("body").Get("clientHeight").Float()
		if curBodyW != width || curBodyH != height {
			width, height = curBodyW, curBodyH
			canvasEl.Set("width", width)
			canvasEl.Set("height", height)
		}
		ctx.Call("clearRect", 0, 0, width, height)

		drawCard(exCard)

		js.Global().Call("requestAnimationFrame", draw)
		return nil
	})
	defer draw.Release()

	js.Global().Call("requestAnimationFrame", draw)
	<-done
}

func drawCard(card *Card) {
	ctx.Set("fillStyle", "rgb(200 0 0)")
	ctx.Call("fillRect", card.rect.x, card.rect.y, card.rect.width, card.rect.height)
}
