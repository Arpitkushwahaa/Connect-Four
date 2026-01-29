# Complete Deployment Guide with Kafka Integration

## Prerequisites

1. GitHub account with repository: `https://github.com/Arpitkushwahaa/Connect-Four`
2. Render.com account (free tier available)
3. Vercel account (free tier available)
4. Upstash account for Kafka (free tier available)

---

## Step 1: Set Up Upstash Kafka (5 minutes)

### 1.1 Create Kafka Cluster

1. Go to [console.upstash.com](https://console.upstash.com)
2. Sign up or log in with GitHub
3. Click **"Kafka"** in the sidebar
4. Click **"Create Cluster"**
5. Configure:
   - **Name**: `connect-four-kafka`
   - **Region**: `us-east-1` (or closest to you)
   - **Type**: **Single Replica** (free tier)
6. Click **"Create Cluster"**

### 1.2 Create Kafka Topic

1. Once cluster is created, click on it
2. Go to **"Topics"** tab
3. Click **"Create Topic"**
4. Configure:
   - **Topic Name**: `game-events`
   - **Partitions**: `1`
   - **Retention Time**: `604800000` (7 days in milliseconds)
   - **Retention Size**: `1GB`
5. Click **"Create"**

### 1.3 Copy Connection Details

From your cluster **Overview** page, copy these values (you'll need them later):

```
Bootstrap Endpoint: [YOUR_CLUSTER].upstash.io:9092
Username: [LONG_STRING]
Password: [LONG_STRING]
SASL Mechanism: SCRAM-SHA-512
```

**Keep these credentials safe!** You'll add them to Render in Step 3.

---

## Step 2: Deploy PostgreSQL Database on Render

1. Go to [dashboard.render.com](https://dashboard.render.com)
2. Click **"New +"** → **"PostgreSQL"**
3. Configure:
   - **Name**: `connect-four-db`
   - **Database**: `connectfour`
   - **User**: `postgres`
   - **Region**: Same as your backend (e.g., `Oregon (US West)`)
   - **Plan**: **Free**
4. Click **"Create Database"**
5. Wait for database to provision (~2 minutes)
6. **Copy the Internal Database URL** (starts with `postgresql://`)

---

## Step 3: Deploy Backend to Render

### 3.1 Create Backend Web Service

1. Click **"New +"** → **"Web Service"**
2. Connect your GitHub repository: `Arpitkushwahaa/Connect-Four`
3. Configure:
   - **Name**: `connect-four-backend`
   - **Root Directory**: `backend`
   - **Runtime**: `Go`
   - **Build Command**: `go build -o main .`
   - **Start Command**: `./main`
   - **Plan**: **Free**

### 3.2 Add Environment Variables

Click **"Advanced"** and add these environment variables:

| Key | Value |
|-----|-------|
| `PORT` | `8080` |
| `DATABASE_URL` | [Paste your database Internal URL from Step 2] |
| `KAFKA_ENABLED` | `true` |
| `KAFKA_BROKERS` | [Your Upstash endpoint, e.g., `smart-mantis-12345.upstash.io:9092`] |
| `KAFKA_TOPIC` | `game-events` |
| `KAFKA_USERNAME` | [Your Upstash username] |
| `KAFKA_PASSWORD` | [Your Upstash password] |
| `KAFKA_SASL_MECHANISM` | `SCRAM-SHA-512` |

### 3.3 Deploy

1. Click **"Create Web Service"**
2. Wait for deployment (~3-5 minutes)
3. **Copy your backend URL**: `https://connect-four-backend.onrender.com`

---

## Step 4: Deploy Analytics Service to Render

### 4.1 Create Analytics Web Service

1. Click **"New +"** → **"Web Service"**
2. Connect your GitHub repository: `Arpitkushwahaa/Connect-Four`
3. Configure:
   - **Name**: `connect-four-analytics`
   - **Root Directory**: `analytics`
   - **Runtime**: `Go`
   - **Build Command**: `go build -o analytics .`
   - **Start Command**: `./analytics`
   - **Plan**: **Free**

### 4.2 Add Environment Variables

| Key | Value |
|-----|-------|
| `DATABASE_URL` | [Same database URL from Step 2] |
| `KAFKA_BROKERS` | [Same Upstash endpoint] |
| `KAFKA_TOPIC` | `game-events` |
| `KAFKA_GROUP_ID` | `analytics-consumer` |
| `KAFKA_USERNAME` | [Same Upstash username] |
| `KAFKA_PASSWORD` | [Same Upstash password] |
| `KAFKA_SASL_MECHANISM` | `SCRAM-SHA-512` |

### 4.3 Deploy

1. Click **"Create Web Service"**
2. Wait for deployment (~3-5 minutes)

---

## Step 5: Deploy Frontend to Vercel

### 5.1 Import Project

1. Go to [vercel.com/new](https://vercel.com/new)
2. Click **"Import Git Repository"**
3. Select `Arpitkushwahaa/Connect-Four`
4. Configure:
   - **Project Name**: `connect-four`
   - **Framework Preset**: `Create React App`
   - **Root Directory**: `frontend`

### 5.2 Add Environment Variables

Click **"Environment Variables"** and add:

| Name | Value |
|------|-------|
| `REACT_APP_WS_URL` | `wss://connect-four-backend.onrender.com/ws` |
| `REACT_APP_API_URL` | `https://connect-four-backend.onrender.com` |

*(Replace with your actual backend URL from Step 3)*

### 5.3 Deploy

1. Click **"Deploy"**
2. Wait for deployment (~2-3 minutes)
3. **Your live URL**: `https://connect-four-[random].vercel.app`

---

## Step 6: Update Backend CORS

After frontend deploys, update CORS to allow your Vercel domain:

1. Go to your GitHub repository
2. Edit `backend/main.go` line 70
3. Change:
   ```go
   AllowedOrigins: []string{"*"},
   ```
   to:
   ```go
   AllowedOrigins: []string{"https://connect-four-[your-id].vercel.app", "http://localhost:3000"},
   ```
4. Commit and push - Render will auto-redeploy

---

## Step 7: Verify Everything Works

### 7.1 Check Backend Logs (Render)

Go to your backend service → **Logs** tab. You should see:
```
Server started on :8080
Kafka producer initialized with SASL authentication
Game service initialized with Kafka enabled
```

### 7.2 Check Analytics Logs (Render)

Go to your analytics service → **Logs** tab. You should see:
```
Analytics service started, listening for events...
Kafka consumer configured with SASL authentication
```

### 7.3 Check Upstash Kafka

1. Go to your Upstash cluster dashboard
2. Click on `game-events` topic
3. Go to **"Messages"** tab
4. You should see events appear when games are played

### 7.4 Test the Game

1. Open your Vercel URL: `https://connect-four-[your-id].vercel.app`
2. Enter a username and click **"Find Game"**
3. After 10 seconds, bot should join
4. Play a few moves
5. Check Upstash - you should see `game_start`, `game_move`, `game_end` events

### 7.5 Test Analytics

Query the leaderboard API:
```bash
curl https://connect-four-backend.onrender.com/api/leaderboard
```

You should see player stats being tracked!

---

## Architecture Summary

```
┌─────────────────┐
│   Frontend      │  (Vercel)
│   React App     │  https://connect-four.vercel.app
└────────┬────────┘
         │ WebSocket (WSS)
         ▼
┌─────────────────┐
│   Backend       │  (Render)
│   Go Server     │  https://connect-four-backend.onrender.com
└────┬────────┬───┘
     │        │
     │        └──────────────┐
     ▼                       ▼
┌─────────────┐      ┌──────────────┐
│ PostgreSQL  │      │ Upstash      │
│ Database    │      │ Kafka        │
│ (Render)    │      │ game-events  │
└──────┬──────┘      └───────┬──────┘
       │                     │
       │                     ▼
       │             ┌───────────────┐
       └────────────→│  Analytics    │
                     │  Service      │
                     │  (Render)     │
                     └───────────────┘
```

---

## What Gets Tracked by Kafka?

### Event Types

1. **game_start** - Emitted when two players are matched
   ```json
   {
     "type": "game_start",
     "gameId": "uuid",
     "player1": "username1",
     "player2": "Bot",
     "player1Bot": false,
     "player2Bot": true,
     "timestamp": "2026-01-29T..."
   }
   ```

2. **game_move** - Emitted for each move
   ```json
   {
     "type": "game_move",
     "gameId": "uuid",
     "player": "username1",
     "column": 3,
     "row": 5,
     "timestamp": "2026-01-29T..."
   }
   ```

3. **game_end** - Emitted when game finishes
   ```json
   {
     "type": "game_end",
     "gameId": "uuid",
     "winner": "username1",
     "duration": 45,
     "totalMoves": 12,
     "reason": "win",
     "timestamp": "2026-01-29T..."
   }
   ```

### Analytics Computed

- Total games played
- Average game duration
- Games per hour
- Win rates per player
- Most active players
- Bot vs Human stats

---

## Costs (All Free!)

- **Render Free Tier**:
  - 3 web services (Backend + Analytics + Database)
  - 750 hours/month (enough for 24/7 operation)
  - Database: 1GB storage, 97 connection limit

- **Vercel Free Tier**:
  - Unlimited deployments
  - 100GB bandwidth/month
  - Custom domain support

- **Upstash Free Tier**:
  - 10,000 messages/day
  - 10MB storage
  - Single replica

Perfect for a demo/assignment submission! 🎉

---

## Your Live URLs

After deployment, you'll have:

1. **Frontend (Share this!)**: `https://connect-four-[your-id].vercel.app`
2. **Backend API**: `https://connect-four-backend.onrender.com`
3. **Analytics Service**: Runs in background (no public URL needed)
4. **Kafka Dashboard**: `https://console.upstash.com` (to monitor events)

---

## Troubleshooting

### Backend won't start
- Check DATABASE_URL is correct (should start with `postgresql://`)
- Verify all Kafka credentials are correct
- Check Render logs for specific error

### Frontend can't connect to WebSocket
- Ensure REACT_APP_WS_URL uses `wss://` (not `ws://`)
- Check CORS settings in backend
- Verify backend is running (check `/api/health`)

### No events in Kafka
- Check KAFKA_ENABLED is set to `true`
- Verify Kafka credentials (username/password)
- Check backend logs for "Kafka producer initialized"

### Analytics not processing events
- Check analytics service logs for errors
- Verify it can connect to both database and Kafka
- Ensure KAFKA_GROUP_ID is unique

---

**Need help?** Check Render logs or Upstash dashboard for specific error messages!
