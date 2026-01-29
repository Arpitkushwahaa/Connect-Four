package services

import (
	"connect-four-backend/models"
	"math/rand"
	"time"
)

type Bot struct {
	PlayerNum int
}

func NewBot(playerNum int) *Bot {
	return &Bot{PlayerNum: playerNum}
}

// GetMove returns the bot's next move using strategic AI
func (b *Bot) GetMove(game *models.Game) int {
	opponentNum := 1
	if b.PlayerNum == 1 {
		opponentNum = 2
	}

	// Priority 1: Win if possible
	if col := b.findWinningMove(game, b.PlayerNum); col != -1 {
		return col
	}

	// Priority 2: Block opponent's winning move
	if col := b.findWinningMove(game, opponentNum); col != -1 {
		return col
	}

	// Priority 3: Create a threat (setup for next turn win)
	if col := b.findThreatMove(game, b.PlayerNum); col != -1 {
		return col
	}

	// Priority 4: Block opponent's threat
	if col := b.findThreatMove(game, opponentNum); col != -1 {
		return col
	}

	// Priority 5: Take center columns (strategic advantage)
	centerCols := []int{3, 2, 4, 1, 5, 0, 6}
	for _, col := range centerCols {
		if IsValidMove(game, col) {
			return col
		}
	}

	// Fallback: random valid move
	validCols := []int{}
	for c := 0; c < models.Columns; c++ {
		if IsValidMove(game, c) {
			validCols = append(validCols, c)
		}
	}

	if len(validCols) > 0 {
		rand.Seed(time.Now().UnixNano())
		return validCols[rand.Intn(len(validCols))]
	}

	return -1
}

// findWinningMove finds a column that would result in a win for the player
func (b *Bot) findWinningMove(game *models.Game, playerNum int) int {
	for col := 0; col < models.Columns; col++ {
		if !IsValidMove(game, col) {
			continue
		}

		// Simulate the move
		tempGame := b.copyGame(game)
		row, err := MakeMove(tempGame, col, playerNum)
		if err != nil {
			continue
		}

		// Check if this move wins
		if hasWon, _, _ := CheckWinner(tempGame); hasWon {
			return col
		}

		// Undo the move
		tempGame.Board[row][col] = 0
	}

	return -1
}

// findThreatMove finds a move that creates multiple winning opportunities
func (b *Bot) findThreatMove(game *models.Game, playerNum int) int {
	for col := 0; col < models.Columns; col++ {
		if !IsValidMove(game, col) {
			continue
		}

		// Simulate the move
		tempGame := b.copyGame(game)
		row, err := MakeMove(tempGame, col, playerNum)
		if err != nil {
			continue
		}

		// Count potential wins after this move
		winCount := 0
		for nextCol := 0; nextCol < models.Columns; nextCol++ {
			if !IsValidMove(tempGame, nextCol) {
				continue
			}

			tempGame2 := b.copyGame(tempGame)
			_, err := MakeMove(tempGame2, nextCol, playerNum)
			if err != nil {
				continue
			}

			if hasWon, _, _ := CheckWinner(tempGame2); hasWon {
				winCount++
			}
		}

		// If this move creates 2+ winning opportunities, it's a threat
		if winCount >= 2 {
			return col
		}

		// Undo the move
		tempGame.Board[row][col] = 0
	}

	return -1
}

// copyGame creates a deep copy of the game for simulation
func (b *Bot) copyGame(game *models.Game) *models.Game {
	newGame := &models.Game{
		ID:          game.ID,
		Player1:     game.Player1,
		Player2:     game.Player2,
		CurrentTurn: game.CurrentTurn,
		State:       game.State,
		Board:       make([][]int, models.Rows),
	}

	for i := range game.Board {
		newGame.Board[i] = make([]int, models.Columns)
		copy(newGame.Board[i], game.Board[i])
	}

	if game.LastMoveCol != nil {
		col := *game.LastMoveCol
		newGame.LastMoveCol = &col
	}
	if game.LastMoveRow != nil {
		row := *game.LastMoveRow
		newGame.LastMoveRow = &row
	}

	return newGame
}
