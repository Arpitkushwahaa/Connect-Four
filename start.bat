@echo off
echo ======================================
echo  4 in a Row - Connect Four Game
echo  Starting all services...
echo ======================================
echo.

echo Checking Docker...
docker --version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Docker is not installed or not running!
    echo Please install Docker Desktop and make sure it's running.
    pause
    exit /b 1
)

echo Docker is running!
echo.

echo Starting services with Docker Compose...
docker-compose up -d

if errorlevel 1 (
    echo.
    echo ERROR: Failed to start services!
    echo Check the error messages above.
    pause
    exit /b 1
)

echo.
echo ======================================
echo  Services are starting...
echo  Please wait 30-60 seconds for all
echo  services to be fully ready.
echo ======================================
echo.
echo  Frontend:  http://localhost:3000
echo  Backend:   http://localhost:8080
echo  WebSocket: ws://localhost:8080/ws
echo.
echo To view logs: docker-compose logs -f
echo To stop:      docker-compose down
echo.
pause
