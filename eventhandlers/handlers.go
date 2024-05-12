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

func OnMouseDown(this js.Value, args []js.Value) any {
	e := args[0]
	dragging = true
	startPoint = types.Point{
		X: e.Get("clientX").Int(),
		Y: e.Get("clientY").Int(),
	}
	startCenter = global.Center
	return nil
}
func OnMouseMove(this js.Value, args []js.Value) any {
	e := args[0]
	if !dragging {
		return nil
	}
	newPoint := types.Point{
		X: e.Get("clientX").Int(),
		Y: e.Get("clientY").Int(),
	}
	diff := newPoint.Sub(startPoint)
	global.Center = startCenter.Add(diff)
	return nil
}
func OnMouseUp(this js.Value, args []js.Value) any {
	dragging = false
	return nil
}
