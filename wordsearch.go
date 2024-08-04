package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"slices"
	"sort"
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

func (ws *WordSearch) Validate() bool {
	errors := 0
	for i := 0; i < len(ws.Words); i++ {
		for j := i + 1; j < len(ws.Words); j++ {
			if ws.Words[i].Word == ws.Words[j].Word {
				fmt.Println("Duplicate word:", ws.Words[i].Word)
				errors++
			}
			if strings.Contains(ws.Words[i].Word, ws.Words[j].Word) {
				fmt.Printf("%s contains %s\n", ws.Words[i].Word, ws.Words[j].Word)
				errors++
			}
			if strings.Contains(ws.Words[j].Word, ws.Words[i].Word) {
				fmt.Printf("%s contains %s\n", ws.Words[j].Word, ws.Words[i].Word)
				errors++
			}
		}
	}
	return errors == 0
}

func (ws *WordSearch) PlaceHorizontal(word string, row int, col int) bool {
	if len(word)+col > ws.NCols {
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
	if len(word)+row > ws.NRows {
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
	if len(word)+row > ws.NRows || len(word)+col > ws.NCols {
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
	if row-len(word) < 0 || len(word)+col > ws.NCols {
		return false
	}
	for _, place := range []bool{false, true} {
		for i, c := range word {
			if ws.Data[row-i][col+i] != 0 && ws.Data[row-i][col+i] != byte(c) {
				return false
			}
			if place {
				ws.Data[row-i][col+i] = byte(c)
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

// We build a single array of all the rows, columns, and directions and then shuffle them
// to visit them in random order. This is a simple way to randomize the placement of the words.
type rcd struct {
	Row int
	Col int
	Dir Direction
}

func (ws *WordSearch) Build(dirs []Direction) bool {
	rcds := make([]rcd, ws.NCols*ws.NRows*len(dirs))
	for row := 0; row < ws.NCols; row++ {
		for col := 0; col < ws.NRows; col++ {
			for _, dir := range dirs {
				rcds = append(rcds, rcd{Row: row, Col: col, Dir: dir})
			}
		}
	}

	dirCounts := make(map[Direction]int)
	for i := range ws.Words {
		// shuffle the rcds to randomize the placement for each word
		rand.Shuffle(len(rcds), func(i, j int) {
			rcds[i], rcds[j] = rcds[j], rcds[i]
		})
		placed := false
		for _, rcd := range rcds {
			if ws.Place(ws.Words[i].Word, rcd.Row, rcd.Col, rcd.Dir) {
				ws.Words[i].Row = rcd.Row + 1
				ws.Words[i].Col = rcd.Col + 1
				ws.Words[i].Dir = rcd.Dir
				dirCounts[rcd.Dir]++
				placed = true
				break
			}
		}
		if !placed {
			return false
		}
	}
	// let's make sure that we have at least one word in each direction
	for _, dir := range dirs {
		if dirCounts[dir] < 1 {
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

func (ws *WordSearch) Longest() int {
	longest := 0
	for _, p := range ws.Words {
		if len(p.Word) > longest {
			longest = len(p.Word)
		}
	}
	return longest
}

func (ws *WordSearch) WordList() []string {
	words := make([]string, len(ws.Words))
	for i, p := range ws.Words {
		words[i] = p.Original
	}
	sort.StringSlice(words).Sort()
	return words
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
	slices.SortFunc(ws.Words, func(a, b *Placement) int {
		return strings.Compare(a.Original, b.Original)
	})
	for _, p := range ws.Words {
		fmt.Printf("%22s: %16s (R%d, C%d)\n", p.Original, p.Dir, p.Row, p.Col)
	}
}
