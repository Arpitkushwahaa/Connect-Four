package models

import (
	"time"
)

const (
	Rows    = 6
	Columns = 7
)

type Player struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	IsBot    bool   `json:"isBot"`
}

type GameState string

const (
	GameStateWaiting  GameState = "waiting"
	GameStatePlaying  GameState = "playing"
	GameStateFinished GameState = "finished"
)

type Game struct {
	ID          string     `json:"id"`
	Player1     *Player    `json:"player1"`
	Player2     *Player    `json:"player2"`
	Board       [][]int    `json:"board"` // 0 = empty, 1 = player1, 2 = player2
	CurrentTurn int        `json:"currentTurn"`
	State       GameState  `json:"state"`
	Winner      *Player    `json:"winner,omitempty"`
	WinningLine [][]int    `json:"winningLine,omitempty"`
	StartTime   time.Time  `json:"startTime"`
	EndTime     *time.Time `json:"endTime,omitempty"`
	LastMoveCol *int       `json:"lastMoveCol,omitempty"`
	LastMoveRow *int       `json:"lastMoveRow,omitempty"`
}

type Move struct {
	Column    int       `json:"column"`
	PlayerID  string    `json:"playerId"`
	Timestamp time.Time `json:"timestamp"`
}

type GameResult struct {
	GameID       string    `json:"gameId"`
	Player1      string    `json:"player1"`
	Player2      string    `json:"player2"`
	Winner       string    `json:"winner"`
	Duration     int       `json:"duration"` // in seconds
	TotalMoves   int       `json:"totalMoves"`
	CompletedAt  time.Time `json:"completedAt"`
	Player1IsBot bool      `json:"player1IsBot"`
	Player2IsBot bool      `json:"player2IsBot"`
}

type LeaderboardEntry struct {
	Username string `json:"username"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
	Draws    int    `json:"draws"`
}

func NewGame(player1 *Player) *Game {
	board := make([][]int, Rows)
	for i := range board {
		board[i] = make([]int, Columns)
	}

	return &Game{
		ID:          generateGameID(),
		Player1:     player1,
		Board:       board,
		CurrentTurn: 1,
		State:       GameStateWaiting,
		StartTime:   time.Now(),
	}
}

func generateGameID() string {
	return time.Now().Format("20060102150405") + randString(6)
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
