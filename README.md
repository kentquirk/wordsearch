# wordsearch
A command-line generator for wordsearch puzzles; creates plain text output or a PDF from a collection of data files.

The data files are YAML files with this structure:

```yaml
title: My Title
description: A sample wordsearch
words:
  - first
  - second
  - it's a good word
  - find this
```

The words are converted to all upper-case with non-alpha characters removed.

The algorithm is simple -- it just tries to place each word, in shuffled order, at each possible grid location until it finds a fit. Each location is tried horizontally, vertically, and optionally diagonally down and diagonally up.

The order of rows, columns, and directions in which things are tried is also randomized so the puzzle doesn't cluster in one direction or another.

If it gets to the point where it cannot place one of the words, it starts again with a larger grid. The range of grid sizes it will try is also controllable.

Specifying a PDF filename generates a single PDF file with all of puzzles in a single document.

To see all the options, type -h.

