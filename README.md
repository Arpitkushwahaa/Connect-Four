# 🎮 4 in a Row - Real-Time Multiplayer Connect Four

<div align="center">

![Status](https://img.shields.io/badge/Status-Production%20Ready-success?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-1.21-00ADD8?style=for-the-badge&logo=go)
![React](https://img.shields.io/badge/React-18.2-61DAFB?style=for-the-badge&logo=react)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)
![Kafka](https://img.shields.io/badge/Kafka-Enabled-231F20?style=for-the-badge&logo=apache-kafka)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=for-the-badge&logo=postgresql)

**A production-ready real-time multiplayer Connect Four game with intelligent AI, live analytics, and beautiful UI**

[Quick Start](#-quick-start) • [Features](#-features) • [Demo](#-how-to-play) • [Docs](#-documentation) • [Deploy](#-deployment)

</div>

---

## 🌟 Features

### Core Gameplay
- ✅ **Real-time Multiplayer** - Play against other players in real-time using WebSockets
- ✅ **Smart Bot Opponent** - Competitive AI bot that plays strategically, not randomly
- ✅ **10-Second Matchmaking** - Automatic bot match if no player joins within 10 seconds
- ✅ **Player Reconnection** - 30-second window to reconnect if disconnected
- ✅ **Game State Persistence** - Active games stored in-memory, completed games in PostgreSQL
- ✅ **Leaderboard System** - Track wins, losses, and draws for all players

### Technical Features
- ✅ **WebSocket Communication** - Real-time bidirectional communication
- ✅ **Kafka Analytics** - Decoupled analytics service tracking game events
- ✅ **Docker Deployment** - Complete containerized setup
- ✅ **Microservices Architecture** - Separate backend, analytics, and frontend services
- ✅ **Strategic Bot AI** - Bot analyzes board to block wins and create opportunities

## 🏗️ Architecture

```
┌─────────────┐      WebSocket       ┌─────────────┐
│   Frontend  │ ◄──────────────────► │   Backend   │
│   (React)   │                       │   (GoLang)  │
└─────────────┘                       └──────┬──────┘
                                             │
                                      ┌──────┴──────┐
                                      │             │
                                 ┌────▼────┐   ┌───▼────┐
                                 │  Kafka  │   │Postgres│
                                 └────┬────┘   └────────┘
                                      │
                                 ┌────▼────────┐
                                 │  Analytics  │
                                 │  (Consumer) │
                                 └─────────────┘
```

## 📋 Requirements

- Docker & Docker Compose
- (Optional) Go 1.21+ for local development
- (Optional) Node.js 18+ for local frontend development

## 🚀 Quick Start

### Using Docker (Recommended)

1. **Clone the repository**
```bash
cd Connect-four
```

2. **Start all services**
```bash
docker-compose up -d
```

3. **Access the application**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- WebSocket: ws://localhost:8080/ws

4. **View logs**
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f analytics
```

5. **Stop services**
```bash
docker-compose down
```

## 💻 Local Development

### Backend

```bash
cd backend
go mod download
go run main.go
```

Environment variables:
- `DATABASE_URL`: PostgreSQL connection string (default: `postgres://postgres:postgres@localhost:5432/connectfour?sslmode=disable`)
- `KAFKA_BROKER`: Kafka broker address (default: `localhost:9092`)
- `PORT`: Server port (default: `8080`)

### Analytics Service

```bash
cd analytics
go mod download
go run main.go
```

### Frontend

```bash
cd frontend
npm install
npm start
```

Environment variables (in `.env` file):
- `REACT_APP_WS_URL`: WebSocket URL (default: `ws://localhost:8080/ws`)
- `REACT_APP_API_URL`: API URL (default: `http://localhost:8080`)

## 🎯 How to Play

1. **Enter Your Username** - Type your name and click "Join Game"
2. **Wait for Matchmaking** - You'll be matched with another player or a bot
3. **Play Your Turn** - Click on a column to drop your disc
4. **Win the Game** - Connect 4 discs horizontally, vertically, or diagonally
5. **View Leaderboard** - Check rankings and stats

### Game Rules

- **Board**: 7 columns × 6 rows
- **Players**: 2 players (Red vs Yellow)
- **Objective**: Connect 4 discs in a row (horizontal, vertical, or diagonal)
- **Turns**: Alternating turns between players
- **Draw**: Game ends in a draw if the board fills up with no winner

## 🤖 Bot Strategy

The bot uses strategic AI with the following priority system:

1. **Win if possible** - Takes winning move immediately
2. **Block opponent's win** - Prevents opponent from winning
3. **Create threats** - Sets up multiple winning opportunities
4. **Block opponent's threats** - Prevents opponent from creating multiple wins
5. **Strategic positioning** - Prefers center columns for better control
6. **Random valid move** - Fallback if no strategic move is available

The bot **does NOT** make random moves - it analyzes the board state and makes intelligent decisions.

## 📊 Analytics & Kafka

The analytics service consumes events from Kafka and tracks:

### Event Types
- `game_start` - Game initialization
- `game_move` - Player/bot moves
- `game_end` - Game completion

### Tracked Metrics
- Total games started
- Total games completed
- Total moves made
- Average game duration
- Games per hour
- Most frequent winners
- Player-specific statistics

### Viewing Analytics

Connect to the database:
```bash
docker exec -it connectfour-db psql -U postgres -d connectfour
```

Query analytics:
```sql
-- View all metrics
SELECT * FROM analytics_summary;

-- View recent game events
SELECT * FROM game_events ORDER BY timestamp DESC LIMIT 10;

-- View leaderboard
SELECT * FROM leaderboard ORDER BY wins DESC;
```

## 🗄️ Database Schema

### Tables

**games**
- Stores completed game records
- Fields: id, player1, player2, winner, duration, total_moves, completed_at, etc.

**leaderboard**
- Tracks player statistics
- Fields: username, wins, losses, draws

**game_events**
- Stores all game events from Kafka
- Fields: id, event_type, game_id, player, data, timestamp

**analytics_summary**
- Aggregated metrics
- Fields: id, metric_name, metric_value, updated_at

## 🔌 API Endpoints

### WebSocket (`/ws`)

**Join Queue**
```json
{
  "type": "join_queue",
  "payload": {
    "username": "PlayerName"
  }
}
```

**Make Move**
```json
{
  "type": "move",
  "payload": {
    "column": 3
  }
}
```

**Reconnect**
```json
{
  "type": "reconnect",
  "payload": {
    "username": "PlayerName",
    "gameId": "game-id"
  }
}
```

### REST API

**GET /api/leaderboard**
- Returns top 10 players by wins

**GET /api/health**
- Health check endpoint

## 🐳 Docker Services

| Service | Port | Description |
|---------|------|-------------|
| Frontend | 3000 | React application |
| Backend | 8080 | Go WebSocket server |
| PostgreSQL | 5432 | Database |
| Kafka | 9092 | Message broker |
| Zookeeper | 2181 | Kafka coordination |

## 🔧 Configuration

### Docker Compose Override

Create `docker-compose.override.yml` for custom settings:

```yaml
version: '3.8'
services:
  backend:
    environment:
      - DEBUG=true
  frontend:
    ports:
      - "3001:80"
```

### Environment Variables

**Backend:**
- `DATABASE_URL` - PostgreSQL connection
- `KAFKA_BROKER` - Kafka broker address
- `PORT` - Server port

**Frontend:**
- `REACT_APP_WS_URL` - WebSocket endpoint
- `REACT_APP_API_URL` - API endpoint

## 🧪 Testing

### Test the Game Flow

1. Open two browser windows at http://localhost:3000
2. Enter different usernames in each
3. Click "Join Game" in both
4. They should be matched together
5. Play a game to test real-time sync

### Test Bot Matchmaking

1. Open one browser window
2. Enter username and join
3. Wait 10 seconds
4. Bot should automatically join

### Test Reconnection

1. Start a game
2. Close the browser tab
3. Reopen within 30 seconds
4. Enter same username and game should reconnect

## 📈 Monitoring

### View Kafka Messages

```bash
# Enter Kafka container
docker exec -it connectfour-kafka bash

# Consume messages from game-events topic
kafka-console-consumer --bootstrap-server localhost:9092 --topic game-events --from-beginning
```

### Database Queries

```sql
-- Active games count
SELECT COUNT(*) FROM games WHERE completed_at > NOW() - INTERVAL '1 hour';

-- Top winners
SELECT username, wins FROM leaderboard ORDER BY wins DESC LIMIT 5;

-- Average game duration
SELECT AVG(duration) FROM games;
```

## 🛠️ Troubleshooting

### WebSocket Connection Issues

- Ensure backend is running on port 8080
- Check CORS settings in backend
- Verify WebSocket URL in frontend `.env`

### Kafka Not Working

- Wait for Kafka to fully start (can take 30-60 seconds)
- Check Kafka health: `docker-compose ps`
- View Kafka logs: `docker-compose logs kafka`

### Database Connection Errors

- Ensure PostgreSQL is healthy: `docker-compose ps postgres`
- Check connection string format
- Verify database exists: `docker exec -it connectfour-db psql -U postgres -l`

### Port Already in Use

```bash
# Find process using port
netstat -ano | findstr :8080

# Kill process (replace PID)
taskkill /PID <PID> /F
```

## 🚢 Deployment

### Production Build

```bash
# Build all services
docker-compose build

# Start in production mode
docker-compose up -d
```

### Environment-Specific Configs

For production deployment, update:
1. Database credentials
2. Kafka broker addresses
3. Frontend WebSocket URLs
4. Add SSL/TLS certificates
5. Configure proper networking

### Hosting Recommendations

- **Backend**: Any cloud platform supporting Docker (AWS ECS, GCP Cloud Run, Azure Container Instances)
- **Frontend**: Nginx, Vercel, Netlify, or S3 + CloudFront
- **Database**: AWS RDS, Google Cloud SQL, or managed PostgreSQL
- **Kafka**: AWS MSK, Confluent Cloud, or self-hosted

## 📝 Project Structure

```
Connect-four/
├── backend/
│   ├── handlers/          # WebSocket & HTTP handlers
│   ├── models/            # Data models
│   ├── services/          # Business logic
│   ├── main.go           # Entry point
│   ├── Dockerfile
│   └── go.mod
├── analytics/
│   ├── main.go           # Kafka consumer
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── public/
│   ├── src/
│   │   ├── components/   # React components
│   │   ├── App.js
│   │   └── App.css
│   ├── Dockerfile
│   ├── nginx.conf
│   └── package.json
├── docker-compose.yml
└── README.md
```

## 🎓 Technical Highlights

### Backend (GoLang)
- Gorilla WebSocket for real-time communication
- Concurrent game state management with mutexes
- Strategic bot AI with move analysis
- Kafka producer for event streaming
- PostgreSQL integration with connection pooling

### Frontend (React)
- Real-time UI updates via WebSocket
- Smooth animations for disc drops
- Responsive design for mobile/desktop
- Auto-reconnection logic
- Live leaderboard updates

### Analytics
- Kafka consumer group for scalability
- Real-time metric calculations
- Event sourcing pattern
- Time-series data aggregation

## 🤝 Contributing

This is a backend engineering assignment project. Feel free to:
- Report bugs
- Suggest improvements
- Fork and enhance

## 📄 License

This project is created as part of a backend engineering intern assignment.

## 👨‍💻 Author

Built with ❤️ using GoLang, React, Kafka, and PostgreSQL

---

**Enjoy playing 4 in a Row! 🎮**
