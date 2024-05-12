package eventhandlers

import (
	"syscall/js"
)

func RegisterAll(canvasEl js.Value) {
	canvasEl.Call("addEventListener", "mousedown", js.FuncOf(OnMouseDown))
	canvasEl.Call("addEventListener", "mousemove", js.FuncOf(OnMouseMove))
	canvasEl.Call("addEventListener", "mouseup", js.FuncOf(OnMouseUp))
}
