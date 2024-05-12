package main

import (
	"encoding/csv"
	"fmt"
	"strings"
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

type User struct {
	employeeId   string
	firstName    string
	lastName     string
	supervisorId string
}

var (
	width    float64
	height   float64
	document js.Value
	canvasEl js.Value
	ctx      js.Value
	draw     js.Func
	exCard   *Card
	users    []*User
)

func main() {
	forever := make(chan struct{}, 0)
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

	js.Global().Call("requestAnimationFrame", draw)
	<-forever
}

func drawCard(card *Card) {
	padding := 16
	// border
	ctx.Set("strokeStyle", "white")
	ctx.Call("beginPath")
	ctx.Call("roundRect", card.rect.x-padding, card.rect.y-padding, card.rect.width+(padding*2), card.rect.height+(padding*2), padding)
	ctx.Call("stroke")

	// title
	ctx.Set("font", "24px Roboto")
	ctx.Set("fillStyle", "white")
	ctx.Set("textBaseline", "top")
	ctx.Call("fillText", card.title, card.rect.x, card.rect.y)
}

func drawAllCards() {
	y := 16 + 4
	for _, user := range users {
		fmt.Println(user.firstName)
		drawCard(&Card{
			Rect{16 + 4, y, 200, 80},
			fmt.Sprintf("%s %s", user.firstName, user.lastName),
		})
		y += (80 + 32 + 8)
	}

}

func readCsvAsUserList(data string) []*User {
	fmt.Println("here we go")
	_users := []*User{}

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
			_users = append(_users, &User{
				employeeId:   line[columnMap["Employee Id"]],
				firstName:    line[columnMap["First Name"]],
				lastName:     line[columnMap["Last Name"]],
				supervisorId: line[columnMap["Supervisor Id"]],
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
