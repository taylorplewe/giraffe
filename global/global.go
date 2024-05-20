package global

import (
	"giraffe/types"
	"syscall/js"
)

var (
	Center   types.Point
	CanvasEl js.Value
	Ctx      js.Value
)
