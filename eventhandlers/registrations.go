package eventhandlers

import (
	"syscall/js"
)

var handlers = map[string]func(js.Value){
	"mousedown": OnMouseDown,
	"mousemove": OnMouseMove,
	"mouseup":   OnMouseUp,
}

func RegisterAll(canvasEl js.Value) {
	for event, handler := range handlers {
		canvasEl.Call("addEventListener", event, EventHandler(handler))
	}
}
