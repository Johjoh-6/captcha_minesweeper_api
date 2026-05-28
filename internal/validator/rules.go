package validator

const (
	MinDifficulty = 1
	MaxDifficulty = 10
	MaxGridSize   = 20
)

// ValidateCaptchaMovementInput validates captcha movement (row/col) input.
// Returns errors for invalid coordinates.
func ValidateCaptchaMovementInput(v *Validator, row, col int32, gridSize int32) {
	v.CheckField(row >= 0, "row", "row must be >= 0")
	v.CheckField(row < gridSize, "row", "row must be < grid_size")

	v.CheckField(col >= 0, "col", "col must be >= 0")
	v.CheckField(col < gridSize, "col", "col must be < grid_size")
}

// ValidateDifficultyInput validates difficulty level input.
// Returns an error if the level is outside the valid range.
func ValidateDifficultyInput(v *Validator, level int32) {
	v.CheckField(Between(level, int32(MinDifficulty), int32(MaxDifficulty)), "difficulty_level",
		"difficulty_level must be between 1 and 10")
}
