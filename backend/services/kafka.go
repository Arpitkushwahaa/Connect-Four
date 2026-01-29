package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(broker string) *KafkaProducer {
	// Check if SASL authentication is needed
	username := os.Getenv("KAFKA_USERNAME")
	password := os.Getenv("KAFKA_PASSWORD")
	mechanism := os.Getenv("KAFKA_SASL_MECHANISM")

	writerConfig := kafka.WriterConfig{
		Brokers:      []string{broker},
		Topic:        "game-events",
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	// Configure SASL authentication if credentials are provided
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
			log.Printf("Failed to create SASL mechanism: %v", err)
		} else {
			dialer := &kafka.Dialer{
				Timeout:       10 * time.Second,
				DualStack:     true,
				SASLMechanism: scramMechanism,
				TLS:           &tls.Config{},
			}
			writerConfig.Dialer = dialer
			log.Println("Kafka producer configured with SASL authentication")
		}
	}

	writer := kafka.NewWriter(writerConfig)
	return &KafkaProducer{writer: writer}
}

func (kp *KafkaProducer) SendEvent(event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Value: data,
		Time:  time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = kp.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("Failed to send Kafka event: %v", err)
		return err
	}

	return nil
}

func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}

// Event types
type GameStartEvent struct {
	Type       string    `json:"type"`
	GameID     string    `json:"gameId"`
	Player1    string    `json:"player1"`
	Player2    string    `json:"player2"`
	Player1Bot bool      `json:"player1Bot"`
	Player2Bot bool      `json:"player2Bot"`
	Timestamp  time.Time `json:"timestamp"`
}

type GameMoveEvent struct {
	Type      string    `json:"type"`
	GameID    string    `json:"gameId"`
	Player    string    `json:"player"`
	Column    int       `json:"column"`
	Row       int       `json:"row"`
	Timestamp time.Time `json:"timestamp"`
}

type GameEndEvent struct {
	Type       string    `json:"type"`
	GameID     string    `json:"gameId"`
	Winner     string    `json:"winner"`
	Duration   int       `json:"duration"`
	TotalMoves int       `json:"totalMoves"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
}

type PlayerJoinEvent struct {
	Type      string    `json:"type"`
	Username  string    `json:"username"`
	Timestamp time.Time `json:"timestamp"`
}

type PlayerDisconnectEvent struct {
	Type      string    `json:"type"`
	GameID    string    `json:"gameId"`
	Username  string    `json:"username"`
	Timestamp time.Time `json:"timestamp"`
}
