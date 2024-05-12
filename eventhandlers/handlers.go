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

func OnMouseDown(e js.Value) {
	dragging = true
	startPoint = types.Point{
		X: e.Get("clientX").Int(),
		Y: e.Get("clientY").Int(),
	}
	startCenter = global.Center
}
func OnMouseMove(e js.Value) {
	if !dragging {
		return
	}
	newPoint := types.Point{
		X: e.Get("clientX").Int(),
		Y: e.Get("clientY").Int(),
	}
	diff := newPoint.Sub(startPoint)
	global.Center = startCenter.Add(diff)
}
func OnMouseUp(e js.Value) {
	dragging = false
}
