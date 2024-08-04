package main

import (
	"fmt"
	"os"

	"github.com/go-pdf/fpdf"
	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
)

func AddSearchPageToPDF(ws *WordSearch, pdf *fpdf.Fpdf) {
	width, height, _ := pdf.PageSize(0)

	pdf.AddPage()
	cellWidth := width / float64(ws.NRows) * .75
	cellHeight := height / float64(ws.NCols) * .75

	cellWidth = min(cellWidth, cellHeight)
	cellHeight = cellWidth
	ml, _, mr, _ := pdf.GetMargins()

	pdf.SetFont("Arial", "B", 24)
	pdf.CellFormat(width-ml-mr, 10, ws.Title, "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 18)
	pdf.CellFormat(width-ml-mr, 10, ws.Description, "", 1, "C", false, 0, "")

	// let's center the grid horizontally and keep a small gap vertically
	gridwidth := cellWidth * float64(ws.NRows)

	xindent := (width - ml - mr - gridwidth) / 2
	pdf.CellFormat(width-ml-mr, cellHeight/2, "", "", 1, "", false, 0, "")

	pdf.SetFont("Arial", "B", 24)
	for row := 0; row < ws.NCols; row++ {
		pdf.CellFormat(xindent, 0, "", "", 0, "", false, 0, "")
		for col := 0; col < ws.NRows; col++ {
			pdf.CellFormat(cellWidth, cellHeight, string(ws.Data[row][col]), "", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	}

	// now let's add the word list
	numCols := 5
	pdf.SetFont("Arial", "", 15)
	for i, word := range ws.Words {
		if i%numCols == 0 {
			pdf.Ln(-1)
		}
		pdf.CellFormat((width-ml-mr)/float64(numCols), 10, word.Original, "", 0, "L", false, 0, "")
	}

}

type Data struct {
	Title       string
	Description string
	Words       []string
}

type Opts struct {
	MinCols          int    `short:"c" long:"min-cols" description:"Minimum number of columns" default:"12"`
	MinRows          int    `short:"r" long:"min-rows" description:"Minimum number of rows" default:"12"`
	MaxCols          int    `short:"C" long:"max-cols" description:"Maximum number of cols" default:"25"`
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
				AddSearchPageToPDF(ws, pdf)
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
