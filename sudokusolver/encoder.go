package sudokusolver

import (
	"github.com/rkkautsar/sudoku-solver/sudoku"
	"log"
	"time"
)

func GenerateCNFConstraints(board *sudoku.Board, algorithm string) CNFInterface {
	var cnf CNFInterface

	board.InitCompressedLits()

	shouldUseParallel := false

	cnf = &CNF{
		Board:   board,
		Clauses: make([][]int, 0, board.Size2*board.Size2*board.Size2*3),
	}

	if shouldUseParallel {
		cnf = &CNFParallel{
			CNF: cnf.(*CNF),
		}
	}

	// log.Println("known", board.Size2*board.Size2*board.Size2-board.NumCandidates)
	// log.Println("unknown", board.NumCandidates)
	cnf.setInitialNbVar(board.NumCandidates)
	cnf.initializeLits()

	// var b bytes.Buffer
	// cnf.Print(&b)
	// log.Println(b.String())

	if shouldUseParallel {
		cnf.(*CNFParallel).initWorkers()
	}
	exactly1 := cnfExactly1
	if algorithm == "product" {
		exactly1 = cnfExactly1Product
	}
	// log.Println("here", cnf.clauseLen())
	start := time.Now()
	buildCNFCellConstraints(cnf, exactly1)
	// log.Println("here", cnf.clauseLen())
	buildCNFRangeConstraints(cnf, exactly1)
	elapsed := time.Since(start)
	log.Printf("Generating Clauses took %s", elapsed)
	// buildCNFRangeConstraints2(cnf, cnf.getBoard().Rows(), cnfExactly1)
	// buildCNFRangeConstraints2(cnf, cnf.getBoard().Columns(), cnfExactly1)
	// buildCNFRangeConstraints2(cnf, cnf.getBoard().Blocks(), cnfExactly1)

	if shouldUseParallel {
		cnf.(*CNFParallel).closeAndWait()
	}

	// if board.Size > 6 {
	// 	cnf.Simplify(SimplifyOptions{})
	// }

	return cnf
}

func (c *CNFParallel) initializeLits() {
	c.CNF.initializeLits()
}

// Size2 = 4: kích thước 1 chiều sudoku
// Lookup[]: Lưu giá trị đã điền của từng ô
// Nếu b.CLit(r, c, v) trả về index tương ứng của ô đó trong mảng lưu các biến đã compressed.
// Nếu b.CLit(r, c, v) == 0 thì loại bỏ để làm giảm số biến, mệnh đề.
func buildCNFCellConstraints(cnf CNFInterface, builder CNFBuilder) {
	b := cnf.getBoard()
	idx := 0
	for r := 0; r < b.Size2; r++ {
		for c := 0; c < b.Size2; c++ {
			if b.Lookup[idx] != 0 {
				idx++
				continue
			}
			idx++
			lits := make([]int, b.Size2)
			for v := 1; v <= b.Size2; v++ {
				lits[v-1] = b.CLit(r, c, v)
			}
			cnf.addFormula(filterZero(lits), builder)
		}
	}
}

func buildCNFRangeConstraints(
	c CNFInterface,
	builder CNFBuilder,
) {
	b := c.getBoard()
	size := b.Size
	size2 := b.Size2

	// triadAux := c.requestLiterals(2 * size * size2 * size2)
	// getTriadAuxIdx := func(i, j, v, isCol int) int {
	// 	return (v-1)*size*size2 + i*size + j + isCol*(size*size2*size2)
	// }

	for v := 1; v <= size2; v++ {
		for i := 0; i < size2; i++ {
			// blkRowStart := (i / size) * size
			// blkColStart := (i % size) * size
			// vblkTriads := make([]int, size)
			// hblkTriads := make([]int, size)

			// // row triad
			// c.addFormula(triadAux[getTriadAuxIdx(i, 0, v, 0):getTriadAuxIdx(i, size, v, 0)], builder)
			// // col triad
			// c.addFormula(triadAux[getTriadAuxIdx(i, 0, v, 1):getTriadAuxIdx(i, size, v, 1)], builder)

			// for j := 0; j < size; j++ {
			// 	hblkTriads[j] = triadAux[getTriadAuxIdx(blkRowStart+j, i%size, v, 0)]
			// 	vblkTriads[j] = triadAux[getTriadAuxIdx(blkColStart+j, i/size, v, 1)]
			// }

			// // block triads
			// c.addFormula(hblkTriads, builder)
			// c.addFormula(vblkTriads, builder)

			// for j := 0; j < size; j++ {
			// 	rowTriadLits := make([]int, size+1)
			// 	colTriadLits := make([]int, size+1)
			// 	for k := 0; k < size; k++ {
			// 		rowTriadLits[k] = b.CLit(i, j*size+k, v)
			// 		colTriadLits[k] = b.CLit(j*size+k, i, v)
			// 	}
			// 	rowTriadLits[size] = -triadAux[getTriadAuxIdx(i, j, v, 0)]
			// 	colTriadLits[size] = -triadAux[getTriadAuxIdx(i, j, v, 1)]
			// 	c.addFormula(filterZero(rowTriadLits), builder)
			// 	c.addFormula(filterZero(colTriadLits), builder)
			// }

			blkRowStart := (i / size) * size
			blkColStart := (i % size) * size
			rowLits := make([]int, size2)
			colLits := make([]int, size2)
			blkLits := make([]int, size2)
			// rowLits_ := make([]int, size2)
			// colLits_ := make([]int, size2)
			// blkLits_ := make([]int, size2)
			for j := 0; j < size2; j++ {
				// block
				blkRow := blkRowStart + j/size
				blkCol := blkColStart + j%size
				blkLits[j] = b.CLit(blkRow, blkCol, v)
				// blkLits_[j] = b.Lit(blkRow, blkCol, v)

				// row
				rowLits[j] = b.CLit(i, j, v)
				// rowLits_[j] = b.Lit(i, j, v)

				// col
				colLits[j] = b.CLit(j, i, v)
				// colLits_[j] = b.Lit(j, i, v)

			}

			// log.Println("blk", blkRowStart, blkColStart, v)
			// log.Println(blkLits, blkLits_)
			c.addFormula(filterZero(blkLits), builder)
			// log.Println("row", i, v)
			// log.Println(rowLits, rowLits_)
			c.addFormula(filterZero(rowLits), builder)
			// log.Println("col", i, v)
			// log.Println(colLits, rowLits_)
			c.addFormula(filterZero(colLits), builder)
		}
	}
}

func filterZero(slice []int) []int {
	newSlice := slice[:0]
	for _, x := range slice {
		if x != 0 {
			newSlice = append(newSlice, x)
		}
	}
	return newSlice
}
