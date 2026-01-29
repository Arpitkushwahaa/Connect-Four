# Architecture Documentation

## System Overview

The 4 in a Row application is built using a microservices architecture with the following components:

```
┌──────────────────────────────────────────────────────────────┐
│                          Client Layer                         │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │         React Frontend (Port 3000)                   │   │
│  │  - WebSocket Client                                  │   │
│  │  - Game UI Components                                │   │
│  │  - State Management                                  │   │
│  └─────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────┘
                           ↓ (WebSocket)
┌──────────────────────────────────────────────────────────────┐
│                      Application Layer                        │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │       GoLang Backend Server (Port 8080)             │   │
│  │                                                       │   │
│  │  ├─── WebSocket Handler                             │   │
│  │  ├─── Game Service                                  │   │
│  │  ├─── Matchmaking Service                           │   │
│  │  ├─── Bot AI Service                                │   │
│  │  └─── Kafka Producer                                │   │
│  └─────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────┘
           ↓                              ↓
┌──────────────────────┐     ┌────────────────────────────────┐
│   Data Layer         │     │    Message Queue               │
│                      │     │                                │
│  PostgreSQL          │     │    Kafka (Port 9092)           │
│  (Port 5432)         │     │    + Zookeeper (Port 2181)     │
│                      │     │                                │
│  - games             │     │    Topic: game-events          │
│  - leaderboard       │     │                                │
│  - analytics_summary │     │                                │
│  - game_events       │     │                                │
└──────────────────────┘     └────────────────────────────────┘
                                        ↓
                           ┌────────────────────────────────┐
                           │    Analytics Service           │
                           │    (Kafka Consumer)            │
                           │                                │
                           │    - Event Processing          │
                           │    - Metrics Calculation       │
                           │    - Data Aggregation          │
                           └────────────────────────────────┘
```

## Component Details

### 1. Frontend (React)

**Technology Stack:**
- React 18.2
- WebSocket API
- CSS3 with animations
- Responsive design

**Key Components:**
- `App.js` - Main application container
- `GameBoard.js` - 7x6 game grid with animations
- `Leaderboard.js` - Live leaderboard display

**State Management:**
- Local state using React hooks
- WebSocket connection management
- Real-time game state synchronization

**Features:**
- Real-time disc drop animations
- Turn indicators
- Winning line highlighting
- Auto-reconnection logic
- Responsive layout

### 2. Backend (GoLang)

**Technology Stack:**
- Go 1.21
- Gorilla WebSocket
- PostgreSQL driver (lib/pq)
- Kafka client (segmentio/kafka-go)

**Architecture Layers:**

#### Handlers Layer
- `websocket.go` - WebSocket connection management
- `leaderboard.go` - REST API for leaderboard

#### Services Layer
- `game_service.go` - Core game state management
- `matchmaking.go` - Player matching logic
- `bot.go` - AI opponent implementation
- `game_logic.go` - Game rules and win detection
- `kafka.go` - Event publishing
- `database.go` - Database operations

#### Models Layer
- `game.go` - Game data structures
- `messages.go` - WebSocket message types

**Concurrency:**
- Goroutines for matchmaking loop
- Mutex-protected shared state
- Channel-based client communication

### 3. Database (PostgreSQL)

**Schema Design:**

