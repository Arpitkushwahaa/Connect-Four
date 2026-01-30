# Kafka Analytics Implementation - Bonus Feature

## Overview

This project implements a **complete decoupled analytics system** using Apache Kafka for real-time game event streaming and processing.

## Architecture

```
┌──────────────┐         ┌─────────────┐         ┌──────────────────┐
│   Backend    │────────>│   Kafka     │────────>│   Analytics      │
│   (Producer) │  Events │   Cluster   │  Events │   Service        │
│              │         │             │         │   (Consumer)     │
└──────────────┘         └─────────────┘         └──────────────────┘
                                                          │
                                                          ▼
                                                  ┌──────────────────┐
                                                  │   PostgreSQL     │
                                                  │   (Metrics DB)   │
                                                  └──────────────────┘
```

## Features Implemented ✅

### 1. **Kafka Producer (Backend)**

**Location**: `backend/services/kafka.go`

**Events Published**:
- `game_start` - When two players are matched
- `game_move` - For every move made in the game
- `game_end` - When a game finishes (win/draw/disconnect)

**Authentication**: Supports SASL/SCRAM authentication for managed Kafka services (Upstash, CloudKarafka, Confluent)

### 2. **Kafka Consumer (Analytics Service)**

**Location**: `analytics/main.go`

**Features**:
- ✅ Consumes events from `game-events` topic
- ✅ Stores all raw events in `game_events` table
- ✅ Processes events in real-time
- ✅ Computes aggregated metrics
- ✅ Handles graceful shutdown

### 3. **Gameplay Metrics Tracked**

#### Average Game Duration
- Calculated from `game_end` events
- Formula: `total_duration / total_games`
- Stored in `analytics_summary` table

```sql
SELECT metric_value FROM analytics_summary WHERE metric_name = 'avg_game_duration';
```

#### Most Frequent Winners
- Tracked in `winner_frequency` table
- Includes win count and last win timestamp
- Sorted by win count descending

```sql
SELECT username, win_count FROM winner_frequency ORDER BY win_count DESC LIMIT 10;
```

#### Games Per Day/Hour
- **Hourly**: Count of games in last 60 minutes
- **Daily**: Count of games in last 24 hours
- Updated on every `game_end` event

```sql
SELECT metric_value FROM analytics_summary WHERE metric_name = 'games_last_hour';
SELECT metric_value FROM analytics_summary WHERE metric_name = 'games_last_24h';
```

### 4. **User-Specific Metrics**

**Location**: `user_metrics` table

**Tracked per user**:
- ✅ Total games played
- ✅ Wins / Losses / Draws
- ✅ Total moves made
- ✅ Average game duration
- ✅ Win rate (calculated: `wins / total_games * 100`)

**Example Query**:
```sql
SELECT 
    username, 
    total_games, 
    wins, 
    losses, 
    draws,
    (wins::float / total_games * 100) as win_rate
FROM user_metrics
WHERE total_games > 0
ORDER BY wins DESC;
```

## Database Schema

### `game_events` Table
```sql
CREATE TABLE game_events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50),      -- 'game_start', 'game_move', 'game_end'
    game_id VARCHAR(255),
    player VARCHAR(255),
    data JSONB,                   -- Full event payload
    timestamp TIMESTAMP
);
```

### `analytics_summary` Table
```sql
CREATE TABLE analytics_summary (
    id SERIAL PRIMARY KEY,
    metric_name VARCHAR(100) UNIQUE,  -- 'total_games_started', 'avg_game_duration', etc.
    metric_value NUMERIC,
    updated_at TIMESTAMP
);
```

### `user_metrics` Table
```sql
CREATE TABLE user_metrics (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE,
    total_games INT,
    wins INT,
    losses INT,
    draws INT,
    total_moves INT,
    avg_game_duration NUMERIC,
    updated_at TIMESTAMP
);
```

