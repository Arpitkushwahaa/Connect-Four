# Quick Start Guide

## Prerequisites

- Install Docker Desktop: https://www.docker.com/products/docker-desktop/
- Make sure Docker is running

## Start the Application

### Windows:
```bash
start.bat
```

### Mac/Linux:
```bash
chmod +x start.sh
./start.sh
```

### Manual Start:
```bash
docker-compose up -d
```

## Access the Application

Wait 30-60 seconds for all services to start, then open:

- **Game**: http://localhost:3000
- **Backend API**: http://localhost:8080/api/health

## How to Play

1. Open http://localhost:3000
2. Enter your username
3. Click "Join Game"
4. Wait for matchmaking (or bot will join in 10 seconds)
5. Click columns to drop your disc
6. Connect 4 in a row to win!

## View Leaderboard

Click "Show Leaderboard" button in the top-right corner

## Testing Multiplayer

1. Open http://localhost:3000 in two different browser windows
2. Enter different usernames
3. Join game in both windows
4. They will be matched together!

## Stop the Application

```bash
docker-compose down
```

## View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f analytics
```

## Troubleshooting

**Can't connect to the game?**
- Wait 60 seconds for all services to start
- Check Docker Desktop is running
- Run: `docker-compose ps` to see if all services are up

**Port already in use?**
- Stop other applications using ports 3000, 8080, 5432, 9092
- Or change ports in docker-compose.yml

**Game not loading?**
- Clear browser cache
- Open browser console (F12) to see errors
- Check backend logs: `docker-compose logs backend`

## Advanced

**Access Database:**
```bash
docker exec -it connectfour-db psql -U postgres -d connectfour
```

**View Kafka Messages:**
```bash
docker exec -it connectfour-kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic game-events --from-beginning
```

**Rebuild Services:**
```bash
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## Development Mode

### Backend:
```bash
cd backend
go run main.go
```

### Frontend:
```bash
cd frontend
npm install
npm start
```

### Analytics:
```bash
cd analytics
go run main.go
```

---

For full documentation, see README.md
