package types

type Point struct {
	X int
	Y int
}
type Size struct {
	Width  int
	Height int
}
type Rect struct {
	Point
	Size
}

type User struct {
	EmployeeId    string
	FirstName     string
	LastName      string
	SupervisorId  string
	Supervisor    *User
	DirectReports []*User
	Card          *Card
}
type Card struct {
	Rect
	Mod   int
	Texts []Text
}
type Text struct {
	Point
	Font  string
	Color string
	Text  string
}

func (p *Point) Add(p2 Point) Point {
	return Point{
		p.X + p2.X,
		p.Y + p2.Y,
	}
}
func (p *Point) Sub(p2 Point) Point {
	return Point{
		p.X - p2.X,
		p.Y - p2.Y,
	}
}
