package handlers

import (
	"connect-four-backend/models"
	"connect-four-backend/services"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn        *websocket.Conn
	player      *models.Player
	gameID      string
	send        chan []byte
	service     *services.GameService
	matchmaking *services.MatchmakingService
}

var (
	clients      = make(map[string]*Client) // playerID -> client
	clientsMutex sync.RWMutex
)

func HandleWebSocket(w http.ResponseWriter, r *http.Request, gameService *services.GameService, matchmakingService *services.MatchmakingService) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		conn:        conn,
		send:        make(chan []byte, 256),
		service:     gameService,
		matchmaking: matchmakingService,
	}

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		c.handleDisconnect()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg models.WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}

		c.handleMessage(msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg models.WSMessage) {
	switch msg.Type {
	case models.MsgTypeJoinQueue:
		c.handleJoinQueue(msg.Payload)
	case models.MsgTypeMove:
		c.handleMove(msg.Payload)
	case models.MsgTypeReconnect:
		c.handleReconnect(msg.Payload)
	}
}

func (c *Client) handleJoinQueue(payload interface{}) {
	data, _ := json.Marshal(payload)
	var joinData models.JoinQueuePayload
	if err := json.Unmarshal(data, &joinData); err != nil {
		c.sendError("Invalid join queue data")
		return
	}

	// Check if reconnecting
	if joinData.GameID != "" {
		c.handleReconnect(payload)
		return
	}

	// Create new player
	c.player = &models.Player{
		ID:       services.GeneratePlayerID(),
		Username: joinData.Username,
		IsBot:    false,
	}

	// Register client
	clientsMutex.Lock()
	clients[c.player.ID] = c
	clientsMutex.Unlock()

	// Add to matchmaking queue
	c.matchmaking.AddToQueue(c.player)

	// Start checking for game start
	go c.waitForGameStart()
}

func (c *Client) waitForGameStart() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(15 * time.Second)

	for {
		select {
		case <-timeout:
			return
		case <-ticker.C:
			game := c.service.GetPlayerGame(c.player.ID)
			if game != nil && game.State == models.GameStatePlaying {
				c.gameID = game.ID
				c.matchmaking.RemoveFromQueue(c.player.ID)

				// Notify player
				c.sendMessage(models.WSMessage{
					Type: models.MsgTypeGameStart,
					Payload: models.GameStartPayload{
						Game:         game,
						YourPlayerID: c.player.ID,
					},
				})

				// Notify opponent if not bot
				if game.Player2 != nil && !game.Player2.IsBot {
					c.notifyOpponent(game, models.MsgTypeGameStart)
				}

				// If opponent is bot, make bot move if it's bot's turn
				if game.Player2.IsBot && game.CurrentTurn == 2 {
					time.Sleep(1 * time.Second)
					c.makeBotMove(game)
				}

				return
			}
		}
	}
}

func (c *Client) handleMove(payload interface{}) {
	if c.player == nil || c.gameID == "" {
		c.sendError("Not in a game")
		return
	}

	data, _ := json.Marshal(payload)
	var moveData models.MovePayload
	if err := json.Unmarshal(data, &moveData); err != nil {
		c.sendError("Invalid move data")
		return
	}

	game := c.service.GetGame(c.gameID)
	if game == nil {
		c.sendError("Game not found")
		return
	}

	// Check if it's the player's turn
	playerNum := 1
	if c.player.ID == game.Player2.ID {
		playerNum = 2
	}

	if game.CurrentTurn != playerNum {
		c.sendMessage(models.WSMessage{
			Type: models.MsgTypeInvalidMove,
			Payload: models.ErrorPayload{
				Message: "Not your turn",
			},
		})
		return
	}

	// Validate move
	if !services.IsValidMove(game, moveData.Column) {
		c.sendMessage(models.WSMessage{
			Type: models.MsgTypeInvalidMove,
			Payload: models.ErrorPayload{
				Message: "Invalid move",
			},
		})
		return
	}

	// Make the move
	if err := c.service.MakeMove(c.gameID, c.player.ID, moveData.Column); err != nil {
		c.sendError(err.Error())
		return
	}

	// Get updated game
	game = c.service.GetGame(c.gameID)

	// Send update to both players
	c.sendMessage(models.WSMessage{
		Type: models.MsgTypeGameUpdate,
		Payload: models.GameUpdatePayload{
			Game: game,
		},
	})
	c.notifyOpponent(game, models.MsgTypeGameUpdate)

	// Check if game is over
	if game.State == models.GameStateFinished {
		winnerName := ""
		reason := "draw"
		message := "Game ended in a draw!"

		if game.Winner != nil {
			winnerName = game.Winner.Username
			reason = "win"
			message = winnerName + " wins!"
		}

		gameOverPayload := models.GameOverPayload{
			Game:    game,
			Winner:  winnerName,
			Reason:  reason,
			Message: message,
		}

		c.sendMessage(models.WSMessage{
			Type:    models.MsgTypeGameOver,
			Payload: gameOverPayload,
		})
		c.notifyOpponent(game, models.MsgTypeGameOver)

		return
	}

	// If opponent is bot, make bot move
	if game.Player2.IsBot && game.CurrentTurn == 2 {
		time.Sleep(time.Duration(500+time.Now().UnixNano()%1000) * time.Millisecond)
		c.makeBotMove(game)
	} else if game.Player1.IsBot && game.CurrentTurn == 1 {
		time.Sleep(time.Duration(500+time.Now().UnixNano()%1000) * time.Millisecond)
		c.makeBotMove(game)
	}
}