```sql
-- Completed games
CREATE TABLE games (
    id VARCHAR(255) PRIMARY KEY,
    player1 VARCHAR(255) NOT NULL,
    player2 VARCHAR(255) NOT NULL,
    winner VARCHAR(255),
    duration INTEGER NOT NULL,
    total_moves INTEGER NOT NULL,
    completed_at TIMESTAMP NOT NULL,
    player1_is_bot BOOLEAN DEFAULT FALSE,
    player2_is_bot BOOLEAN DEFAULT FALSE
);

-- Player statistics
CREATE TABLE leaderboard (
    username VARCHAR(255) PRIMARY KEY,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    draws INTEGER DEFAULT 0
);

-- Raw event log
CREATE TABLE game_events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    game_id VARCHAR(255),
    player VARCHAR(255),
    data JSONB,
    timestamp TIMESTAMP NOT NULL
);

-- Aggregated metrics
CREATE TABLE analytics_summary (
    id SERIAL PRIMARY KEY,
    metric_name VARCHAR(100) UNIQUE NOT NULL,
    metric_value NUMERIC NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

**Indexes:**
- `idx_games_completed_at` - Fast date queries
- `idx_leaderboard_wins` - Leaderboard sorting
- `idx_game_events_type` - Event filtering
- `idx_game_events_timestamp` - Time-series queries

### 4. Message Queue (Kafka)

**Topics:**
- `game-events` - All game-related events

**Event Types:**
```json
// Game Start
{
  "type": "game_start",
  "gameId": "...",
  "player1": "...",
  "player2": "...",
  "player1Bot": false,
  "player2Bot": true,
  "timestamp": "..."
}

// Game Move
{
  "type": "game_move",
  "gameId": "...",
  "player": "...",
  "column": 3,
  "row": 5,
  "timestamp": "..."
}

// Game End
{
  "type": "game_end",
  "gameId": "...",
  "winner": "...",
  "duration": 120,
  "totalMoves": 25,
  "reason": "win",
  "timestamp": "..."
}
```

**Consumer Groups:**
- `analytics-service` - Processes events for analytics

### 5. Analytics Service

**Responsibilities:**
- Consume events from Kafka
- Store raw events in database
- Calculate real-time metrics
- Aggregate statistics

**Tracked Metrics:**
- Total games started
- Total games completed
- Total moves made
- Average game duration
- Games per hour
- Win/loss ratios

## Data Flow

### Game Start Flow

```
1. User enters username → Frontend
2. WebSocket connect → Backend
3. Send "join_queue" message → Backend
4. Add to matchmaking queue → Matchmaking Service
5. Wait 10 seconds or match with player → Matchmaking Service
6. Create game → Game Service
7. Publish "game_start" event → Kafka
8. Send game state to both clients → WebSocket
9. Update UI → Frontend
```

### Move Flow

```
1. User clicks column → Frontend
2. Send "move" message → Backend via WebSocket
3. Validate move → Game Logic
4. Update board state → Game Service
5. Check win condition → Game Logic
6. Publish "game_move" event → Kafka
7. Send updated state to both clients → WebSocket
8. Animate disc drop → Frontend
9. If game over:
   - Save to database → Database
   - Update leaderboard → Database
   - Publish "game_end" event → Kafka
```

### Analytics Flow

```
1. Game event published → Kafka
2. Consumer receives event → Analytics Service
3. Store raw event → Database
4. Update metrics → Analytics Service
5. Calculate aggregations → Analytics Service
6. Store summary → Database
```

## Bot AI Algorithm

### Strategy Priority

1. **Immediate Win** - Check all columns for winning move
2. **Block Opponent** - Check if opponent can win next turn
3. **Create Double Threat** - Setup position with multiple winning paths
4. **Block Opponent Threat** - Prevent opponent's double threats
5. **Strategic Position** - Prefer center columns
6. **Random Valid Move** - Fallback

### Implementation

```go
// Pseudo-code
func GetMove(game) {
    // Check for immediate win
    if col := findWinningMove(game, botPlayer); col != -1 {
        return col
    }
    
    // Block opponent win
    if col := findWinningMove(game, opponent); col != -1 {
        return col
    }
    
    // Create threat (2+ ways to win next turn)
    if col := findThreatMove(game, botPlayer); col != -1 {
        return col
    }
    
    // Block opponent threat
    if col := findThreatMove(game, opponent); col != -1 {
        return col
    }
    
    // Center columns preferred
    for col in [3, 2, 4, 1, 5, 0, 6] {
        if isValid(col) {
            return col
        }
    }
}
```

## WebSocket Protocol

### Message Format

```typescript
{
  type: "message_type",
  payload: { /* type-specific data */ }
}
```

### Client → Server Messages

```typescript
// Join matchmaking
{
  type: "join_queue",
  payload: {
    username: string
  }
}

