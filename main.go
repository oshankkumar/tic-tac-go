package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"
	"time"
)

const N = 3

type Marker string

const (
	X Marker = "X"
	O Marker = "O"
)

const boardTmpl = `
{{range .Players}}
Name: {{.Name}} Choice: {{.Marker}}
{{end}}
-------------{{range .Board}} 
|{{range .}} {{.}} |{{end}}
-------------{{end}}
`

type Cell [2]int

type WinningRule func(cells []Cell) bool

func ColumnMatch(cells []Cell) bool {
	n := 0
	colNum := cells[0][1]
	for _, c := range cells[1:] {
		if c[1] == colNum {
			n++
		}
	}
	return n >= N-1
}

func RowMatch(cells []Cell) bool {
	n := 0
	rowNum := cells[0][0]
	for _, c := range cells[1:] {
		if c[0] == rowNum {
			n++
		}
	}
	return n >= N-1
}

func CrossDiagonalMatch(cells []Cell) bool {
	n := 0
	sum := cells[0][0] + cells[0][1]
	for _, c := range cells[1:] {
		if c[0]+c[1] == sum {
			n++
		}
	}
	return n >= N-1
}

func DiagonalMatch(cells []Cell) bool {
	n := 0
	for _, c := range cells[1:] {
		if c[0] == c[1] {
			n++
		}
	}
	return n >= N
}

func IsWinner(cells []Cell) bool {
	if len(cells) < N {
		return false
	}
	rules := []WinningRule{CrossDiagonalMatch, RowMatch, ColumnMatch, DiagonalMatch}
	for _, rule := range rules {
		if rule(cells) {
			return true
		}
	}
	return false
}

type Board [N][N]Marker

func NewBoard() *Board {
	b := &Board{}
	for i := 0; i < N; i++ {
		line := [N]Marker{}
		for j := 0; j < N; j++ {
			line[j] = "_"
		}
		b[i] = line
	}
	return b
}

func DisplayBoard(b *Board, p [2]*Player) string {
	buf := &bytes.Buffer{}
	template.Must(template.New("board").Parse(boardTmpl)).Execute(buf, struct {
		Board   *Board
		Players [2]*Player
	}{b, p})
	return buf.String()
}

func (b *Board) CanFillAtCell(c Cell) bool {
	return b[c[0]][c[1]] != X && b[c[0]][c[1]] != O
}

func (b *Board) FillCell(c Cell, m Marker) {
	b[c[0]][c[1]] = m
}

type Player struct {
	Name   string
	Marker Marker
}

func readPlayers() [2]*Player {
	n1 := readString("Enter Player1 Name: ")
	marker := readString("Enter Marker Choice For %s [X/O]: ", n1)
	for marker != string(X) && marker != string(O) {
		marker = readString("Invalid Input, Please Enter Marker Choice [X/O]: ")
	}
	n2 := readString("Enter Player2 Name: ")

	p1 := NewPlayer(n1, Marker(marker))
	marker2 := O
	if p1.Marker == O {
		marker2 = X
	}
	p2 := NewPlayer(n2, marker2)

	return [2]*Player{p1, p2}
}

func NewPlayer(n string, m Marker) *Player {
	return &Player{Marker: m, Name: n}
}

type Game struct {
	board *Board
}

func (g *Game) Mark(p *Player, pos int) bool {
	if pos < 0 || pos > N*N-1 {
		return false
	}

	if !g.board.CanFillAtCell(Cell{pos / N, pos % N}) {
		return false
	}

	g.board.FillCell(Cell{pos / N, pos % N}, p.Marker)
	return true
}

func (g *Game) CheckWinner(players [2]*Player) (*Player, bool) {
	p1, p2 := players[0], players[1]
	p1Cells, p2Cells := []Cell{}, []Cell{}

	for i, line := range g.board {
		for j, m := range line {
			switch m {
			case p1.Marker:
				p1Cells = append(p1Cells, Cell{i, j})
			case p2.Marker:
				p2Cells = append(p2Cells, Cell{i, j})
			}
		}
	}

	if IsWinner(p1Cells) {
		return p1, true
	}
	if IsWinner(p2Cells) {
		return p2, true
	}
	return nil, false
}

func (g *Game) CheckTie() bool {
	n := 0
	for _, l := range g.board {
		for _, m := range l {
			if m == X || m == O {
				n++
			}
		}
	}
	return n == N*N
}

func (g *Game) startRender() {
	ClearScreen()
	SlowPrint(`TIC TAC TOE LOADING ..................................................`, time.Millisecond*50)
	ClearScreen()
	SlowPrint("\nGame Started\n", time.Millisecond*50)
	SlowPrint(DisplayBoard(g.board, [2]*Player{{}, {}}), time.Millisecond*30)
}

func (g *Game) Start(players [2]*Player) {
	for {
		for _, player := range players {
			ClearAndPrintBoard(g.board, players)
			pos := readInt("Enter Marker Position (%s): ", player.Name)
			for !g.Mark(player, pos) {
				pos = readInt("Invalid Position, Enter Correct Marker Position (%s): ", player.Name)
			}

			p, ok := g.CheckWinner(players)
			if ok {
				ClearAndPrintBoard(g.board, players)
				fmt.Printf("Congrats %s Wins\n", p.Name)
				return
			}
			if g.CheckTie() {
				ClearAndPrintBoard(g.board, players)
				fmt.Println("Match Got Tie")
				return
			}
		}
	}
}

func ClearAndPrintBoard(b *Board, p [2]*Player) {
	ClearScreen()
	fmt.Println(DisplayBoard(b, p))
}

func SlowPrint(s string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for _, s := range s {
		fmt.Print(string(s))
		<-ticker.C
	}
	ticker.Stop()
}

func ClearScreen() {
	//fmt.Println("\033[2J")
	cmd := exec.Command("clear") //Linux example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func readString(prompt string, args ...interface{}) string {
	var s string
	fmt.Printf(prompt, args...)
	fmt.Scanln(&s)
	return s
}

func readInt(prompt string, args ...interface{}) int {
	var d int
	fmt.Printf(prompt, args...)
	fmt.Scanln(&d)
	return d
}

func PromptConfirm(prompt string, args ...interface{}) bool {
	for {
		val := readString(prompt, args...)
		switch val {
		case "yes", "Y", "y":
			return true
		case "n", "no", "N":
			return false
		}
	}
}

func main() {
	g := &Game{board: NewBoard()}
	g.startRender()
	players := readPlayers()
	g.Start(players)
	for {
		if !PromptConfirm("Do you want to Play Again [y/n]: ") {
			break
		}
		g.board = NewBoard()
		g.Start(players)
	}
}
