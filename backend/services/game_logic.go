package services

import (
	"connect-four-backend/models"
	"errors"
)

// MakeMove attempts to make a move in the specified column
func MakeMove(game *models.Game, column int, playerNum int) (int, error) {
	if column < 0 || column >= models.Columns {
		return -1, errors.New("invalid column")
	}

	// Find the lowest empty row in the column
	row := -1
	for r := models.Rows - 1; r >= 0; r-- {
		if game.Board[r][column] == 0 {
			row = r
			break
		}
	}

	if row == -1 {
		return -1, errors.New("column is full")
	}

	// Place the disc
	game.Board[row][column] = playerNum
	game.LastMoveCol = &column
	game.LastMoveRow = &row

	return row, nil
}

// CheckWinner checks if there's a winner after the last move
func CheckWinner(game *models.Game) (bool, *models.Player, [][]int) {
	if game.LastMoveRow == nil || game.LastMoveCol == nil {
		return false, nil, nil
	}

	row := *game.LastMoveRow
	col := *game.LastMoveCol
	playerNum := game.Board[row][col]

	// Check horizontal
	if line := checkDirection(game.Board, row, col, 0, 1, playerNum); line != nil {
		winner := game.Player1
		if playerNum == 2 {
			winner = game.Player2
		}
		return true, winner, line
	}

	// Check vertical
	if line := checkDirection(game.Board, row, col, 1, 0, playerNum); line != nil {
		winner := game.Player1
		if playerNum == 2 {
			winner = game.Player2
		}
		return true, winner, line
	}

	// Check diagonal (down-right)
	if line := checkDirection(game.Board, row, col, 1, 1, playerNum); line != nil {
		winner := game.Player1
		if playerNum == 2 {
			winner = game.Player2
		}
		return true, winner, line
	}

	// Check diagonal (down-left)
	if line := checkDirection(game.Board, row, col, 1, -1, playerNum); line != nil {
		winner := game.Player1
		if playerNum == 2 {
			winner = game.Player2
		}
		return true, winner, line
	}

	return false, nil, nil
}

// checkDirection checks for 4 in a row in a specific direction
func checkDirection(board [][]int, row, col, dRow, dCol, playerNum int) [][]int {
	positions := [][]int{{row, col}}

	// Check forward
	r, c := row+dRow, col+dCol
	for len(positions) < 4 && r >= 0 && r < models.Rows && c >= 0 && c < models.Columns {
		if board[r][c] == playerNum {
			positions = append(positions, []int{r, c})
			r += dRow
			c += dCol
		} else {
			break
		}
	}

	// Check backward
	r, c = row-dRow, col-dCol
	for len(positions) < 4 && r >= 0 && r < models.Rows && c >= 0 && c < models.Columns {
		if board[r][c] == playerNum {
			positions = append([][]int{{r, c}}, positions...)
			r -= dRow
			c -= dCol
		} else {
			break
		}
	}

	if len(positions) >= 4 {
		return positions[:4]
	}

	return nil
}

// IsBoardFull checks if the board is completely full
func IsBoardFull(game *models.Game) bool {
	for c := 0; c < models.Columns; c++ {
		if game.Board[0][c] == 0 {
			return false
		}
	}
	return true
}

// IsValidMove checks if a move is valid
func IsValidMove(game *models.Game, column int) bool {
	if column < 0 || column >= models.Columns {
		return false
	}
	return game.Board[0][column] == 0
}
