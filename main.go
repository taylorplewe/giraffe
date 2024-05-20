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
	draw         js.Func
	usersById    map[string]*types.User
	userLevels   map[int][]*types.User
	cardLevels   map[int][]*types.Card
)

const (
	CARD_WIDTH  = 200
	CARD_HEIGHT = 100
	GAP_X       = 16
	GAP_Y       = 64
	PADDING     = 16
)

func main() {
	forever := make(chan struct{})

	document = js.Global().Get("document")
	global.CanvasEl = document.Call("querySelector", "canvas")
	global.Ctx = global.CanvasEl.Call("getContext", "2d")
	global.Ctx.Set("textBaseline", "top")

	getFile()

	draw = js.FuncOf(func(this js.Value, args []js.Value) any {
		// Pull window size to handle resize
		curBodyW := document.Get("body").Get("clientWidth").Float()
		curBodyH := document.Get("body").Get("clientHeight").Float()
		if curBodyW != canvasWidth || curBodyH != canvasHeight {
			canvasWidth, canvasHeight = curBodyW, curBodyH
			global.CanvasEl.Set("width", canvasWidth)
			global.CanvasEl.Set("height", canvasHeight)
		}
		global.Ctx.Call("save")
		global.Ctx.Call("setTransform", 1, 0, 0, 1, 0, 0)
		global.Ctx.Call("clearRect", 0, 0, global.CanvasEl.Get("width"), global.CanvasEl.Get("height"))
		global.Ctx.Call("restore")

		drawAllCards()

		js.Global().Call("requestAnimationFrame", draw)
		return nil
	})
	defer draw.Release()

	global.Center = types.Point{24, 24}
	eventhandlers.RegisterAll(global.CanvasEl)
	js.Global().Call("requestAnimationFrame", draw)
	<-forever
}

func createCardFromPointAndUser(point types.Point, user *types.User) *types.Card {
	title := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	global.Ctx.Set("font", "24px sans-serif")
	textMetrics := global.Ctx.Call("measureText", title)
	width := textMetrics.Get("width").Int()
	height := textMetrics.Get("fontBoundingBoxDescent").Int()

	card := &types.Card{}
	card.Texts = append(card.Texts, types.Text{
		Point: types.Point{
			X: point.X + PADDING,
			Y: point.Y + PADDING,
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
			Width:  width + (PADDING * 2),
			Height: height + (PADDING * 2),
		},
	}

	return card
}
func drawCard(user *types.User) {
	card := user.Card
	x := global.Center.X + card.X
	y := global.Center.Y + card.Y

	// text
	for _, text := range card.Texts {
		global.Ctx.Set("font", text.Font)
		global.Ctx.Set("fillStyle", text.Color)
		global.Ctx.Call("fillText", text.Text, x+text.X, y+text.Y)
	}

	// border
	global.Ctx.Set("strokeStyle", "white")
	global.Ctx.Call("beginPath")
	global.Ctx.Call("roundRect", x, y, card.Width, card.Height, PADDING)
	global.Ctx.Call("stroke")

	if len(user.DirectReports) > 0 {
		// horizontal line between myself and children
		global.Ctx.Call("beginPath")
		global.Ctx.Call("moveTo", global.Center.X+user.DirectReports[0].Card.X+(CARD_WIDTH/2), y+card.Height+(GAP_Y/2))
		global.Ctx.Call("lineTo", global.Center.X+user.DirectReports[len(user.DirectReports)-1].Card.X+(CARD_WIDTH/2), y+card.Height+(GAP_Y/2))
		global.Ctx.Call("stroke")

		// vertical line connecting me to that line
		global.Ctx.Call("beginPath")
		global.Ctx.Call("moveTo", x+(card.Width/2), y+card.Height)
		global.Ctx.Call("lineTo", x+(card.Width/2), y+card.Height+(GAP_Y/2))
		global.Ctx.Call("stroke")
	}

	if user.Supervisor != nil {
		// line to parent's horizontal line
		global.Ctx.Call("beginPath")
		global.Ctx.Call("moveTo", x+(card.Width/2), y)
		global.Ctx.Call("lineTo", x+(card.Width/2), global.Center.Y+user.Supervisor.Card.Y+user.Supervisor.Card.Height+(GAP_Y/2))
		global.Ctx.Call("stroke")
	}

	// straight line from me to parent
	// global.Ctx.Call("beginPath")
	// global.Ctx.Call("moveTo", x+(card.Width/2), y)
	// global.Ctx.Call("lineTo", user.Supervisor.Card.X+global.Center.X+(CARD_WIDTH/2), user.Supervisor.Card.Y+global.Center.Y+user.Supervisor.Card.Height)
	// global.Ctx.Call("stroke")
}

func drawAllCards() {
	for _, levelUsers := range userLevels {
		for _, user := range levelUsers {
			drawCard(user)
		}
	}
}

func setUsersById(data string) {
	usersById = map[string]*types.User{}
	employeesToSetSupervisor := []*types.User{}

	csvReader := csv.NewReader(strings.NewReader(data))
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Errorf("Error reading csv: %v", err)
		return
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
			if supervisor, exists := usersById[newUser.SupervisorId]; exists {
				newUser.Supervisor = supervisor
				supervisor.DirectReports = append(supervisor.DirectReports, newUser)
			} else {
				employeesToSetSupervisor = append(employeesToSetSupervisor, newUser)
			}
			usersById[newUser.EmployeeId] = newUser
		}
	}

	for _, user := range employeesToSetSupervisor {
		if supervisor, exists := usersById[user.SupervisorId]; exists {
			user.Supervisor = supervisor
			supervisor.DirectReports = append(supervisor.DirectReports, user)
		}
	}
}

