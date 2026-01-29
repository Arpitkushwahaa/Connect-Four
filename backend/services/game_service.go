package services

import (
	"connect-four-backend/models"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type GameService struct {
	db           *sql.DB
	games        map[string]*models.Game
	playerGames  map[string]string // playerID -> gameID
	gamesMutex   sync.RWMutex
	kafka        *KafkaProducer
	disconnected map[string]time.Time // playerID -> disconnect time
}

func NewGameService(db *sql.DB, kafkaBroker string) *GameService {
	kafka := NewKafkaProducer(kafkaBroker)

	gs := &GameService{
		db:           db,
		games:        make(map[string]*models.Game),
		playerGames:  make(map[string]string),
		kafka:        kafka,
		disconnected: make(map[string]time.Time),
	}

	// Start cleanup goroutine for disconnected players
	go gs.cleanupDisconnectedPlayers()

	return gs
}

func (gs *GameService) CreateGame(player *models.Player) *models.Game {
	gs.gamesMutex.Lock()
	defer gs.gamesMutex.Unlock()

	game := models.NewGame(player)
	gs.games[game.ID] = game
	gs.playerGames[player.ID] = game.ID

	return game
}

func (gs *GameService) JoinGame(game *models.Game, player *models.Player) {
	gs.gamesMutex.Lock()
	defer gs.gamesMutex.Unlock()

	game.Player2 = player
	game.State = models.GameStatePlaying
	gs.playerGames[player.ID] = game.ID

	// Send Kafka event
	gs.kafka.SendEvent(GameStartEvent{
		Type:       "game_start",
		GameID:     game.ID,
		Player1:    game.Player1.Username,
		Player2:    game.Player2.Username,
		Player1Bot: game.Player1.IsBot,
		Player2Bot: game.Player2.IsBot,
		Timestamp:  time.Now(),
	})
}

func (gs *GameService) MakeMove(gameID string, playerID string, column int) error {
	gs.gamesMutex.Lock()
	defer gs.gamesMutex.Unlock()

	game, exists := gs.games[gameID]
	if !exists {
		return nil
	}

	// Determine player number
	playerNum := 1
	playerName := game.Player1.Username
	if playerID == game.Player2.ID {
		playerNum = 2
		playerName = game.Player2.Username
	}

	// Check if it's the player's turn
	if game.CurrentTurn != playerNum {
		return nil
	}

	// Make the move
	row, err := MakeMove(game, column, playerNum)
	if err != nil {
		return err
	}

	// Send move event to Kafka
	gs.kafka.SendEvent(GameMoveEvent{
		Type:      "game_move",
		GameID:    gameID,
		Player:    playerName,
		Column:    column,
		Row:       row,
		Timestamp: time.Now(),
	})

	// Check for winner
	hasWon, winner, winningLine := CheckWinner(game)
	if hasWon {
		game.State = models.GameStateFinished
		game.Winner = winner
		game.WinningLine = winningLine
		endTime := time.Now()
		game.EndTime = &endTime
		gs.saveGameResult(game, "win")
		return nil
	}

	// Check for draw
	if IsBoardFull(game) {
		game.State = models.GameStateFinished
		endTime := time.Now()
		game.EndTime = &endTime
		gs.saveGameResult(game, "draw")
		return nil
	}

	// Switch turn
	game.CurrentTurn = 3 - game.CurrentTurn

	return nil
}

func (gs *GameService) GetGame(gameID string) *models.Game {
	gs.gamesMutex.RLock()
	defer gs.gamesMutex.RUnlock()
	return gs.games[gameID]
}

func (gs *GameService) GetPlayerGame(playerID string) *models.Game {
	gs.gamesMutex.RLock()
	defer gs.gamesMutex.RUnlock()

	gameID, exists := gs.playerGames[playerID]
	if !exists {
		return nil
	}

	return gs.games[gameID]
}

func (gs *GameService) RemoveGame(gameID string) {
	gs.gamesMutex.Lock()
	defer gs.gamesMutex.Unlock()

	game := gs.games[gameID]
	if game != nil {
		delete(gs.playerGames, game.Player1.ID)
		if game.Player2 != nil {
			delete(gs.playerGames, game.Player2.ID)
		}
	}
	delete(gs.games, gameID)
}

func (gs *GameService) MarkPlayerDisconnected(playerID string) {
	gs.gamesMutex.Lock()
	defer gs.gamesMutex.Unlock()
	gs.disconnected[playerID] = time.Now()
}

func (gs *GameService) ReconnectPlayer(playerID string) {
	gs.gamesMutex.Lock()
	defer gs.gamesMutex.Unlock()
	delete(gs.disconnected, playerID)
}

func (gs *GameService) cleanupDisconnectedPlayers() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		gs.gamesMutex.Lock()
		now := time.Now()

		for playerID, disconnectTime := range gs.disconnected {
			if now.Sub(disconnectTime) > 30*time.Second {
				// Player didn't reconnect in time
				gameID, exists := gs.playerGames[playerID]
				if exists {
					game := gs.games[gameID]
					if game != nil && game.State == models.GameStatePlaying {
						// Forfeit the game
						var winner *models.Player
						if game.Player1.ID == playerID {
							winner = game.Player2
						} else {
							winner = game.Player1
						}

						game.State = models.GameStateFinished
						game.Winner = winner
						endTime := now
						game.EndTime = &endTime

						gs.saveGameResult(game, "forfeit")
					}
				}

				delete(gs.disconnected, playerID)
			}
		}

		gs.gamesMutex.Unlock()
	}
}

