package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type Event struct {
	Type      string    `json:"type"`
	GameID    string    `json:"gameId,omitempty"`
	Player    string    `json:"player,omitempty"`
	Player1   string    `json:"player1,omitempty"`
	Player2   string    `json:"player2,omitempty"`
	Winner    string    `json:"winner,omitempty"`
	Duration  int       `json:"duration,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type Analytics struct {
	db *sql.DB
}

func main() {
	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/connectfour?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Wait for database
	for i := 0; i < 30; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		log.Println("Waiting for database...")
		time.Sleep(time.Second)
	}

	// Initialize analytics tables
	if err := initAnalyticsTables(db); err != nil {
		log.Fatal("Failed to initialize analytics tables:", err)
	}

	analytics := &Analytics{db: db}

	// Set up Kafka consumer
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	brokerList := strings.Split(kafkaBrokers, ",")

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "game-events"
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "analytics-consumer"
	}

	// Configure Kafka reader
	readerConfig := kafka.ReaderConfig{
		Brokers:  brokerList,
		Topic:    kafkaTopic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	}

	// Configure SASL authentication if credentials are provided
	username := os.Getenv("KAFKA_USERNAME")
	password := os.Getenv("KAFKA_PASSWORD")
	mechanism := os.Getenv("KAFKA_SASL_MECHANISM")

	if username != "" && password != "" {
		var scramMechanism sasl.Mechanism
		var err error

		switch mechanism {
		case "SCRAM-SHA-256":
			scramMechanism, err = scram.Mechanism(scram.SHA256, username, password)
		case "SCRAM-SHA-512":
			scramMechanism, err = scram.Mechanism(scram.SHA512, username, password)
		default:
			scramMechanism, err = scram.Mechanism(scram.SHA512, username, password)
		}

		if err != nil {
			log.Fatal("Failed to create SASL mechanism:", err)
		}

		dialer := &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: scramMechanism,
			TLS:           &tls.Config{},
		}
		readerConfig.Dialer = dialer
		log.Println("Kafka consumer configured with SASL authentication")
	}

	reader := kafka.NewReader(readerConfig)
	defer reader.Close()

	log.Println("Analytics service started, listening for events...")

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Consume messages
	for {
		select {
		case <-ctx.Done():
			return
		default:
			m, err := reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error fetching message: %v", err)
				continue
			}

			var event Event
			if err := json.Unmarshal(m.Value, &event); err != nil {
				log.Printf("Error unmarshaling event: %v", err)
				reader.CommitMessages(ctx, m)
				continue
			}

			analytics.processEvent(event)
			reader.CommitMessages(ctx, m)
		}
	}
}

func initAnalyticsTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS game_events (
		id SERIAL PRIMARY KEY,
		event_type VARCHAR(50) NOT NULL,
		game_id VARCHAR(255),
		player VARCHAR(255),
		data JSONB,
		timestamp TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS analytics_summary (
		id SERIAL PRIMARY KEY,
		metric_name VARCHAR(100) NOT NULL UNIQUE,
		metric_value NUMERIC NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_game_events_type ON game_events(event_type);
	CREATE INDEX IF NOT EXISTS idx_game_events_timestamp ON game_events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_game_events_game_id ON game_events(game_id);
	`

	_, err := db.Exec(schema)
	return err
}

func (a *Analytics) processEvent(event Event) {
	// Store raw event
	data, _ := json.Marshal(event)
	_, err := a.db.Exec(`
		INSERT INTO game_events (event_type, game_id, player, data, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`, event.Type, event.GameID, event.Player, data, event.Timestamp)

	if err != nil {
		log.Printf("Failed to store event: %v", err)
	}

	// Process specific event types
	switch event.Type {
	case "game_start":
		a.incrementMetric("total_games_started")
		log.Printf("Game started: %s (Player1: %s, Player2: %s)", event.GameID, event.Player1, event.Player2)

	case "game_move":
		a.incrementMetric("total_moves")
		log.Printf("Move made in game %s by %s", event.GameID, event.Player)

	case "game_end":
		a.incrementMetric("total_games_completed")
		a.updateAverageGameDuration(event.Duration)
		log.Printf("Game ended: %s (Winner: %s, Duration: %ds)", event.GameID, event.Winner, event.Duration)

		if event.Winner != "" {
			a.trackWinner(event.Winner)
		}
	}

	// Calculate and update analytics
	a.updateHourlyGames()
}

func (a *Analytics) incrementMetric(metricName string) {
	_, err := a.db.Exec(`
		INSERT INTO analytics_summary (metric_name, metric_value, updated_at)
		VALUES ($1, 1, $2)
		ON CONFLICT (metric_name) DO UPDATE
		SET metric_value = analytics_summary.metric_value + 1,
		    updated_at = $2
	`, metricName, time.Now())

	if err != nil {
		log.Printf("Failed to increment metric %s: %v", metricName, err)
	}
}

func (a *Analytics) updateAverageGameDuration(duration int) {
	var totalGames, totalDuration int64

	// Get current values
	err := a.db.QueryRow(`
		SELECT metric_value FROM analytics_summary WHERE metric_name = 'total_games_completed'
	`).Scan(&totalGames)

	if err != nil {
		totalGames = 1
	}

	err = a.db.QueryRow(`
		SELECT metric_value FROM analytics_summary WHERE metric_name = 'total_duration'
	`).Scan(&totalDuration)

	if err != nil {
		totalDuration = 0
	}

	// Update total duration
	newTotalDuration := totalDuration + int64(duration)
	_, err = a.db.Exec(`
		INSERT INTO analytics_summary (metric_name, metric_value, updated_at)
		VALUES ('total_duration', $1, $2)
		ON CONFLICT (metric_name) DO UPDATE
		SET metric_value = $1, updated_at = $2
	`, newTotalDuration, time.Now())

	// Calculate and update average
	if totalGames > 0 {
		avgDuration := newTotalDuration / totalGames
		_, err = a.db.Exec(`
			INSERT INTO analytics_summary (metric_name, metric_value, updated_at)
			VALUES ('avg_game_duration', $1, $2)
			ON CONFLICT (metric_name) DO UPDATE
			SET metric_value = $1, updated_at = $2
		`, avgDuration, time.Now())
	}
}

func (a *Analytics) trackWinner(winner string) {
	// This could track most frequent winners
	log.Printf("Winner tracked: %s", winner)
}

func (a *Analytics) updateHourlyGames() {
	// Count games in the last hour
	var count int
	err := a.db.QueryRow(`
		SELECT COUNT(*) FROM game_events
		WHERE event_type = 'game_end'
		AND timestamp > NOW() - INTERVAL '1 hour'
	`).Scan(&count)

	if err == nil {
		_, _ = a.db.Exec(`
			INSERT INTO analytics_summary (metric_name, metric_value, updated_at)
			VALUES ('games_last_hour', $1, $2)
			ON CONFLICT (metric_name) DO UPDATE
			SET metric_value = $1, updated_at = $2
		`, count, time.Now())
	}
}
