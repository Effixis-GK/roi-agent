@echo off
echo ===== ROI Agent Windows Build Script =====

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

REM Check if Python is installed
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Python is not installed or not in PATH
    echo Please install Python from https://python.org/downloads/
    pause
    exit /b 1
)

echo Building Go agent...
go mod tidy
go build -o roi-agent-windows.exe main.go
if %errorlevel% neq 0 (
    echo Error: Failed to build Go agent
    pause
    exit /b 1
)

echo Installing Python dependencies...
pip install -r requirements.txt
if %errorlevel% neq 0 (
    echo Error: Failed to install Python dependencies
    pause
    exit /b 1
)

echo Build completed successfully!
echo.
echo Files created:
echo - roi-agent-windows.exe (Windows monitoring agent)
echo - web_app.py (Web interface)
echo - templates/index.html (Web UI template)
echo.
echo To run:
echo 1. Start the agent: roi-agent-windows.exe
echo 2. Start the web UI: python web_app.py
echo 3. Open browser: http://localhost:5002
echo.
pause
