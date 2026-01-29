# Getting Started with 4 in a Row

Welcome! This guide will get you up and running in **less than 5 minutes**.

## Prerequisites

**Required:**
- Docker Desktop installed and running
- 2GB free RAM
- 10GB free disk space

**Optional (for development):**
- Go 1.21+ 
- Node.js 18+

## Step 1: Verify Docker

Open a terminal and run:

```bash
docker --version
docker-compose --version
```

If you see version numbers, you're good! If not, install Docker Desktop:
- Windows/Mac: https://www.docker.com/products/docker-desktop/
- Linux: https://docs.docker.com/engine/install/

## Step 2: Start the Application

### Windows

Double-click `start.bat` or run in terminal:
```bash
start.bat
```

### Mac/Linux

In terminal:
```bash
chmod +x start.sh
./start.sh
```

### Or use Docker Compose directly:
```bash
docker-compose up -d
```

## Step 3: Wait for Services

**Important:** Wait 30-60 seconds for all services to start!

You can check status:
```bash
docker-compose ps
```

All services should show "Up" or "healthy".

## Step 4: Open the Game

Open your browser and go to:
```
http://localhost:3000
```

You should see the 4 in a Row game!

## Step 5: Play!

1. **Enter your username** (any name you like)
2. **Click "Join Game"**
3. **Wait for matchmaking:**
   - If another player joins: Play against them!
   - If no one joins in 10 seconds: Bot will join!
4. **Click columns to drop your disc**
5. **Connect 4 in a row to win!**

## Testing Multiplayer

Want to test multiplayer?

1. Open http://localhost:3000 in **Chrome**
2. Open http://localhost:3000 in **Firefox** (or another Chrome window)
3. Enter different usernames in each
4. Click "Join Game" in both
5. They will be matched together!

## Testing the Bot

1. Open http://localhost:3000
2. Enter username
3. Join game
4. Wait 10 seconds
5. Bot joins automatically
6. Watch the bot play strategically!

## View Leaderboard

Click the **"Show Leaderboard"** button in the top-right corner.

## Troubleshooting

### "Can't connect" or "Connection error"

**Solution:**
1. Wait 60 seconds (services take time to start)
2. Refresh the page
3. Check Docker is running: `docker-compose ps`

### Ports already in use

**Solution:**
1. Stop other applications using ports 3000, 8080, 5432, 9092
2. Or edit `docker-compose.yml` to use different ports

### Docker not found

**Solution:**
Install Docker Desktop from https://www.docker.com/products/docker-desktop/

### Services won't start

**Solution:**
```bash
# Stop everything
docker-compose down

# Remove old volumes
docker-compose down -v

# Rebuild and start
docker-compose up -d --build
```

## View Logs

If something's not working, check the logs:

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f analytics
```

## Stop the Application

When you're done:

```bash
docker-compose down
```

This stops all services but keeps your data.

To remove everything including data:

```bash
docker-compose down -v
```

## Quick Commands Reference

| Action | Command |
|--------|---------|
| Start | `docker-compose up -d` |
| Stop | `docker-compose down` |
| Restart | `docker-compose restart` |
| Logs | `docker-compose logs -f` |
| Status | `docker-compose ps` |
| Rebuild | `docker-compose up -d --build` |

## What's Running?

When you start the application, these services run:

| Service | Port | What it does |
|---------|------|--------------|
| Frontend | 3000 | React game UI |
| Backend | 8080 | GoLang server |
| PostgreSQL | 5432 | Database |
| Kafka | 9092 | Event streaming |
| Analytics | - | Processes game events |
| Zookeeper | 2181 | Kafka coordinator |

## Accessing Services

**Game UI:**
```
http://localhost:3000
```

**Leaderboard API:**
```
http://localhost:8080/api/leaderboard
```

**Health Check:**
```
http://localhost:8080/api/health
```

**Database:**
```bash
docker exec -it connectfour-db psql -U postgres -d connectfour
```

**Kafka Messages:**
```bash
docker exec -it connectfour-kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic game-events --from-beginning
```

## Development Mode

Want to modify the code?

### Backend
```bash
cd backend
go run main.go
```

### Frontend
```bash
cd frontend
npm install
npm start
```

### Analytics
```bash
cd analytics
go run main.go
```

Make sure PostgreSQL and Kafka are running (via Docker Compose).

## Next Steps

**Learn More:**
- Read [README.md](README.md) for complete documentation
- Check [API.md](API.md) for API details
- See [ARCHITECTURE.md](ARCHITECTURE.md) for system design
- Review [DEPLOYMENT.md](DEPLOYMENT.md) for production deployment

**Customize:**
- Modify frontend in `frontend/src/`
- Change backend logic in `backend/services/`
- Add features and have fun!

## Support

If you encounter issues:

1. Check logs: `docker-compose logs -f`
2. Verify Docker is running
3. Make sure no other apps use the same ports
4. Try rebuilding: `docker-compose up -d --build`
5. Check the troubleshooting sections in documentation

## That's It!

You're all set! Enjoy playing 4 in a Row! 🎮

---

**Need help?** Check the documentation files or review the code!

**Want to contribute?** Feel free to enhance and customize!