// Make move
{
  type: "move",
  payload: {
    column: number
  }
}

// Reconnect
{
  type: "reconnect",
  payload: {
    username: string,
    gameId: string
  }
}
```

### Server → Client Messages

```typescript
// Game start
{
  type: "game_start",
  payload: {
    game: Game,
    yourPlayerId: string
  }
}

// Game update
{
  type: "game_update",
  payload: {
    game: Game,
    message?: string
  }
}

// Game over
{
  type: "game_over",
  payload: {
    game: Game,
    winner: string,
    reason: string,
    message: string
  }
}

// Error
{
  type: "error",
  payload: {
    message: string
  }
}
```

## Reconnection Logic

### Server Side

1. **Disconnect Detection:**
   - WebSocket connection closed
   - Mark player as disconnected with timestamp
   - Keep game state in memory

2. **30-Second Window:**
   - Background goroutine checks every 5 seconds
   - If > 30 seconds, forfeit game
   - Opponent wins automatically

3. **Successful Reconnect:**
   - Client sends reconnect message with username/gameId
   - Verify game exists and player identity
   - Restore WebSocket connection
   - Send current game state

### Client Side

1. **Connection Lost:**
   - Detect WebSocket close
   - Show reconnection message
   - Attempt reconnect after 2 seconds

2. **Reconnect Attempt:**
   - Open new WebSocket
   - Send reconnect message
   - Restore game state on success

## Performance Considerations

### Scalability

**Current Limits:**
- Single server: ~1000 concurrent games
- Database: Optimized indexes for fast queries
- Kafka: Single partition (can scale to multiple)

**Scaling Strategy:**
- Horizontal scaling with load balancer
- Redis for distributed game state
- Kafka partitioning for throughput
- Database read replicas

### Optimization

1. **Frontend:**
   - Lazy loading components
   - Memoized calculations
   - Debounced updates

2. **Backend:**
   - Connection pooling
   - Mutex granularity
   - Efficient board checking

3. **Database:**
   - Indexed queries
   - Prepared statements
   - Connection pooling

## Security Considerations

### Current Implementation

- CORS enabled (configurable)
- Input validation on moves
- SQL injection prevention (parameterized queries)
- WebSocket origin checking

### Production Recommendations

1. Rate limiting on endpoints
2. Authentication/authorization
3. SSL/TLS for WebSocket (WSS)
4. Database encryption at rest
5. Secrets management
6. DDoS protection
7. Input sanitization

## Monitoring & Observability

### Metrics to Track

1. **Application:**
   - Active WebSocket connections
   - Games in progress
   - Matchmaking queue length
   - Bot vs human game ratio

2. **Performance:**
   - Request latency
   - Database query time
   - Kafka publish time
   - Memory usage

3. **Business:**
   - Daily active users
   - Game completion rate
   - Average game duration
   - Player retention

### Logging

**Levels:**
- ERROR - System failures
- WARN - Degraded performance
- INFO - Important events
- DEBUG - Detailed flow

**Key Events to Log:**
- Game creation
- Player connections/disconnections
- Move validation failures
- Database errors
- Kafka publish failures

## Future Enhancements

1. **Features:**
   - User authentication
   - Game history
   - Replay functionality
   - Custom room creation
   - Tournament mode

2. **Technical:**
   - Redis for distributed state
   - Horizontal scaling
   - Advanced bot difficulty levels
   - Move time limits
   - Spectator mode

3. **Analytics:**
   - Advanced player statistics
   - Heatmap of most used columns
   - Win pattern analysis
   - Bot performance metrics

---

This architecture is designed to be scalable, maintainable, and production-ready.
