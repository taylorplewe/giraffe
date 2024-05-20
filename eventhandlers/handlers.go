package eventhandlers

import (
	"syscall/js"

	"giraffe/global"
	"giraffe/types"
)

var (
	dragging    = false
	startPoint  types.Point
	startCenter types.Point
)

func EventHandler(callback func(js.Value)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		e := args[0]
		e.Call("preventDefault")
		callback(e)
		return nil
	})
}

func getMousePos(e js.Value) types.Point {
	rect := global.CanvasEl.Call("getBoundingClientRect")
	x := e.Get("clientX").Int() - rect.Get("offsetLeft").Int()
	y := e.Get("clientY").Int() - rect.Get("offsetTop").Int()
	return types.Point{X: x, Y: y}
}
func OnMouseDown(e js.Value) {
	dragging = true
	startPoint = getMousePos(e)
	startCenter = global.Center
}
func OnMouseMove(e js.Value) {
	if !dragging {
		return
	}
	newPoint := getMousePos(e)
	diff := newPoint.Sub(startPoint)
	global.Center = startCenter.Add(diff)
}
func OnMouseUp(e js.Value) {
	dragging = false
}

func OnWheel(e js.Value) {
	delta := e.Get("deltaY").Float()
	global.Ctx.Call("scale", 1+delta/1000, 1+delta/1000)
}
