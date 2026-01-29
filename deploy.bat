@echo off
echo ==========================================
echo   Deploy Connect Four to Vercel + Render
echo ==========================================
echo.

REM Check if git is initialized
if not exist .git (
    echo Step 1: Initializing git repository...
    git init
    git add .
    git commit -m "Initial commit - Connect Four game"
    echo [32mGit initialized[0m
) else (
    echo [32mGit already initialized[0m
)

echo.
echo Step 2: Push to GitHub
echo Choose option:
echo   1) Create new repo with GitHub CLI (gh)
echo   2) I'll do it manually
set /p choice="Enter choice (1 or 2): "

if "%choice%"=="1" (
    echo Creating GitHub repository...
    gh repo create connect-four --public --source=. --remote=origin --push
    echo [32mRepository created and pushed[0m
) else (
    echo.
    echo Manual steps:
    echo 1. Go to https://github.com/new
    echo 2. Create repository 'connect-four'
    echo 3. Run these commands:
    echo    git remote add origin https://github.com/YOUR_USERNAME/connect-four.git
    echo    git branch -M main
    echo    git push -u origin main
    echo.
    pause
)

echo.
echo ==========================================
echo   Backend Deployment (Render)
echo ==========================================
echo.
echo 1. Go to: https://dashboard.render.com/
echo 2. Click 'New +' -^> 'Web Service'
echo 3. Connect your GitHub repo
echo 4. Configure:
echo    - Name: connect4-backend
echo    - Root Directory: backend
echo    - Runtime: Docker
echo    - Instance Type: Free
echo 5. Add Environment Variable:
echo    PORT = 8080
echo 6. Create database:
echo    - New -^> PostgreSQL
echo    - Name: connect4-db
echo    - Copy Internal Database URL
echo    - Add to backend as DATABASE_URL
echo.
set /p BACKEND_URL="Enter your Render backend URL (e.g., https://connect4-backend.onrender.com): "

echo.
echo ==========================================
echo   Frontend Deployment (Vercel)
echo ==========================================
echo.

REM Check if vercel CLI is installed
where vercel >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Installing Vercel CLI...
    call npm install -g vercel
)

echo Logging in to Vercel...
call vercel login

echo.
echo Deploying frontend...
cd frontend

echo Setting environment variables...
echo %BACKEND_URL%/ws | vercel env add REACT_APP_WS_URL production
echo %BACKEND_URL% | vercel env add REACT_APP_API_URL production

echo Deploying to production...
call vercel --prod

cd ..

echo.
echo ==========================================
echo   [32m Deployment Complete![0m
echo ==========================================
echo.
echo Your app is now live!
echo.
echo Backend: %BACKEND_URL%
echo Frontend: Check Vercel dashboard for URL
echo.
echo Next steps:
echo 1. Test your live app
echo 2. Share the frontend URL
echo 3. Monitor on Render ^& Vercel dashboards
echo.
pause