func (gs *GameService) saveGameResult(game *models.Game, reason string) {
	duration := int(game.EndTime.Sub(game.StartTime).Seconds())

	// Count total moves
	totalMoves := 0
	for r := 0; r < models.Rows; r++ {
		for c := 0; c < models.Columns; c++ {
			if game.Board[r][c] != 0 {
				totalMoves++
			}
		}
	}

	// Save to database
	winnerName := ""
	if game.Winner != nil {
		winnerName = game.Winner.Username
	}

	_, err := gs.db.Exec(`
		INSERT INTO games (id, player1, player2, winner, duration, total_moves, completed_at, player1_is_bot, player2_is_bot)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, game.ID, game.Player1.Username, game.Player2.Username, winnerName, duration, totalMoves, time.Now(),
		game.Player1.IsBot, game.Player2.IsBot)

	if err != nil {
		log.Printf("Failed to save game result: %v", err)
	}

	// Update leaderboard
	if game.Winner != nil {
		loserName := game.Player1.Username
		if game.Winner.ID == game.Player1.ID {
			loserName = game.Player2.Username
		}

		if !game.Winner.IsBot {
			gs.updateLeaderboard(winnerName, "win")
		}

		// Find loser
		var loser *models.Player
		if game.Winner.ID == game.Player1.ID {
			loser = game.Player2
		} else {
			loser = game.Player1
		}

		if !loser.IsBot {
			gs.updateLeaderboard(loserName, "loss")
		}
	} else {
		// Draw
		if !game.Player1.IsBot {
			gs.updateLeaderboard(game.Player1.Username, "draw")
		}
		if !game.Player2.IsBot {
			gs.updateLeaderboard(game.Player2.Username, "draw")
		}
	}

	// Send Kafka event
	gs.kafka.SendEvent(GameEndEvent{
		Type:       "game_end",
		GameID:     game.ID,
		Winner:     winnerName,
		Duration:   duration,
		TotalMoves: totalMoves,
		Reason:     reason,
		Timestamp:  time.Now(),
	})
}

func (gs *GameService) updateLeaderboard(username string, result string) {
	var wins, losses, draws int

	// Get current stats
	err := gs.db.QueryRow("SELECT wins, losses, draws FROM leaderboard WHERE username = $1", username).
		Scan(&wins, &losses, &draws)

	if err == sql.ErrNoRows {
		// New player
		wins, losses, draws = 0, 0, 0
	}

	// Update based on result
	switch result {
	case "win":
		wins++
	case "loss":
		losses++
	case "draw":
		draws++
	}

	// Upsert
	_, err = gs.db.Exec(`
		INSERT INTO leaderboard (username, wins, losses, draws)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (username) DO UPDATE
		SET wins = $2, losses = $3, draws = $4
	`, username, wins, losses, draws)

	if err != nil {
		log.Printf("Failed to update leaderboard: %v", err)
	}
}

func (gs *GameService) GetLeaderboard() ([]models.LeaderboardEntry, error) {
	rows, err := gs.db.Query(`
		SELECT username, wins, losses, draws
		FROM leaderboard
		ORDER BY wins DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaderboard []models.LeaderboardEntry
	for rows.Next() {
		var entry models.LeaderboardEntry
		if err := rows.Scan(&entry.Username, &entry.Wins, &entry.Losses, &entry.Draws); err != nil {
			continue
		}
		leaderboard = append(leaderboard, entry)
	}

	return leaderboard, nil
}

func GeneratePlayerID() string {
	return uuid.New().String()
}
