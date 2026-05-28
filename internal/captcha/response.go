package captcha

import (
	db "captcha_sweeper/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	hiddenCellValue int32 = -2
	mineCellValue   int32 = -1
)

type CaptchaResponse struct {
	CaptchaID       string    `json:"captcha_id"`
	Solved          bool      `json:"solved"`
	GameOver        bool      `json:"game_over"`
	MineHit         bool      `json:"mine_hit"`
	GridSize        int32     `json:"grid_size"`
	GridRevealed    [][]int32 `json:"grid_revealed"`
	DifficultyLevel int32     `json:"difficulty_level"`
	Status          string    `json:"status"`
}

func BuildCaptchaResponse(c db.Captcha) CaptchaResponse {
	solved := c.Solved.Valid && c.Solved.Bool
	mineHit := mineHit(c)
	gameOver := solved || mineHit

	return CaptchaResponse{
		CaptchaID:       c.CaptchaID.String(),
		Solved:          solved,
		GameOver:        gameOver,
		MineHit:         mineHit,
		GridSize:        c.GridSize,
		GridRevealed:    maskGrid(c.Grid, c.Revealed),
		DifficultyLevel: difficultyValue(c.DifficultyLevel),
		Status:          stateOfGame(c),
	}
}

func difficultyValue(v pgtype.Int4) int32 {
	if v.Valid {
		return v.Int32
	}
	return 0
}

func stateOfGame(c db.Captcha) string {
	if c.Solved.Valid && c.Solved.Bool {
		return "solved"
	}
	if mineHit(c) {
		return "mine_hit"
	}
	return "unsolved"
}

func mineHit(c db.Captcha) bool {
	if c.Failed.Valid {
		return c.Failed.Bool
	}
	if len(c.Grid) == 0 || len(c.Revealed) == 0 {
		return false
	}
	for i := range c.Grid {
		if i >= len(c.Revealed) {
			continue
		}
		for j := range c.Grid[i] {
			if j >= len(c.Revealed[i]) {
				continue
			}
			if c.Revealed[i][j] && c.Grid[i][j] == mineCellValue {
				return true
			}
		}
	}
	return false
}

func maskGrid(grid [][]int32, revealed [][]bool) [][]int32 {
	if len(grid) == 0 {
		return nil
	}

	out := make([][]int32, len(grid))
	for i := range grid {
		out[i] = make([]int32, len(grid[i]))
		for j := range grid[i] {
			if i < len(revealed) && j < len(revealed[i]) && revealed[i][j] {
				out[i][j] = grid[i][j]
			} else {
				out[i][j] = hiddenCellValue
			}
		}
	}

	return out
}
