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
	usersById    map[string]*types.User
	userLevels	 map[int][]*types.User
	cardLevels	 map[int][]*types.Card
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
	ctx.Set("font", "24px sans-serif")
	textMetrics := ctx.Call("measureText", title)
	width := textMetrics.Get("width").Int()
	height := textMetrics.Get("fontBoundingBoxDescent").Int()

	card := &types.Card{}
	card.Texts = append(card.Texts, types.Text{
		Point: types.Point{
			X: point.X + padding,
			Y: point.Y + padding,
		},
		Font:  "24px sans-serif",
		Color: "white",
		Text:  title,
	})

	card.Rect = types.Rect{
		Point: types.Point{
			X: point.X,
			Y: point.Y,
		},
		Size: types.Size{
			Width:  width + (padding * 2),
			Height: height + (padding * 2),
		},
	}

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
	ctx.Set("textBaseline", "top")	
	y := global.Center.Y
	x := global.Center.X
	cardLevels = map[int][]*types.Card{}

	level := 0
	for {
		if _, exists := userLevels[level]; !exists {
			break
		}
		levelUsers := userLevels[level]
		cardLevels[level] = []*types.Card{}
		for _, user := range levelUsers {
			card := createCardFromPointAndUser(types.Point{
				X: x,
				Y: y,
			}, user)
			cardLevels[level] = append(cardLevels[level], card)
			x += card.Rect.Size.Width + 8
		}
		x = global.Center.X
		y += cardLevels[level][0].Rect.Size.Height + 8
		level++
	}
}
func drawAllCards() {
	for _, levelCards := range cardLevels {
		for _, card := range levelCards {
			drawCard(card)
		}
	}
}

func readCsvAsUserMap(data string) map[string]*types.User {
	_users := map[string]*types.User{}
	employeesToSetSupervisor := []*types.User{}

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
			newUser := &types.User{
				EmployeeId:    line[columnMap["Employee Id"]],
				FirstName:     line[columnMap["First Name"]],
				LastName:      line[columnMap["Last Name"]],
				SupervisorId:  line[columnMap["Supervisor Id"]],
				Supervisor:    nil,
				DirectReports: []*types.User{},
			}
			// check if supervisor exists
			if supervisor, exists := _users[newUser.SupervisorId]; exists {
				newUser.Supervisor = supervisor
				supervisor.DirectReports = append(supervisor.DirectReports, newUser)
			} else {
				employeesToSetSupervisor = append(employeesToSetSupervisor, newUser)
			}
			_users[newUser.EmployeeId] = newUser
		}
	}

	for _, user := range employeesToSetSupervisor {
		if supervisor, exists := _users[user.SupervisorId]; exists {
			user.Supervisor = supervisor
			supervisor.DirectReports = append(supervisor.DirectReports, user)
		}
	}

	return _users
}

func organizeHierarchy() {
	userLevels = map[int][]*types.User{}
	for _, user := range usersById {
		userIsTopLevel := len(user.DirectReports) > 0 && user.Supervisor == nil
		if userIsTopLevel {
			userLevels[0] = append(userLevels[0], user)
		}
	}

	for _, user := range userLevels[0] {
		userLevels[1] = append(userLevels[1], user.DirectReports...)
	}
}

func getFile() {
	inputEl := document.Call("querySelector", "input")
	inputEl.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		file := inputEl.Get("files").Index(0)
		reader := js.Global().Get("FileReader").New()
		reader.Call("readAsText", file)
		reader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) any {
			data := reader.Get("result")
			usersById = readCsvAsUserMap(data.String())
			organizeHierarchy()
			createAllCards()
			return nil
		}))
		return nil
	}))
}
