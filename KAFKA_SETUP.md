# Kafka Integration Setup Guide

## Option 1: Upstash Kafka (Recommended - Free Tier Available)

### Step 1: Create Upstash Kafka Cluster

1. Go to [upstash.com](https://upstash.com) and sign up
2. Click **"Create Cluster"** in the Kafka section
3. Choose:
   - **Name**: `connect-four-kafka`
   - **Region**: Choose closest to your Render region
   - **Type**: Single Replica (free tier)
4. Click **"Create"**

### Step 2: Create Kafka Topic

1. In your cluster, go to **"Topics"**
2. Click **"Create Topic"**
3. Configure:
   - **Topic Name**: `game-events`
   - **Partitions**: 1
   - **Retention Time**: 7 days
4. Click **"Create"**

### Step 3: Get Connection Details

From your Upstash cluster dashboard, copy:
- **Bootstrap Server** (e.g., `smart-mantis-12345-us1-kafka.upstash.io:9092`)
- **Username** (looks like a long string)
- **Password** (looks like a long string)
- **SASL Mechanism**: SCRAM-SHA-256 or SCRAM-SHA-512

## Option 2: CloudKarafka (Alternative)

1. Go to [cloudkarafka.com](https://www.cloudkarafka.com)
2. Sign up and create a free plan
3. Create topic `game-events`
4. Copy KAFKA_URL from dashboard

## Option 3: Confluent Cloud (Enterprise-grade)

1. Go to [confluent.cloud](https://confluent.cloud)
2. Sign up for free trial ($400 credit)
3. Create cluster and topic
4. Generate API keys

## Environment Variables for Backend

Add these to your Render backend service:

```env
KAFKA_ENABLED=true
KAFKA_BROKERS=smart-mantis-12345-us1-kafka.upstash.io:9092
KAFKA_TOPIC=game-events
KAFKA_USERNAME=your-username-from-upstash
KAFKA_PASSWORD=your-password-from-upstash
KAFKA_SASL_MECHANISM=SCRAM-SHA-512
```

## Environment Variables for Analytics Service

Add these to your Render analytics service:

```env
DATABASE_URL=your-database-url (same as backend)
KAFKA_BROKERS=smart-mantis-12345-us1-kafka.upstash.io:9092
KAFKA_TOPIC=game-events
KAFKA_GROUP_ID=analytics-consumer
KAFKA_USERNAME=your-username-from-upstash
KAFKA_PASSWORD=your-password-from-upstash
KAFKA_SASL_MECHANISM=SCRAM-SHA-512
```

## Deploy Analytics Service to Render

1. Go to Render dashboard
2. Click **"New +"** → **"Web Service"**
3. Connect repository: `Arpitkushwahaa/Connect-Four`
4. Configure:
   - **Name**: `connect-four-analytics`
   - **Root Directory**: `analytics`
   - **Runtime**: `Go`
   - **Build Command**: `go build -o analytics .`
   - **Start Command**: `./analytics`
5. Add all environment variables listed above
6. Click **"Create Web Service"**

## Testing Kafka Integration

Once deployed, Kafka will automatically:
- Emit events when games start (`game_start`)
- Emit events for each move (`game_move`)
- Emit events when games end (`game_end`)
- Analytics service consumes and processes these events
- Updates analytics_summary table in PostgreSQL

## Verify It's Working

Check Render logs:
- **Backend logs**: Should show "Successfully published event to Kafka"
- **Analytics logs**: Should show "Processing event: game_start/game_move/game_end"

Query analytics from leaderboard API:
```bash
curl https://your-backend-url.onrender.com/leaderboard
```
