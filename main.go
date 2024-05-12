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
	canvasWidth  float64
	canvasHeight float64
	document     js.Value
	canvasEl     js.Value
	ctx          js.Value
	draw         js.Func
	users        []*types.User
	cards        []*types.Card
)

func main() {
	forever := make(chan struct{})

	document = js.Global().Get("document")
	canvasEl = document.Call("querySelector", "canvas")
	ctx = canvasEl.Call("getContext", "2d")
	ctx.Set("textBaseline", "top")

	getFile()

	draw = js.FuncOf(func(this js.Value, args []js.Value) any {
		// Pull window size to handle resize
		curBodyW := document.Get("body").Get("clientWidth").Float()
		curBodyH := document.Get("body").Get("clientHeight").Float()
		if curBodyW != canvasWidth || curBodyH != canvasHeight {
			canvasWidth, canvasHeight = curBodyW, curBodyH
			canvasEl.Set("width", canvasWidth)
			canvasEl.Set("height", canvasHeight)
		}
		ctx.Call("clearRect", 0, 0, canvasWidth, canvasHeight)

		drawAllCards()

		js.Global().Call("requestAnimationFrame", draw)
		return nil
	})
	defer draw.Release()

	global.Center = types.Point{24, 24}
	eventhandlers.RegisterAll(canvasEl)
	js.Global().Call("requestAnimationFrame", draw)
	<-forever
}

func createCardFromPointAndUser(point types.Point, user *types.User) *types.Card {
	padding := 16
	title := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	card := &types.Card{}
	card.Texts = append(card.Texts, types.Text{
		Point: types.Point{
			X: point.X,
			Y: point.Y + padding,
		},
		Font:  "24px Roboto",
		Color: "white",
		Text:  title,
	})

	ctx.Set("font", "24px Roboto")
	textMetrics := ctx.Call("measureText", title)
	width := textMetrics.Get("width").Int()
	height := textMetrics.Get("fontBoundingBoxDescent").Int()

	card.Rect = types.Rect{
		Point: types.Point{
			X: point.X - padding,
			Y: point.Y - padding,
		},
		Size: types.Size{
			Width:  width + (padding * 2),
			Height: height + (padding * 2),
		},
	}

	cards = append(cards, card)
	return card
}
func drawCard(card *types.Card) {
	padding := 16
	x := global.Center.X + card.Rect.Point.X
	y := global.Center.Y + card.Rect.Point.Y

	// text
	for _, text := range card.Texts {
		ctx.Set("font", text.Font)
		ctx.Set("fillStyle", text.Color)
		ctx.Call("fillText", text.Text, text.Point.X+global.Center.X, text.Point.Y+global.Center.Y)
	}

	// border
	ctx.Set("strokeStyle", "white")
	ctx.Call("beginPath")
	ctx.Call("roundRect", x, y, card.Rect.Size.Width, card.Rect.Size.Height, padding)
	ctx.Call("stroke")
}

func createAllCards() {
	y := global.Center.Y
	x := global.Center.X
	for _, user := range users {
		card := createCardFromPointAndUser(types.Point{
			X: x,
			Y: y,
		}, user)
		x += card.Rect.Size.Width + 8
		if x > int(canvasWidth) {
			x = global.Center.X
			y += card.Rect.Size.Height + 8
		}
	}
}
func drawAllCards() {
	for _, card := range cards {
		drawCard(card)
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
			createAllCards()
			return nil
		}))
		return nil
	}))
}
