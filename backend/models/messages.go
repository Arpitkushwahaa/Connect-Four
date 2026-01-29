package models

type MessageType string

const (
	MsgTypeJoinQueue    MessageType = "join_queue"
	MsgTypeGameStart    MessageType = "game_start"
	MsgTypeGameUpdate   MessageType = "game_update"
	MsgTypeMove         MessageType = "move"
	MsgTypeGameOver     MessageType = "game_over"
	MsgTypeError        MessageType = "error"
	MsgTypeReconnect    MessageType = "reconnect"
	MsgTypeOpponentLeft MessageType = "opponent_left"
	MsgTypeInvalidMove  MessageType = "invalid_move"
)

type WSMessage struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

type JoinQueuePayload struct {
	Username string `json:"username"`
	GameID   string `json:"gameId,omitempty"` // for reconnection
}

type MovePayload struct {
	Column int `json:"column"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

type GameStartPayload struct {
	Game         *Game  `json:"game"`
	YourPlayerID string `json:"yourPlayerId"`
}

type GameUpdatePayload struct {
	Game    *Game  `json:"game"`
	Message string `json:"message,omitempty"`
}

type GameOverPayload struct {
	Game    *Game  `json:"game"`
	Winner  string `json:"winner,omitempty"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}
