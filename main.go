package main

import (
	"encoding/csv"
	"fmt"
	"strings"
	"syscall/js"

	"giraffe/eventhandlers"
	"giraffe/global"
	"giraffe/types"
)

var (
	width    float64
	height   float64
	document js.Value
	canvasEl js.Value
	ctx      js.Value
	draw     js.Func
	users    []*types.User
)

func main() {
	forever := make(chan struct{})

	document = js.Global().Get("document")
	canvasEl = document.Call("querySelector", "canvas")
	ctx = canvasEl.Call("getContext", "2d")

	getFile()

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

		drawAllCards()

		js.Global().Call("requestAnimationFrame", draw)
		return nil
	})
	defer draw.Release()

	global.Center = types.Point{0, 0}
	eventhandlers.RegisterAll(canvasEl)
	js.Global().Call("requestAnimationFrame", draw)
	<-forever
}

func drawCard(point types.Point, user *types.User) {
	padding := 16
	title := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	width := ctx.Call("measureText", title).Get("width").Int()

	// border
	ctx.Set("strokeStyle", "white")
	ctx.Call("beginPath")
	ctx.Call("roundRect", point.X-padding, point.Y-padding, width+(padding*2), 80+(padding*2), padding)
	ctx.Call("stroke")

	// title
	ctx.Set("font", "24px Roboto")
	ctx.Set("fillStyle", "white")
	ctx.Set("textBaseline", "top")
	ctx.Call("fillText", title, point.X, point.Y)
}

func drawAllCards() {
	y := global.Center.Y + (16 + 4)
	x := global.Center.X + (16 + 4)
	for _, user := range users {
		drawCard(types.Point{
			X: x,
			Y: y,
		}, user)
		y += (80 + 32 + 8)
	}

}

func readCsvAsUserList(data string) []*types.User {
	_users := []*types.User{}

	csvReader := csv.NewReader(strings.NewReader(data))
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Errorf("Error reading csv: %v", err)
		return _users
	}

	columnMap := map[string]int{}
	for i, line := range records {
		if i == 0 {
			for j, column := range line {
				columnMap[column] = j
			}
		} else {
			_users = append(_users, &types.User{
				EmployeeId:   line[columnMap["Employee Id"]],
				FirstName:    line[columnMap["First Name"]],
				LastName:     line[columnMap["Last Name"]],
				SupervisorId: line[columnMap["Supervisor Id"]],
			})
		}
	}

	return _users
}

func getFile() {
	inputEl := document.Call("querySelector", "input")
	inputEl.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		file := inputEl.Get("files").Index(0)
		reader := js.Global().Get("FileReader").New()
		reader.Call("readAsText", file)
		reader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) any {
			data := reader.Get("result")
			users = readCsvAsUserList(data.String())
			return nil
		}))
		return nil
	}))
}
