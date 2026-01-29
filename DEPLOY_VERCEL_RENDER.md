# 🚀 Deployment Guide: Vercel + Render

## Overview
- **Frontend**: Vercel (React app)
- **Backend**: Render (GoLang API + Database)

---

## Part 1: Deploy Backend to Render

### Step 1: Push to GitHub

```bash
cd "c:\Users\kushw\Downloads\Connect-four"
git init
git add .
git commit -m "Initial commit - Connect Four game"
```

Create GitHub repository:
```bash
# Using GitHub CLI
gh repo create connect-four --public --source=. --remote=origin --push

# OR manually:
# 1. Go to https://github.com/new
# 2. Create repository "connect-four"
# 3. Run these commands:
git remote add origin https://github.com/YOUR_USERNAME/connect-four.git
git branch -M main
git push -u origin main
```

### Step 2: Deploy Backend on Render

1. **Go to Render Dashboard**: https://dashboard.render.com/

2. **New → Web Service**

3. **Connect GitHub Repository**: Select "connect-four"

4. **Configure Backend Service**:
   - **Name**: `connect4-backend`
   - **Region**: Choose closest to you
   - **Branch**: `main`
   - **Root Directory**: `backend`
   - **Runtime**: `Docker`
   - **Docker Build Context**: `backend`
   - **Dockerfile Path**: `./Dockerfile`
   - **Instance Type**: `Free`

5. **Add Environment Variables**:
   ```
   PORT = 8080
   ```

6. **Click "Create Web Service"**

7. **Wait for deployment** (5-10 minutes)

8. **Copy your backend URL**: 
   - Will be like: `https://connect4-backend.onrender.com`

### Step 3: Add PostgreSQL Database

1. In Render Dashboard: **New → PostgreSQL**

2. **Configure**:
   - **Name**: `connect4-db`
   - **Database**: `connectfour`
   - **User**: `postgres`
   - **Region**: Same as backend
   - **Instance Type**: `Free`

3. **Create Database**

4. **Connect to Backend**:
   - Go to your backend service settings
   - Add Environment Variable:
     ```
     DATABASE_URL = [Copy Internal Database URL from PostgreSQL service]
     ```

5. **Redeploy backend** (it will auto-redeploy)

---

## Part 2: Deploy Frontend to Vercel

### Step 1: Update Environment Variables

Before deploying, update the backend URLs in `frontend/vercel.json`:

Replace `connect4-backend.onrender.com` with your actual Render backend URL.

### Step 2: Deploy to Vercel

**Option A: Using Vercel Dashboard (Easiest)**

1. **Go to**: https://vercel.com/new

2. **Import Git Repository**
   - Click "Add GitHub Account" if needed
   - Select your "connect-four" repository

3. **Configure Project**:
   - **Framework Preset**: `Create React App`
   - **Root Directory**: `frontend`
   - **Build Command**: `npm run build`
   - **Output Directory**: `build`

4. **Environment Variables**:
   ```
   REACT_APP_WS_URL = wss://YOUR_BACKEND.onrender.com/ws
   REACT_APP_API_URL = https://YOUR_BACKEND.onrender.com
   ```
   Replace `YOUR_BACKEND` with your actual Render URL

5. **Click "Deploy"**

6. **Wait 2-3 minutes**

7. **Get your live URL**: `https://your-app.vercel.app`

**Option B: Using Vercel CLI**

```bash
# Install Vercel CLI
npm install -g vercel

# Login
vercel login

# Deploy from frontend folder
cd frontend
vercel

# Follow prompts:
# - Link to existing project? No
# - Project name: connect4-frontend
# - Directory: ./
# - Override settings? No

# Set environment variables
vercel env add REACT_APP_WS_URL
# Enter: wss://YOUR_BACKEND.onrender.com/ws

vercel env add REACT_APP_API_URL
# Enter: https://YOUR_BACKEND.onrender.com

# Deploy to production
vercel --prod
```

---

## Part 3: Final Configuration

### Update Backend CORS

After deploying frontend, update backend to allow your Vercel domain.

In `backend/main.go`, the CORS is already set to allow all origins (`*`).

For production, you should restrict it:

```go
c := cors.New(cors.Options{
    AllowedOrigins:   []string{"https://your-app.vercel.app"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"*"},
    AllowCredentials: true,
})
```

Then commit and push to trigger redeployment.

---

## Your Live URLs

After deployment, you'll have:

- **Frontend (Game UI)**: `https://your-app.vercel.app`
- **Backend API**: `https://connect4-backend.onrender.com`
- **Database**: Internal connection via Render

---

## Testing Your Live App

1. **Open frontend URL** in browser
2. **Enter username** and join game
3. **Open in another tab/browser** to test multiplayer
4. **Check leaderboard** after playing

---

## Troubleshooting

### Frontend can't connect to backend

**Check**:
1. Backend is running on Render (check dashboard)
2. Environment variables are correct in Vercel
3. CORS allows your Vercel domain
4. WebSocket URL uses `wss://` (not `ws://`)

**Fix**: Redeploy frontend with correct env variables:
```bash
vercel --prod
```

### Backend errors

**Check logs on Render**:
1. Go to Render dashboard
2. Click on backend service
3. View "Logs" tab

### Database connection failed

**Check**:
1. DATABASE_URL is set correctly in backend env vars
2. Database is running on Render
3. Use "Internal Database URL" (not external)

---

## Monitoring

**Render Dashboard**: https://dashboard.render.com/
- View logs
- Monitor usage
- Check health

**Vercel Dashboard**: https://vercel.com/dashboard
- View deployments
- Check analytics
- Monitor performance

---

## Free Tier Limits

**Render Free Tier**:
- Backend sleeps after 15 min inactivity
- First request after sleep takes 30s to wake up
- 750 hours/month

**Vercel Free Tier**:
- Unlimited deployments
- 100GB bandwidth/month
- No sleep time

**Note**: For production, upgrade to paid plans for better performance.

---

## Updating Your App

**Frontend changes**:
```bash
git add .
git commit -m "Update frontend"
git push
# Vercel auto-deploys
```

**Backend changes**:
```bash
git add .
git commit -m "Update backend"
git push
# Render auto-deploys
```

---

## Cost: $0/month (Free Tier)

Both services offer generous free tiers perfect for this project!

---

**Ready to deploy!** Follow the steps above to get your live link! 🚀