func (c *Client) makeBotMove(game *models.Game) {
	botPlayerNum := 1
	if game.Player2.IsBot {
		botPlayerNum = 2
	}

	bot := services.NewBot(botPlayerNum)
	column := bot.GetMove(game)

	if column == -1 {
		return
	}

	botID := game.Player1.ID
	if game.Player2.IsBot {
		botID = game.Player2.ID
	}

	// Make bot move
	c.service.MakeMove(game.ID, botID, column)

	// Get updated game
	game = c.service.GetGame(game.ID)

	// Send update
	c.sendMessage(models.WSMessage{
		Type: models.MsgTypeGameUpdate,
		Payload: models.GameUpdatePayload{
			Game:    game,
			Message: "Bot made a move",
		},
	})

	// Check if game is over
	if game.State == models.GameStateFinished {
		winnerName := ""
		reason := "draw"
		message := "Game ended in a draw!"

		if game.Winner != nil {
			winnerName = game.Winner.Username
			reason = "win"
			message = winnerName + " wins!"
		}

		c.sendMessage(models.WSMessage{
			Type: models.MsgTypeGameOver,
			Payload: models.GameOverPayload{
				Game:    game,
				Winner:  winnerName,
				Reason:  reason,
				Message: message,
			},
		})
	}
}

func (c *Client) handleReconnect(payload interface{}) {
	data, _ := json.Marshal(payload)
	var reconnectData models.JoinQueuePayload
	if err := json.Unmarshal(data, &reconnectData); err != nil {
		c.sendError("Invalid reconnect data")
		return
	}

	// Find game
	var game *models.Game
	var playerID string

	// Search for game by ID or username
	if reconnectData.GameID != "" {
		game = c.service.GetGame(reconnectData.GameID)
		if game != nil {
			if game.Player1.Username == reconnectData.Username {
				playerID = game.Player1.ID
			} else if game.Player2 != nil && game.Player2.Username == reconnectData.Username {
				playerID = game.Player2.ID
			}
		}
	}

	if game == nil || playerID == "" {
		c.sendError("Game not found or already finished")
		return
	}

	// Reconnect player
	if game.Player1.ID == playerID {
		c.player = game.Player1
	} else {
		c.player = game.Player2
	}
	c.gameID = game.ID

	// Register client
	clientsMutex.Lock()
	clients[playerID] = c
	clientsMutex.Unlock()

	// Mark player as reconnected
	c.service.ReconnectPlayer(playerID)

	// Send current game state
	c.sendMessage(models.WSMessage{
		Type: models.MsgTypeGameUpdate,
		Payload: models.GameUpdatePayload{
			Game:    game,
			Message: "Reconnected successfully",
		},
	})
}

func (c *Client) handleDisconnect() {
	if c.player == nil {
		return
	}

	clientsMutex.Lock()
	delete(clients, c.player.ID)
	clientsMutex.Unlock()

	// Remove from matchmaking queue
	c.matchmaking.RemoveFromQueue(c.player.ID)

	// Mark as disconnected for reconnection window
	if c.gameID != "" {
		game := c.service.GetGame(c.gameID)
		if game != nil && game.State == models.GameStatePlaying {
			c.service.MarkPlayerDisconnected(c.player.ID)
			c.notifyOpponent(game, models.MsgTypeOpponentLeft)
		}
	}
}

func (c *Client) notifyOpponent(game *models.Game, msgType models.MessageType) {
	if game == nil {
		return
	}

	var opponentID string
	if c.player.ID == game.Player1.ID && game.Player2 != nil {
		if game.Player2.IsBot {
			return
		}
		opponentID = game.Player2.ID
	} else if game.Player1 != nil {
		if game.Player1.IsBot {
			return
		}
		opponentID = game.Player1.ID
	}

	if opponentID == "" {
		return
	}

	clientsMutex.RLock()
	opponent := clients[opponentID]
	clientsMutex.RUnlock()

	if opponent == nil {
		return
	}

	var payload interface{}
	switch msgType {
	case models.MsgTypeGameStart:
		payload = models.GameStartPayload{
			Game:         game,
			YourPlayerID: opponentID,
		}
	case models.MsgTypeGameUpdate:
		payload = models.GameUpdatePayload{
			Game: game,
		}
	case models.MsgTypeGameOver:
		winnerName := ""
		reason := "draw"
		message := "Game ended in a draw!"

		if game.Winner != nil {
			winnerName = game.Winner.Username
			reason = "win"
			message = winnerName + " wins!"
		}

		payload = models.GameOverPayload{
			Game:    game,
			Winner:  winnerName,
			Reason:  reason,
			Message: message,
		}
	case models.MsgTypeOpponentLeft:
		payload = models.ErrorPayload{
			Message: "Opponent disconnected. They have 30 seconds to reconnect.",
		}
	}

	opponent.sendMessage(models.WSMessage{
		Type:    msgType,
		Payload: payload,
	})
}

func (c *Client) sendMessage(msg models.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		log.Println("Client send buffer full")
	}
}

func (c *Client) sendError(message string) {
	c.sendMessage(models.WSMessage{
		Type: models.MsgTypeError,
		Payload: models.ErrorPayload{
			Message: message,
		},
	})
}