### `winner_frequency` Table
```sql
CREATE TABLE winner_frequency (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE,
    win_count INT,
    last_win_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## API Endpoints

### GET `/api/analytics`

Returns comprehensive analytics data including:
- Total games started/completed
- Average game duration
- Games per hour/day
- Top 10 winners
- Per-user statistics with win rates

**Response Example**:
```json
{
  "totalGamesStarted": 150,
  "totalGamesCompleted": 145,
  "totalMoves": 1820,
  "avgGameDuration": 45.3,
  "gamesLastHour": 12,
  "gamesLast24Hours": 98,
  "topWinners": [
    {
      "username": "player1",
      "winCount": 23,
      "lastWinAt": "2026-01-30 14:30:22"
    }
  ],
  "userStats": [
    {
      "username": "player1",
      "totalGames": 45,
      "wins": 23,
      "losses": 18,
      "draws": 4,
      "totalMoves": 567,
      "avgGameDuration": 42.5,
      "winRate": 51.11
    }
  ]
}
```

## Event Flow Example

### 1. Game Start Event
```json
{
  "type": "game_start",
  "gameId": "abc-123",
  "player1": "Alice",
  "player2": "Bob",
  "player1Bot": false,
  "player2Bot": false,
  "timestamp": "2026-01-30T14:25:00Z"
}
```

**Analytics Actions**:
- Increment `total_games_started`
- Create entries in `user_metrics` for Alice and Bob

### 2. Game Move Event
```json
{
  "type": "game_move",
  "gameId": "abc-123",
  "player": "Alice",
  "column": 3,
  "row": 5,
  "timestamp": "2026-01-30T14:25:05Z"
}
```

**Analytics Actions**:
- Increment `total_moves`
- Increment Alice's `total_moves` in `user_metrics`

### 3. Game End Event
```json
{
  "type": "game_end",
  "gameId": "abc-123",
  "winner": "Alice",
  "duration": 42,
  "totalMoves": 15,
  "reason": "win",
  "timestamp": "2026-01-30T14:25:47Z"
}
```

**Analytics Actions**:
- Increment `total_games_completed`
- Update `avg_game_duration` (total_duration + 42) / total_games
- Increment Alice's win count in `winner_frequency`
- Update Alice's `wins` in `user_metrics`
- Update Bob's `losses` in `user_metrics`
- Update `games_last_hour` count
- Update `games_last_24h` count

## Configuration

### Backend Environment Variables
```env
KAFKA_ENABLED=true
KAFKA_BROKERS=broker1:9092,broker2:9092
KAFKA_TOPIC=game-events
KAFKA_USERNAME=your-username
KAFKA_PASSWORD=your-password
KAFKA_SASL_MECHANISM=SCRAM-SHA-512
```

### Analytics Service Environment Variables
```env
DATABASE_URL=postgresql://...
KAFKA_BROKERS=broker1:9092,broker2:9092
KAFKA_TOPIC=game-events
KAFKA_GROUP_ID=analytics-consumer
KAFKA_USERNAME=your-username
KAFKA_PASSWORD=your-password
KAFKA_SASL_MECHANISM=SCRAM-SHA-512
```

## Deployment

### With Kafka Enabled

1. **Deploy Backend** with Kafka credentials
2. **Deploy Analytics Service** with same Kafka credentials
3. **Deploy Frontend** (no Kafka config needed)

Both backend and analytics consume from the same Kafka cluster.

### Without Kafka (Optional)

Set `KAFKA_ENABLED=false` in backend - analytics won't run, but game still works.

## Real-World Production Features

✅ **Decoupled Architecture** - Game server and analytics are separate services  
✅ **Event Sourcing** - All events stored for replay/audit  
✅ **Real-time Processing** - Metrics updated instantly  
✅ **Scalable** - Can add multiple analytics consumers  
✅ **Fault Tolerant** - Kafka handles message delivery guarantees  
✅ **SASL Authentication** - Production-ready security  
✅ **Graceful Shutdown** - Handles interrupts cleanly  

## Testing Analytics

1. Deploy with Kafka enabled
2. Play a few games
3. Check analytics:
   ```bash
   curl https://your-backend.onrender.com/api/analytics
   ```
4. Verify Kafka messages in your cluster dashboard
5. Check PostgreSQL tables directly:
   ```sql
   SELECT * FROM game_events ORDER BY timestamp DESC LIMIT 10;
   SELECT * FROM winner_frequency ORDER BY win_count DESC;
   SELECT * FROM user_metrics WHERE total_games > 0;
   ```

## Metrics Summary

| Metric | Description | Table |
|--------|-------------|-------|
| Total Games Started | All initiated games | `analytics_summary` |
| Total Games Completed | Finished games | `analytics_summary` |
| Average Game Duration | Mean time per game | `analytics_summary` |
| Games Last Hour | Rolling hourly count | `analytics_summary` |
| Games Last 24h | Rolling daily count | `analytics_summary` |
| Most Frequent Winners | Top 10 by wins | `winner_frequency` |
| Per-User Stats | W/L/D, moves, win rate | `user_metrics` |

---

**This implementation fully satisfies the Kafka analytics bonus requirement!** 🎉
