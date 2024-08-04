package main

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
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

func (ws *WordSearch) AddSearchPageToPDF(pdf *fpdf.Fpdf) {
	width, height, _ := pdf.PageSize(0)

	pdf.AddPage()
	cellWidth := width / float64(ws.NRows) * .75
	cellHeight := height / float64(ws.NCols) * .75

	cellWidth = min(cellWidth, cellHeight)
	cellHeight = cellWidth
	ml, mt, mr, mb := pdf.GetMargins()

	pdf.SetFont("Arial", "B", 24)
	pdf.CellFormat(width-ml-mr, 10, ws.Title, "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 18)
	pdf.CellFormat(width-ml-mr, 10, ws.Description, "", 1, "C", false, 0, "")

	// let's center the grid
	gridwidth := cellWidth * float64(ws.NRows)
	gridheight := cellHeight * float64(ws.NCols)

	xindent := (width - ml - mr - gridwidth) / 2

	// calculate vertical space to center grid in remaining space
	yspace := (height - mt - mb - pdf.GetY() - gridheight) / 2
	pdf.CellFormat(width-ml-mr, yspace, "", "", 1, "", false, 0, "")

	pdf.SetFont("Arial", "B", 24)
	for row := 0; row < ws.NCols; row++ {
		pdf.CellFormat(xindent, yspace, "", "", 0, "", false, 0, "")
		for col := 0; col < ws.NRows; col++ {
			pdf.CellFormat(cellWidth, cellHeight, string(ws.Data[row][col]), "", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	}
}

type Data struct {
	Title       string
	Description string
	Words       []string
}

type Opts struct {
	MinCols          int    `short:"c" long:"min-cols" description:"Minimum number of columns" default:"10"`
	MinRows          int    `short:"r" long:"min-rows" description:"Minimum number of rows" default:"13"`
	MaxCols          int    `short:"C" long:"max-cols" description:"Maximum number of cols" default:"22"`
	MaxRows          int    `short:"R" long:"max-rows" description:"Maximum number of rows" default:"25"`
	NumTries         int    `short:"n" long:"num-tries" description:"Max number of tries at each level" default:"100"`
	PrintSolution    bool   `short:"s" long:"print-solution" description:"Print the solution"`
	UseDownDiagonals bool   `short:"d" long:"down-diagonals" description:"Use down diagonals"`
	UseUpDiagonals   bool   `short:"u" long:"up-diagonals" description:"Use up diagonals"`
	FillDots         bool   `short:"D" long:"fill-dots" description:"Fill unused cells with dots instead of letters"`
	PDFName          string `short:"p" long:"pdf" description:"Output PDF file (if not specified, output to stdout)"`
}

func Generate(filename string, opts *Opts) (*WordSearch, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var data Data
	err = yaml.NewDecoder(file).Decode(&data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	file.Close()

	dirs := []Direction{Horizontal, Vertical}
	if opts.UseDownDiagonals {
		dirs = append(dirs, DiagonalDown)
	}
	if opts.UseUpDiagonals {
		dirs = append(dirs, DiagonalUp)
	}
	var ws *WordSearch

	nrows := opts.MinRows
	ncols := opts.MinCols
	for ; nrows <= opts.MaxRows && ncols <= opts.MaxCols; nrows, ncols = nrows+1, ncols+1 {
		fmt.Println("trying", nrows, " rows and ", ncols, " cols")
		ws = NewWordSearch(data.Title, data.Description, nrows, ncols)
		for _, word := range data.Words {
			ws.Add(word)
		}
		for tries := 0; tries < opts.NumTries; tries++ {
			ws.Shuffle()
			if ws.Build(dirs) {
				ws.FillUnused(opts.FillDots)
				return ws, nil
			}
		}
	}
	return nil, fmt.Errorf("failed")
}

func main() {
	var opts Opts
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if len(args) < 1 {
		fmt.Println("Usage: wordsearch filename")
		os.Exit(1)
	}

	var pdf *fpdf.Fpdf
	if opts.PDFName != "" {
		pdf = fpdf.New("P", "mm", "Letter", "")
	}
	for _, f := range args {
		fmt.Printf("------ %s ------\n", f)
		ws, err := Generate(f, &opts)
		if err != nil {
			fmt.Println(err)
		} else {
			if pdf != nil {
				ws.AddSearchPageToPDF(pdf)
			} else {
				ws.Print()
				if opts.PrintSolution {
					ws.PrintSolution()
				}
			}
		}
	}
	if pdf != nil {
		err := pdf.OutputFileAndClose(opts.PDFName)
		if err != nil {
			fmt.Println(err)
		}
	}
}
