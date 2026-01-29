package main

import (
	"connect-four-backend/handlers"
	"connect-four-backend/services"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func main() {
	// Initialize database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/connectfour?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Wait for database to be ready
	for i := 0; i < 30; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		log.Println("Waiting for database...")
		time.Sleep(time.Second)
	}

	// Initialize database schema
	if err := services.InitDB(db); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize Kafka
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	// Initialize services
	gameService := services.NewGameService(db, kafkaBrokers)
	matchmakingService := services.NewMatchmakingService(gameService)

	// Start matchmaking loop
	go matchmakingService.StartMatchmaking()

	// Set up HTTP handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleWebSocket(w, r, gameService, matchmakingService)
	})
	mux.HandleFunc("/api/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleLeaderboard(w, r, gameService)
	})
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Enable CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Server failed:", err)
	}
}