func setUserLevel(user *types.User, level int) {
	if _, exists := userLevels[level]; !exists {
		userLevels[level] = []*types.User{}
	}
	userLevels[level] = append(userLevels[level], user)

	for _, child := range user.DirectReports {
		setUserLevel(child, level+1)
	}
}

func setUserLevels() {
	userLevels = map[int][]*types.User{}
	for _, user := range usersById {
		userIsTopLevel := len(user.DirectReports) > 0 && user.Supervisor == nil
		if userIsTopLevel {
			userLevels[0] = append(userLevels[0], user)
		}
	}

	for _, child := range userLevels[0][0].DirectReports {
		setUserLevel(child, 1)
	}
}
func createAllCards() {
	global.Ctx.Set("textBaseline", "top")
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

func createCard(user *types.User, localX int, y int) {
	user.Card = &types.Card{
		Rect: types.Rect{
			Point: types.Point{
				X: localX,
				Y: y,
			},
			Size: types.Size{
				Width:  CARD_WIDTH,
				Height: CARD_HEIGHT,
			},
		},
		Mod: 0,
		Texts: []types.Text{
			types.Text{
				Point: types.Point{
					X: PADDING,
					Y: PADDING,
				},
				Font:  "24px sans-serif",
				Color: "white",
				Text:  fmt.Sprintf("%s %s", user.FirstName, user.LastName),
			},
		},
	}

	// create children
	for i, child := range user.DirectReports {
		//dipY := (i & 1) * (CARD_HEIGHT + GAP_X)
		createCard(child, i*(CARD_WIDTH+GAP_X) /*/2*/, y+CARD_HEIGHT+GAP_Y /*+dipY*/)
	}

	// center myself over children if I'm leftmost, otherwise center them under me
	if len(user.DirectReports) == 0 {
		return
	}
	centeredX := ((len(user.DirectReports) - 1) * (CARD_WIDTH + GAP_X)) / 2
	// if localX == 0 {
	// 	user.Card.X = centeredX
	// } else {
	user.Card.Mod = user.Card.X - centeredX
	// }
}

func applyParentsMods(user *types.User, modSum int) {
	user.Card.X += modSum
	modSum += user.Card.Mod

	for _, child := range user.DirectReports {
		applyParentsMods(child, modSum)
	}

	user.Card.Mod = 0
}

func spaceTwoNodes(leftUser *types.User, rightUser *types.User, amountToMoveRight int) int {
	if rightUser.Card.X < leftUser.Card.X+CARD_WIDTH+GAP_X {
		amountToMoveRight += (leftUser.Card.X + CARD_WIDTH + GAP_X) - rightUser.Card.X
	}

	if len(rightUser.DirectReports) > 0 && len(leftUser.DirectReports) > 0 {
		amountToMoveRight += spaceTwoNodes(leftUser.DirectReports[len(leftUser.DirectReports)-1], rightUser.DirectReports[0], amountToMoveRight)
	}

	return amountToMoveRight
}

func spaceNodesChildren(user *types.User) {
	totalShiftedAmount := 0
	for i, child := range user.DirectReports {
		for _, leftChild := range user.DirectReports[:i] {
			amountToMoveRight := spaceTwoNodes(leftChild, child, 0)
			child.Card.X += amountToMoveRight
			child.Card.Mod += amountToMoveRight
			if amountToMoveRight > 0 {
				fmt.Println(child.FirstName, child.LastName, "needs to move", amountToMoveRight)
			}
			totalShiftedAmount += amountToMoveRight
		}
		spaceNodesChildren(child)
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
			setUsersById(data.String())
			setUserLevels()
			createCard(userLevels[0][0], 0, 0)
			applyParentsMods(userLevels[0][0], 0)
			spaceNodesChildren(userLevels[0][0])
			applyParentsMods(userLevels[0][0], 0)
			return nil
		}))
		return nil
	}))
}
