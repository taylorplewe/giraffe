package types

type Point struct {
	X int
	Y int
}
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

type User struct {
	EmployeeId   string
	FirstName    string
	LastName     string
	SupervisorId string
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
