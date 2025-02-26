package sudoku

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
)

var SPACE_REGEX = regexp.MustCompile(`  +`)

/*
Parse newline and space separated sudoku problem
0 0 3 ...
9 0 0 ...
0 0 1 ...
...
*/
func NewFromString(input string) *Board {
	input = strings.Trim(input, " \n\t")
	input = SPACE_REGEX.ReplaceAllString(input, " ")
	input = strings.ReplaceAll(input, ".", "0")

	// standard 9x9 single row
	if strings.Index(input, "\n") == -1 {
		return NewFromSingleRowString(input)
	}

	r := bufio.NewReader(strings.NewReader(input))

	rows := strings.Split(input, "\n")
	size2 := len(rows)
	cells := make([][]int, size2)

	for i := 0; i < size2; i++ {
		cells[i] = make([]int, size2)
		for j := 0; j < size2; j++ {
			fmt.Fscan(r, &cells[i][j])
		}
	}

	return NewFromArray(cells)
}

func NewFromSingleRowString(input string) *Board {
	size2 := 9
	size := 3
	board := New(size)

	for i, c := range input {
		if c != '0' && c != '.' {
			board.SetValue(i/size2, i%size2, int(c-'0'))
		}
	}

	return board
}

func (b *Board) ReplaceWithSingleRowString(input string, skipCandidateElimination bool) {
	size2 := 9
	b.NumCandidates = len(b.Candidates) - 1

	for i := 0; i < len(b.Lookup); i++ {
		b.Lookup[i] = 0
	}

	if !skipCandidateElimination {
		for i := 1; i < len(b.Candidates); i++ {
			b.Candidates[i] = true
		}
		for i := 0; i < len(b.rowCandidateCount); i++ {
			b.rowCandidateCount[i] = size2
			b.colCandidateCount[i] = size2
			b.blkCandidateCount[i] = size2
		}
	}

	for i, c := range input {
		if c != '0' && c != '.' {
			if skipCandidateElimination {
				b.Lookup[b.Idx(i/size2, i%size2)] = int(c - '0')
			} else {
				b.SetValue(i/size2, i%size2, int(c-'0'))
			}
		}
	}
}

func NewFromArray(cells [][]int) *Board {
	size2 := len(cells)
	size := getSize(size2)
	board := New(size)

	for r, row := range cells {
		for c, val := range row {
			if val < 1 || val > size2 {
				continue
			}
			board.SetValue(r, c, val)
		}
	}

	return board
}

func (s *Board) Print(w io.Writer) {
	charLen := int(math.Floor(math.Log10(float64(s.Size2 * s.Size2))))
	formatter := fmt.Sprintf("%%s%%%dd%%s", charLen)

	for r := 0; r < s.Size2; r++ {
		fmt.Fprintf(w, formatter, "", s.Lookup[s.Idx(r, 0)], "")
		for c := 1; c < s.Size2-1; c++ {
			fmt.Fprintf(w, formatter, " ", s.Lookup[s.Idx(r, c)], "")
		}
		fmt.Fprintf(w, formatter, " ", s.Lookup[s.Idx(r, s.Size2-1)], "\n")
	}
}

func (s *Board) PrintOneLine(w io.Writer) {
	for i := 0; i < s.Size2*s.Size2; i++ {
		fmt.Fprint(w, s.Lookup[i])
	}
	fmt.Fprintln(w)
}
