package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type AnalyticsData struct {
	// Gameplay metrics
	TotalGamesStarted   int64   `json:"totalGamesStarted"`
	TotalGamesCompleted int64   `json:"totalGamesCompleted"`
	TotalMoves          int64   `json:"totalMoves"`
	AvgGameDuration     float64 `json:"avgGameDuration"`
	GamesLastHour       int64   `json:"gamesLastHour"`
	GamesLast24Hours    int64   `json:"gamesLast24Hours"`

	// Top winners
	TopWinners []WinnerStats `json:"topWinners"`

	// User metrics
	UserStats []UserMetrics `json:"userStats"`
}

type WinnerStats struct {
	Username  string `json:"username"`
	WinCount  int    `json:"winCount"`
	LastWinAt string `json:"lastWinAt"`
}

type UserMetrics struct {
	Username        string  `json:"username"`
	TotalGames      int     `json:"totalGames"`
	Wins            int     `json:"wins"`
	Losses          int     `json:"losses"`
	Draws           int     `json:"draws"`
	TotalMoves      int     `json:"totalMoves"`
	AvgGameDuration float64 `json:"avgGameDuration"`
	WinRate         float64 `json:"winRate"`
}

func HandleAnalytics(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	analytics := AnalyticsData{}

	// Get summary metrics
	rows, err := db.Query(`
		SELECT metric_name, metric_value FROM analytics_summary
		WHERE metric_name IN ('total_games_started', 'total_games_completed', 'total_moves', 
		                      'avg_game_duration', 'games_last_hour', 'games_last_24h')
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			var value float64
			if err := rows.Scan(&name, &value); err == nil {
				switch name {
				case "total_games_started":
					analytics.TotalGamesStarted = int64(value)
				case "total_games_completed":
					analytics.TotalGamesCompleted = int64(value)
				case "total_moves":
					analytics.TotalMoves = int64(value)
				case "avg_game_duration":
					analytics.AvgGameDuration = value
				case "games_last_hour":
					analytics.GamesLastHour = int64(value)
				case "games_last_24h":
					analytics.GamesLast24Hours = int64(value)
				}
			}
		}
	}

	// Get top 10 winners
	winnerRows, err := db.Query(`
		SELECT username, win_count, last_win_at FROM winner_frequency
		ORDER BY win_count DESC
		LIMIT 10
	`)
	if err == nil {
		defer winnerRows.Close()
		for winnerRows.Next() {
			var winner WinnerStats
			var lastWin sql.NullTime
			if err := winnerRows.Scan(&winner.Username, &winner.WinCount, &lastWin); err == nil {
				if lastWin.Valid {
					winner.LastWinAt = lastWin.Time.Format("2006-01-02 15:04:05")
				}
				analytics.TopWinners = append(analytics.TopWinners, winner)
			}
		}
	}

	// Get user-specific metrics
	userRows, err := db.Query(`
		SELECT username, total_games, wins, losses, draws, total_moves, avg_game_duration
		FROM user_metrics
		WHERE total_games > 0
		ORDER BY wins DESC
		LIMIT 50
	`)
	if err == nil {
		defer userRows.Close()
		for userRows.Next() {
			var user UserMetrics
			if err := userRows.Scan(&user.Username, &user.TotalGames, &user.Wins,
				&user.Losses, &user.Draws, &user.TotalMoves, &user.AvgGameDuration); err == nil {

				// Calculate win rate
				if user.TotalGames > 0 {
					user.WinRate = float64(user.Wins) / float64(user.TotalGames) * 100
				}

				analytics.UserStats = append(analytics.UserStats, user)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}
