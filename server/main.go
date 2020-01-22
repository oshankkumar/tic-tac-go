package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"text/template"
	"time"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1200")
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	fmt.Println("listening on", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go renderGame(conn)
	}
}

func renderGame(conn net.Conn) {
	defer conn.Close()
	g := &Game{board: NewBoard()}
	g.startRender(conn)
	players := readPlayers(conn)
	g.Start(conn, players)
	for {
		if !PromptConfirm(conn, "Do you want to Play Again [y/n]: ") {
			break
		}
		g.board = NewBoard()
		g.Start(conn, players)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, "Fatal error : %s", err)
		os.Exit(1)
	}
}

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
	for _, c := range cells {
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

func readPlayers(rw io.ReadWriter) [2]*Player {
	n1 := readString(rw, "Enter Player1 Name: ")
	marker := readString(rw, "Enter Marker Choice For %s [X/O]: ", n1)
	for marker != string(X) && marker != string(O) {
		marker = readString(rw, "Invalid Input, Please Enter Marker Choice [X/O]: ")
	}
	n2 := readString(rw, "Enter Player2 Name: ")

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

func (g *Game) startRender(rw io.ReadWriter) {
	ClearScreen(rw)
	SlowPrint(rw, `TIC TAC TOE LOADING ..................................................`, time.Millisecond*50)
	ClearScreen(rw)
	SlowPrint(rw, "\nGame Started\n", time.Millisecond*50)
	SlowPrint(rw, DisplayBoard(g.board, [2]*Player{{}, {}}), time.Millisecond*30)
}

func (g *Game) Start(rw io.ReadWriter, players [2]*Player) {
	for {
		for _, player := range players {
			ClearAndPrintBoard(rw, g.board, players)
			pos := readInt(rw, "Enter Marker Position (%s): ", player.Name)
			for !g.Mark(player, pos) {
				pos = readInt(rw, "Invalid Position, Enter Correct Marker Position (%s): ", player.Name)
			}

			p, ok := g.CheckWinner(players)
			if ok {
				ClearAndPrintBoard(rw, g.board, players)
				fmt.Fprintf(rw, "Congrats %s Wins\n", p.Name)
				return
			}
			if g.CheckTie() {
				ClearAndPrintBoard(rw, g.board, players)
				fmt.Fprintln(rw, "Match Got Tie")
				return
			}
		}
	}
}

func ClearAndPrintBoard(rw io.ReadWriter, b *Board, p [2]*Player) {
	ClearScreen(rw)
	fmt.Fprintln(rw, DisplayBoard(b, p))
}

func SlowPrint(w io.Writer, s string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for _, s := range s {
		fmt.Fprint(w, string(s))
		<-ticker.C
	}
	ticker.Stop()
}

func ClearScreen(rw io.ReadWriter) {
	fmt.Fprintln(rw, "\033[2J")
}

func clearScr() {
	cmd := exec.Command("clear") //Linux example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func readString(rw io.ReadWriter, prompt string, args ...interface{}) string {
	var s string
	fmt.Fprintf(rw, prompt, args...)
	fmt.Fscanln(rw, &s)
	return s
}

func readInt(rw io.ReadWriter, prompt string, args ...interface{}) int {
	var d int
	fmt.Fprintf(rw, prompt, args...)
	fmt.Fscanln(rw, &d)
	return d
}

func PromptConfirm(rw io.ReadWriter, prompt string, args ...interface{}) bool {
	for {
		val := readString(rw, prompt, args...)
		switch val {
		case "yes", "Y", "y":
			return true
		case "n", "no", "N":
			return false
		}
	}
}
