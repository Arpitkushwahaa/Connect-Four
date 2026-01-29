@echo off
echo ========================================
echo Deploying 4 in a Row to Railway
echo ========================================
echo.

echo Step 1: Installing Railway CLI...
call npm install -g @railway/cli

echo.
echo Step 2: Logging in to Railway...
call railway login

echo.
echo Step 3: Initializing project...
call railway init

echo.
echo Step 4: Deploying application...
call railway up

echo.
echo ========================================
echo Deployment complete!
echo Check Railway dashboard for your URL
echo ========================================
pause
