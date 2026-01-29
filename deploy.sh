#!/bin/bash

echo "=========================================="
echo "  Deploy Connect Four to Vercel + Render"
echo "=========================================="
echo ""

# Check if git is initialized
if [ ! -d .git ]; then
    echo "Step 1: Initializing git repository..."
    git init
    git add .
    git commit -m "Initial commit - Connect Four game"
    echo "✓ Git initialized"
else
    echo "✓ Git already initialized"
fi

echo ""
echo "Step 2: Push to GitHub"
echo "Choose option:"
echo "  1) Create new repo with GitHub CLI (gh)"
echo "  2) I'll do it manually"
read -p "Enter choice (1 or 2): " choice

if [ "$choice" = "1" ]; then
    echo "Creating GitHub repository..."
    gh repo create connect-four --public --source=. --remote=origin --push
    echo "✓ Repository created and pushed"
else
    echo ""
    echo "Manual steps:"
    echo "1. Go to https://github.com/new"
    echo "2. Create repository 'connect-four'"
    echo "3. Run these commands:"
    echo "   git remote add origin https://github.com/YOUR_USERNAME/connect-four.git"
    echo "   git branch -M main"
    echo "   git push -u origin main"
    echo ""
    read -p "Press Enter when done..."
fi

echo ""
echo "=========================================="
echo "  Backend Deployment (Render)"
echo "=========================================="
echo ""
echo "1. Go to: https://dashboard.render.com/"
echo "2. Click 'New +' → 'Web Service'"
echo "3. Connect your GitHub repo"
echo "4. Configure:"
echo "   - Name: connect4-backend"
echo "   - Root Directory: backend"
echo "   - Runtime: Docker"
echo "   - Instance Type: Free"
echo "5. Add Environment Variable:"
echo "   PORT = 8080"
echo "6. Create database:"
echo "   - New → PostgreSQL"
echo "   - Name: connect4-db"
echo "   - Copy Internal Database URL"
echo "   - Add to backend as DATABASE_URL"
echo ""
read -p "Enter your Render backend URL (e.g., https://connect4-backend.onrender.com): " BACKEND_URL

echo ""
echo "=========================================="
echo "  Frontend Deployment (Vercel)"
echo "=========================================="
echo ""

# Check if vercel CLI is installed
if ! command -v vercel &> /dev/null; then
    echo "Installing Vercel CLI..."
    npm install -g vercel
fi

echo "Logging in to Vercel..."
vercel login

echo ""
echo "Deploying frontend..."
cd frontend

# Set environment variables
echo "Setting environment variables..."
echo "$BACKEND_URL/ws" | vercel env add REACT_APP_WS_URL production
echo "$BACKEND_URL" | vercel env add REACT_APP_API_URL production

echo "Deploying to production..."
vercel --prod

echo ""
echo "=========================================="
echo "  🎉 Deployment Complete!"
echo "=========================================="
echo ""
echo "Your app is now live!"
echo ""
echo "Backend: $BACKEND_URL"
echo "Frontend: Check Vercel dashboard for URL"
echo ""
echo "Next steps:"
echo "1. Test your live app"
echo "2. Share the frontend URL"
echo "3. Monitor on Render & Vercel dashboards"
echo ""
