package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

type Direction int

const (
	None Direction = iota
	Vertical
	Horizontal
	DiagonalDown
	DiagonalUp
)

func (d Direction) String() string {
	switch d {
	case None:
		return "None"
	case Vertical:
		return "Vertical"
	case Horizontal:
		return "Horizontal"
	case DiagonalDown:
		return "DiagonalDown"
	case DiagonalUp:
		return "DiagonalUp"
	}
	return "Unknown"
}

type Placement struct {
	Original string
	Word     string
	Dir      Direction
	Row      int
	Col      int
}

type WordSearch struct {
	Title       string
	Description string
	Words       []*Placement
	NRows       int
	NCols       int
	Data        [][]byte
}

func NewWordSearch(title, description string, nrows int, ncols int) *WordSearch {
	ws := &WordSearch{
		Title:       title,
		Description: description,
		Words:       make([]*Placement, 0),
		NRows:       ncols,
		NCols:       nrows,
		Data:        make([][]byte, nrows),
	}
	for i := 0; i < nrows; i++ {
		ws.Data[i] = make([]byte, ncols)
	}
	return ws
}

func (ws *WordSearch) Add(word string) {
	if len(word) < 2 {
		return
	}
	w := strings.ToUpper(word)
	w = regexp.MustCompile("[^A-Z]").ReplaceAllString(w, "")
	p := &Placement{
		Original: word,
		Word:     w,
	}
	ws.Words = append(ws.Words, p)
}

func (ws *WordSearch) PlaceHorizontal(word string, row int, col int) bool {
	if len(word)+col > ws.NRows {
		return false
	}
	// try without placing first, and then place if successful
	for _, place := range []bool{false, true} {
		for i, c := range word {
			if ws.Data[row][col+i] != 0 && ws.Data[row][col+i] != byte(c) {
				return false
			}
			if place {
				ws.Data[row][col+i] = byte(c)
			}
		}
	}
	return true
}

func (ws *WordSearch) PlaceVertical(word string, row int, col int) bool {
	if len(word)+row > ws.NCols {
		return false
	}
	for _, place := range []bool{false, true} {
		for i, c := range word {
			if ws.Data[row+i][col] != 0 && ws.Data[row+i][col] != byte(c) {
				return false
			}
			if place {
				ws.Data[row+i][col] = byte(c)
			}
		}
	}
	return true
}

func (ws *WordSearch) PlaceDiagonalDown(word string, row int, col int) bool {
	if len(word)+row > ws.NCols || len(word)+col > ws.NRows {
		return false
	}
	for _, place := range []bool{false, true} {
		for i, c := range word {
			if ws.Data[row+i][col+i] != 0 && ws.Data[row+i][col+i] != byte(c) {
				return false
			}
			if place {
				ws.Data[row+i][col+i] = byte(c)
			}
		}
	}
	return true
}

func (ws *WordSearch) PlaceDiagonalUp(word string, row int, col int) bool {
	if len(word)+row > ws.NCols || col-len(word) < 0 {
		return false
	}
	for _, place := range []bool{false, true} {
		for i, c := range word {
			if ws.Data[row+i][col-i] != 0 && ws.Data[row+i][col-i] != byte(c) {
				return false
			}
			if place {
				ws.Data[row+i][col-i] = byte(c)
			}
		}
	}
	return true
}

func (ws *WordSearch) Place(word string, row int, col int, direction Direction) bool {
	switch direction {
	case Horizontal:
		return ws.PlaceHorizontal(word, row, col)
	case Vertical:
		return ws.PlaceVertical(word, row, col)
	case DiagonalDown:
		return ws.PlaceDiagonalDown(word, row, col)
	case DiagonalUp:
		return ws.PlaceDiagonalUp(word, row, col)
	}
	return false
}

func (ws *WordSearch) FillUnused(dots bool) {
	for row := 0; row < ws.NCols; row++ {
		for col := 0; col < ws.NRows; col++ {
			if ws.Data[row][col] == 0 {
				if dots {
					ws.Data[row][col] = byte('.')
				} else {
					ws.Data[row][col] = byte('A' + rand.Intn(26))
				}
			}
		}
	}
}

func (ws *WordSearch) Build(dirs []Direction) bool {
	rows := make([]int, ws.NCols)
	for row := 0; row < ws.NCols; row++ {
		rows[row] = row
	}

	cols := make([]int, ws.NRows)
	for col := 0; col < ws.NRows; col++ {
		cols[col] = col
	}

	for _, p := range ws.Words {
		placed := false
		// for each word visit the possibilities in random order
		rand.Shuffle(len(rows), func(i, j int) {
			rows[i], rows[j] = rows[j], rows[i]
		})
		rand.Shuffle(len(cols), func(i, j int) {
			cols[i], cols[j] = cols[j], cols[i]
		})
		rand.Shuffle(len(dirs), func(i, j int) {
			dirs[i], dirs[j] = dirs[j], dirs[i]
		})
	outer:
		for _, row := range rows {
			for _, col := range cols {
				for _, dir := range dirs {
					if ws.Place(p.Word, row, col, dir) {
						p.Dir = dir
						p.Row = row
						p.Col = col
						placed = true
						break outer
					}
				}
			}
		}
		if !placed {
			return false
		}
	}
	return true
}

func (ws *WordSearch) Shuffle() {
	rand.Shuffle(len(ws.Words), func(i, j int) {
		ws.Words[i], ws.Words[j] = ws.Words[j], ws.Words[i]
	})
}

func (ws *WordSearch) Print() {
	for _, row := range ws.Data {
		for _, c := range row {
			fmt.Print(string(c))
		}
		fmt.Println()
	}
}

func (ws *WordSearch) PrintSolution() {
	for _, p := range ws.Words {
		fmt.Printf("%s: %s (%d, %d)\n", p.Original, p.Dir, p.Row, p.Col)
	}
}
