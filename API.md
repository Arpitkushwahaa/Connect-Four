# API Documentation

## Base URLs

- **WebSocket**: `ws://localhost:8080/ws`
- **REST API**: `http://localhost:8080/api`

## WebSocket API

### Connection

Connect to the WebSocket endpoint:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('Connected');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  handleMessage(data);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected');
};
```

### Message Format

All messages follow this structure:

```json
{
  "type": "message_type",
  "payload": {
    // Type-specific data
  }
}
```

---

## Client → Server Messages

### 1. Join Queue

Join the matchmaking queue to find a game.

**Message Type:** `join_queue`

**Payload:**
```json
{
  "username": "string" // Required, 1-20 characters
}
```

**Example:**
```javascript
ws.send(JSON.stringify({
  type: 'join_queue',
  payload: {
    username: 'Alice'
  }
}));
```

**Response:** 
- After matchmaking: `game_start` message
- After 10 seconds with no opponent: `game_start` with bot

---

### 2. Make Move

Drop a disc into a column.

**Message Type:** `move`

**Payload:**
```json
{
  "column": number // 0-6, column index
}
```

**Example:**
```javascript
ws.send(JSON.stringify({
  type: 'move',
  payload: {
    column: 3
  }
}));
```

**Responses:**
- Success: `game_update` with new board state
- Invalid move: `invalid_move` error
- Game ends: `game_over` message

**Validation:**
- Column must be 0-6
- Column must not be full
- Must be your turn

---

### 3. Reconnect

Reconnect to an existing game after disconnection.

**Message Type:** `reconnect`

**Payload:**
```json
{
  "username": "string", // Your username
  "gameId": "string"    // Game ID (optional)
}
```

**Example:**
```javascript
ws.send(JSON.stringify({
  type: 'reconnect',
  payload: {
    username: 'Alice',
    gameId: 'game-12345'
  }
}));
```

**Responses:**
- Success: `game_update` with current state
- Failure: `error` message

**Notes:**
- Must reconnect within 30 seconds of disconnect
- Game ID is optional if username is unique

---

## Server → Client Messages

### 1. Game Start

Sent when a game begins.

**Message Type:** `game_start`

**Payload:**
```json
{
  "game": {
    "id": "string",
    "player1": {
      "id": "string",
      "username": "string",
      "isBot": false
    },
    "player2": {
      "id": "string",
      "username": "string",
      "isBot": true
    },
    "board": [[0,0,0,0,0,0,0], ...], // 6 rows x 7 columns
    "currentTurn": 1, // 1 or 2
    "state": "playing",
    "startTime": "2024-01-01T00:00:00Z"
  },
  "yourPlayerId": "string"
}
```

**Board Values:**
- `0` = Empty
- `1` = Player 1's disc (Red)
- `2` = Player 2's disc (Yellow)

---

### 2. Game Update

Sent after each move.

**Message Type:** `game_update`

**Payload:**
```json
{
  "game": {
    "id": "string",
    "player1": { ... },
    "player2": { ... },
    "board": [[...], ...],
    "currentTurn": 1,
    "state": "playing",
    "lastMoveCol": 3,
    "lastMoveRow": 5
  },
  "message": "Bot made a move" // Optional
}
```

**Last Move:**
- `lastMoveCol`: Column where disc was dropped (0-6)
- `lastMoveRow`: Row where disc landed (0-5)

---

### 3. Game Over

Sent when game ends.

**Message Type:** `game_over`

**Payload:**
```json
{
  "game": {
    "id": "string",
    "player1": { ... },
    "player2": { ... },
    "board": [[...], ...],
    "state": "finished",
    "winner": {
      "id": "string",
      "username": "Alice",
      "isBot": false
    },
    "winningLine": [[3,2], [3,3], [3,4], [3,5]], // Winning positions
    "endTime": "2024-01-01T00:05:30Z"
  },
  "winner": "Alice",
  "reason": "win", // "win" or "draw" or "forfeit"
  "message": "Alice wins!"
}
```

**Winning Line:**
- Array of [row, column] positions
- Shows the 4 connected discs

**Reasons:**
- `win` - Player connected 4
- `draw` - Board full, no winner
- `forfeit` - Player didn't reconnect in 30 seconds

---

### 4. Invalid Move

Sent when a move is rejected.

**Message Type:** `invalid_move`

**Payload:**
```json
{
  "message": "Column is full"
}
```

**Common Reasons:**
- "Column is full"
- "Not your turn"
- "Invalid column"
- "Game is not active"

---

### 5. Error

Sent for general errors.

**Message Type:** `error`

**Payload:**
```json
{
  "message": "Error description"
}
```

---

### 6. Opponent Left

Sent when opponent disconnects.

**Message Type:** `opponent_left`

**Payload:**
```json
{
  "message": "Opponent disconnected. They have 30 seconds to reconnect."
}
```

**Notes:**
- Opponent has 30 seconds to reconnect
- If they don't reconnect, you win by forfeit
- You'll receive `game_over` message if they don't return

---

## REST API Endpoints

### Get Leaderboard

Get top 10 players ranked by wins.

**Endpoint:** `GET /api/leaderboard`

**Response:**
```json
[
  {
    "username": "Alice",
    "wins": 15,
    "losses": 5,
    "draws": 2
  },
  {
    "username": "Bob",
    "wins": 12,
    "losses": 8,
    "draws": 1
  }
]
```

**Example:**
```javascript
fetch('http://localhost:8080/api/leaderboard')
  .then(res => res.json())
  .then(data => console.log(data));
