package captcha

import (
	db "captcha_sweeper/internal/database"
	"crypto/rand"
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
)

type MineSweeper struct {
	GridSize        int32
	MineCount       int32
	Grid            [][]int32
	Revealed        [][]bool
	Solved          pgtype.Bool
	Failed          pgtype.Bool
	DifficultyLevel pgtype.Int4
}

// NewMineSweeperFromCaptcha initializes a MineSweeper instance from a sqlc-generated Captcha model.
func NewMineSweeperFromCaptcha(c db.Captcha) *MineSweeper {
	m := &MineSweeper{
		GridSize:        c.GridSize,
		MineCount:       c.MineCount,
		Grid:            c.Grid,
		Revealed:        c.Revealed,
		Solved:          c.Solved,
		Failed:          c.Failed,
		DifficultyLevel: c.DifficultyLevel,
	}

	return m
}

func (m *MineSweeper) ToCaptcha() db.Captcha {
	return db.Captcha{
		GridSize:        m.GridSize,
		MineCount:       m.MineCount,
		Grid:            m.Grid,
		Revealed:        m.Revealed,
		Solved:          m.Solved,
		Failed:          m.Failed,
		DifficultyLevel: m.DifficultyLevel,
	}
}

func NewMineSweeper(difficultyLevel int32) *MineSweeper {
	gridSize, mineCount := GetDifficultyParams(difficultyLevel)
	return &MineSweeper{
		GridSize:        gridSize,
		MineCount:       mineCount,
		Grid:            make([][]int32, gridSize),
		Revealed:        make([][]bool, gridSize),
		Solved:          pgtype.Bool{Bool: false, Valid: true},
		Failed:          pgtype.Bool{Bool: false, Valid: true},
		DifficultyLevel: pgtype.Int4{Int32: difficultyLevel, Valid: true},
	}
}

func (m *MineSweeper) generateGrid(size, mineCount int) (grid [][]int32, mines [][]bool) {
	// Pre-allocate flat slices for grid and mines
	flatGrid := make([]int32, size*size)
	flatMines := make([]bool, size*size)

	// Place mines randomly
	placed := 0
	for placed < mineCount {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(size*size)))
		if err != nil {
			return nil, nil // Handle error gracefully
		}
		idxInt := int(idx.Int64())
		if !flatMines[idxInt] {
			flatMines[idxInt] = true
			flatGrid[idxInt] = -1 // -1 = mine
			placed++
		}
	}

	// Convert flat slices to 2D for easier access
	grid = make([][]int32, size)
	mines = make([][]bool, size)
	for i := range grid {
		grid[i] = flatGrid[i*size : (i+1)*size]
		mines[i] = flatMines[i*size : (i+1)*size]
	}

	// Calculate numbers for each cell
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if !mines[i][j] {
				grid[i][j] = m.countAdjacentMines(i, j, mines, size)
			}
		}
	}

	return grid, mines
}

// countAdjacentMines avoids bounds checks by iterating only over valid neighbors.
func (m *MineSweeper) countAdjacentMines(row, col int, mines [][]bool, size int) int32 {
	count := int32(0)
	// Define the 8 possible directions
	directions := [][2]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}
	for _, dir := range directions {
		newRow, newCol := row+dir[0], col+dir[1]
		if newRow >= 0 && newRow < size && newCol >= 0 && newCol < size && mines[newRow][newCol] {
			count++
		}
	}
	return count
}

func (m *MineSweeper) Initialize() {
	grid, _ := m.generateGrid(int(m.GridSize), int(m.MineCount))
	m.Grid = grid

	m.Revealed = make([][]bool, len(grid))
	for i := range grid {
		m.Revealed[i] = make([]bool, len(grid[i]))
	}

	m.Solved = pgtype.Bool{Bool: false, Valid: true}
	m.Failed = pgtype.Bool{Bool: false, Valid: true}
}

func (m *MineSweeper) inBounds(r, c int32) bool {
	return r >= 0 && c >= 0 && r < m.GridSize && c < m.GridSize
}

// Reveal reveals the cell at (row,col). It returns whether the move was safe.
// If the revealed cell is a mine, the move is unsafe and the game is considered lost.
func (m *MineSweeper) Reveal(row, col int32) (safe bool) {
	if !m.inBounds(row, col) {
		return true
	}
	if (m.Solved.Valid && m.Solved.Bool) || (m.Failed.Valid && m.Failed.Bool) {
		return true
	}

	// Don't reveal already revealed cells.
	if m.Revealed[row][col] {
		return true
	}

	// Mine hit
	if m.Grid[row][col] == -1 {
		m.Revealed[row][col] = true
		m.Solved = pgtype.Bool{Bool: false, Valid: true}
		m.Failed = pgtype.Bool{Bool: true, Valid: true}
		return false
	}

	m.floodReveal(row, col)
	m.checkSolved()
	return true
}

func (m *MineSweeper) floodReveal(row, col int32) {
	// BFS flood-fill for zero-cells
	type cell struct{ r, c int32 }
	q := []cell{{row, col}}

	for len(q) > 0 {
		cur := q[0]
		q = q[1:]

		if !m.inBounds(cur.r, cur.c) {
			continue
		}
		if m.Revealed[cur.r][cur.c] {
			continue
		}
		if m.Grid[cur.r][cur.c] == -1 {
			continue
		}

		m.Revealed[cur.r][cur.c] = true

		// Only expand neighbors if this cell has 0 adjacent mines
		if m.Grid[cur.r][cur.c] != 0 {
			continue
		}

		for dr := int32(-1); dr <= 1; dr++ {
			for dc := int32(-1); dc <= 1; dc++ {
				if dr == 0 && dc == 0 {
					continue
				}
				n := cell{cur.r + dr, cur.c + dc}
				if m.inBounds(n.r, n.c) && !m.Revealed[n.r][n.c] {
					q = append(q, n)
				}
			}
		}
	}
}

func (m *MineSweeper) checkSolved() {
	if m.Failed.Valid && m.Failed.Bool {
		return
	}
	// solved if all non-mine cells are revealed
	for r := int32(0); r < m.GridSize; r++ {
		for c := int32(0); c < m.GridSize; c++ {
			if m.Grid[r][c] != -1 && !m.Revealed[r][c] {
				m.Solved = pgtype.Bool{Bool: false, Valid: true}
				return
			}
		}
	}
	m.Solved = pgtype.Bool{Bool: true, Valid: true}
}

func GetDifficultyParams(level int32) (gridSize, mineCount int32) {
	switch level {
	case 1:
		return 5, 3
	case 2:
		return 6, 5
	case 3:
		return 8, 10
	case 4:
		return 10, 20
	case 5:
		return 12, 30
	default:
		return 10 + (level-4)*2, 20 + (level-4)*10 // Scale linearly
	}

}
