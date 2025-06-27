@echo off
echo ===== ROI Agent Windows Startup =====

REM Check if executable exists
if not exist "roi-agent-windows.exe" (
    echo Error: roi-agent-windows.exe not found
    echo Please run build.bat first to build the application
    pause
    exit /b 1
)

REM Check if web app exists
if not exist "web_app.py" (
    echo Error: web_app.py not found
    echo Please ensure all files are in the correct location
    pause
    exit /b 1
)

echo Starting ROI Agent Windows...
echo.
echo This will start both:
echo 1. Background monitoring agent
echo 2. Web interface on http://localhost:5002
echo.
echo Press Ctrl+C to stop both services
echo.

REM Start the agent in background
echo Starting monitoring agent...
start /B roi-agent-windows.exe

REM Wait a moment for agent to start
timeout /t 2 /nobreak >nul

REM Start the web interface
echo Starting web interface...
echo Open your browser to: http://localhost:5002
echo.
python web_app.py

REM If we get here, web app was stopped, so stop the agent too
echo.
echo Stopping services...
taskkill /f /im roi-agent-windows.exe >nul 2>&1
echo ROI Agent stopped.
pause