```

**Sorting:**
- Primary: Wins (descending)
- Limit: Top 10 players

---

### Health Check

Check server status.

**Endpoint:** `GET /api/health`

**Response:**
```
OK
```

**Status Codes:**
- `200` - Server is healthy
- `500` - Server error

---

## Game Object Schema

### Complete Game Object

```typescript
interface Game {
  id: string;                    // Unique game ID
  player1: Player;               // First player (Red)
  player2: Player | null;        // Second player (Yellow)
  board: number[][];             // 6x7 grid (0=empty, 1=p1, 2=p2)
  currentTurn: number;           // 1 or 2
  state: GameState;              // "waiting" | "playing" | "finished"
  winner?: Player;               // Present if game finished with winner
  winningLine?: number[][];      // Winning disc positions
  startTime: string;             // ISO timestamp
  endTime?: string;              // ISO timestamp when finished
  lastMoveCol?: number;          // Last move column (0-6)
  lastMoveRow?: number;          // Last move row (0-5)
}

interface Player {
  id: string;                    // Unique player ID
  username: string;              // Display name
  isBot: boolean;                // True if AI opponent
}

type GameState = "waiting" | "playing" | "finished";
```

---

## Error Codes

| Code | Message | Description |
|------|---------|-------------|
| 1000 | Normal Closure | Connection closed normally |
| 1001 | Going Away | Browser navigating away |
| 1002 | Protocol Error | WebSocket protocol error |
| 1003 | Unsupported Data | Received invalid data type |
| 1011 | Internal Error | Server encountered an error |

---

## Usage Examples

### Complete Game Flow

```javascript
// 1. Connect
const ws = new WebSocket('ws://localhost:8080/ws');

// 2. Join queue when connected
ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'join_queue',
    payload: { username: 'Alice' }
  }));
};

// 3. Handle messages
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  switch (message.type) {
    case 'game_start':
      console.log('Game started!', message.payload.game);
      const myPlayerId = message.payload.yourPlayerId;
      break;
      
    case 'game_update':
      console.log('Board updated:', message.payload.game.board);
      // Update UI with new board state
      break;
      
    case 'game_over':
      console.log('Game ended:', message.payload.message);
      console.log('Winner:', message.payload.winner);
      break;
      
    case 'error':
      console.error('Error:', message.payload.message);
      break;
  }
};

// 4. Make a move
function makeMove(column) {
  ws.send(JSON.stringify({
    type: 'move',
    payload: { column }
  }));
}

// 5. Reconnect if disconnected
ws.onclose = () => {
  setTimeout(() => {
    const newWs = new WebSocket('ws://localhost:8080/ws');
    newWs.onopen = () => {
      newWs.send(JSON.stringify({
        type: 'reconnect',
        payload: {
          username: 'Alice',
          gameId: currentGameId
        }
      }));
    };
  }, 2000);
};
```

---

## Rate Limiting

**Current Implementation:** None

**Recommended for Production:**
- Max 10 moves per second per player
- Max 100 WebSocket connections per IP
- Max 1000 requests per hour to REST API

---

## CORS Configuration

**Current:** Allow all origins (`*`)

**Production Recommendation:**
```go
AllowedOrigins: []string{"https://yourdomain.com"}
```

---

## Testing with cURL

### Health Check
```bash
curl http://localhost:8080/api/health
```

### Get Leaderboard
```bash
curl http://localhost:8080/api/leaderboard
```

---

## Testing with wscat

Install wscat:
```bash
npm install -g wscat
```

Connect and test:
```bash
# Connect
wscat -c ws://localhost:8080/ws

# Send join queue message
{"type":"join_queue","payload":{"username":"TestUser"}}

# Make a move
{"type":"move","payload":{"column":3}}
```

---

## WebSocket State Machine

```
[Disconnected] 
      ↓ (connect)
[Connected] 
      ↓ (join_queue)
[In Queue]
      ↓ (matched)
[In Game - Waiting Turn]
      ↓ (your turn)
[In Game - Your Turn]
      ↓ (make move)
[In Game - Waiting Turn]
      ↓ (game ends)
[Game Finished]
      ↓ (close or play again)
[Disconnected]
```

---

## Best Practices

1. **Connection Management:**
   - Implement exponential backoff for reconnection
   - Handle connection timeouts
   - Clean up on component unmount

2. **Error Handling:**
   - Always handle `error` messages
   - Validate data before sending
   - Show user-friendly error messages

3. **State Synchronization:**
   - Trust server state as source of truth
   - Update local state on every `game_update`
   - Don't allow moves when it's not your turn

4. **Performance:**
   - Debounce user inputs
   - Use efficient rendering
   - Clean up event listeners

---

For more details, see the main [README.md](README.md) and [ARCHITECTURE.md](ARCHITECTURE.md).
