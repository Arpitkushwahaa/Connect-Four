package services

import (
	"connect-four-backend/models"
	"log"
	"sync"
	"time"
)

type WaitingPlayer struct {
	Player    *models.Player
	Timestamp time.Time
}

type MatchmakingService struct {
	queue       []*WaitingPlayer
	queueMutex  sync.Mutex
	gameService *GameService
}

func NewMatchmakingService(gameService *GameService) *MatchmakingService {
	return &MatchmakingService{
		queue:       make([]*WaitingPlayer, 0),
		gameService: gameService,
	}
}

func (ms *MatchmakingService) AddToQueue(player *models.Player) {
	ms.queueMutex.Lock()
	defer ms.queueMutex.Unlock()

	ms.queue = append(ms.queue, &WaitingPlayer{
		Player:    player,
		Timestamp: time.Now(),
	})

	log.Printf("Player %s added to queue. Queue size: %d", player.Username, len(ms.queue))
}

func (ms *MatchmakingService) RemoveFromQueue(playerID string) {
	ms.queueMutex.Lock()
	defer ms.queueMutex.Unlock()

	for i, wp := range ms.queue {
		if wp.Player.ID == playerID {
			ms.queue = append(ms.queue[:i], ms.queue[i+1:]...)
			log.Printf("Player removed from queue. Queue size: %d", len(ms.queue))
			return
		}
	}
}

func (ms *MatchmakingService) StartMatchmaking() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ms.processQueue()
	}
}

func (ms *MatchmakingService) processQueue() {
	ms.queueMutex.Lock()
	defer ms.queueMutex.Unlock()

	if len(ms.queue) == 0 {
		return
	}

	now := time.Now()
	processed := make(map[int]bool)

	// Try to match players
	for i := 0; i < len(ms.queue); i++ {
		if processed[i] {
			continue
		}

		wp1 := ms.queue[i]

		// Check if player has been waiting for more than 10 seconds
		if now.Sub(wp1.Timestamp) > 10*time.Second {
			// Match with bot
			log.Printf("Matching %s with bot (waited %v)", wp1.Player.Username, now.Sub(wp1.Timestamp))

			botPlayer := &models.Player{
				ID:       GeneratePlayerID(),
				Username: "Bot",
				IsBot:    true,
			}

			game := ms.gameService.CreateGame(wp1.Player)
			ms.gameService.JoinGame(game, botPlayer)

			processed[i] = true
			continue
		}

		// Try to match with another player
		for j := i + 1; j < len(ms.queue); j++ {
			if processed[j] {
				continue
			}

			wp2 := ms.queue[j]

			log.Printf("Matching %s with %s", wp1.Player.Username, wp2.Player.Username)

			game := ms.gameService.CreateGame(wp1.Player)
			ms.gameService.JoinGame(game, wp2.Player)

			processed[i] = true
			processed[j] = true
			break
		}
	}

	// Remove processed players from queue
	newQueue := make([]*WaitingPlayer, 0)
	for i, wp := range ms.queue {
		if !processed[i] {
			newQueue = append(newQueue, wp)
		}
	}
	ms.queue = newQueue
}
