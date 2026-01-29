package services

import (
	"database/sql"
)

func InitDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS games (
		id VARCHAR(255) PRIMARY KEY,
		player1 VARCHAR(255) NOT NULL,
		player2 VARCHAR(255) NOT NULL,
		winner VARCHAR(255),
		duration INTEGER NOT NULL,
		total_moves INTEGER NOT NULL,
		completed_at TIMESTAMP NOT NULL,
		player1_is_bot BOOLEAN NOT NULL DEFAULT FALSE,
		player2_is_bot BOOLEAN NOT NULL DEFAULT FALSE
	);

	CREATE TABLE IF NOT EXISTS leaderboard (
		username VARCHAR(255) PRIMARY KEY,
		wins INTEGER NOT NULL DEFAULT 0,
		losses INTEGER NOT NULL DEFAULT 0,
		draws INTEGER NOT NULL DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_games_completed_at ON games(completed_at);
	CREATE INDEX IF NOT EXISTS idx_leaderboard_wins ON leaderboard(wins DESC);
	`

	_, err := db.Exec(schema)
	return err
}
