# ROI Agent - Application Monitor

A lightweight productivity monitoring system for macOS that tracks application usage patterns to help analyze and optimize your time investment (ROI - Return on Investment of time).

## Features

- **Real-time Application Monitoring**: 15-second interval tracking of all running applications
- **Three Usage Categories**: 
  - **Foreground time**: Applications actively visible and running
  - **Background time**: Applications running but not in focus
  - **Focus time**: Applications with active window focus
- **Web Dashboard**: Clean, responsive interface at localhost:5002
- **Daily Analytics**: Automatic daily data collection and usage rankings
- **Productivity Insights**: Understand where your time is being invested

## Architecture

Built with enterprise-grade architecture inspired by Datadog:
- **Go Agent**: High-performance monitoring agent for system data collection
- **Python Flask Web UI**: Modern web interface for data visualization
- **JSON Data Storage**: Simple, readable data format for easy integration
- **RESTful API**: Clean API endpoints for data access and automation

## Requirements

- macOS 10.14 or later
- Go 1.21+ (for building the agent)
- Python 3.8+ (for web UI)
- Accessibility permissions (automatically prompted)

## Quick Start

### 1. Clone and Setup
```bash
git clone <your-repo-url>
cd roi-agent
chmod +x *.sh
```

### 2. Build the Agent
```bash
./build_agent.sh
```

### 3. Start Web UI
```bash
./start_web.sh
```

### 4. Grant Permissions
When first running, the system will prompt for accessibility permissions:
1. Go to System Preferences > Security & Privacy > Privacy
2. Select "Accessibility" from the left panel
3. Click the lock and enter your password
4. Add the monitor application to the list

### 5. Start Monitoring
```bash
# In a new terminal
cd agent
./monitor
```

### 6. View Dashboard
Open http://localhost:5002 in your browser

## Development & Testing

### Generate Test Data
```bash
python debug_tools.py testdata
```

### System Diagnostics
```bash
python debug_tools.py full           # Complete diagnostic
python debug_tools.py system         # System requirements
python debug_tools.py permissions    # Accessibility check
python debug_tools.py apps          # App detection test
python debug_tools.py agent         # Agent communication
```

## Project Structure

```
roi-agent/
├── .gitignore                    # Git ignore rules
├── agent/                        # Go monitoring agent
│   ├── main.go                   # Agent source code
│   ├── go.mod                    # Go module definition
│   └── monitor                   # Compiled binary (ignored)
├── web/                          # Python Flask web UI
│   ├── app.py                    # Flask application
│   ├── requirements.txt          # Python dependencies
│   ├── venv/                     # Virtual environment (ignored)
│   └── templates/
│       └── index.html            # Web dashboard
├── config/                       # Configuration files
├── data/                         # Daily usage data (ignored)
├── logs/                         # Application logs (ignored)
├── debug_tools.py                # Debug utilities
├── build_agent.sh                # Agent build script
├── start_web.sh                  # Web UI startup script
└── README.md                     # This file
```

## API Endpoints

- `GET /` - Main dashboard
- `GET /api/status` - Agent status
- `GET /api/data` - Usage data (supports date and category filters)
- `GET /api/dates` - Available data dates

### Example API Usage
```bash
# Get today's data
curl http://localhost:5002/api/data

# Get specific date and category
curl "http://localhost:5002/api/data?date=2024-12-17&category=focus_time"

# Check agent status
curl http://localhost:5002/api/status
```

## Data Format

Daily usage data is stored as JSON files in the `data/` directory:

```json
{
  "date": "2024-12-17",
  "apps": {
    "Safari": {
      "name": "Safari",
      "foreground_time": 7200,
      "background_time": 900,
      "focus_time": 6300,
      "last_seen": "2024-12-17T15:30:00",
      "is_active": true,
      "is_focused": false
    }
  },
  "total": {
    "foreground_time": 16200,
    "background_time": 3300,
    "focus_time": 13800
  }
}
```

## Git Workflow

The `.gitignore` file excludes:
- Compiled binaries (`agent/monitor`)
- Virtual environments (`web/venv/`)
- Personal data files (`data/`, `logs/`)
- OS-specific files (`.DS_Store`, etc.)
- IDE configuration files

Only source code and configuration files are tracked in Git.

## Troubleshooting

### Agent Won't Start
1. Check accessibility permissions: `python debug_tools.py permissions`
2. Verify Go installation: `go version`
3. Rebuild agent: `./build_agent.sh`

### Web UI Not Loading
1. Check Python environment: `python3 --version`
2. Install dependencies: `cd web && pip install -r requirements.txt`
3. Check port availability: `lsof -i :5002`

### No Data Collected
1. Verify agent is running: Check terminal output
2. Test app detection: `python debug_tools.py apps`
3. Generate test data: `python debug_tools.py testdata`

### Permission Errors
1. Grant accessibility permissions in System Preferences
2. Ensure proper file permissions in project directory
3. Check macOS security settings

## Productivity Analysis

Use ROI Agent to:
- **Identify time sinks**: See which applications consume most of your time
- **Analyze focus patterns**: Understand when you're most productive
- **Track daily progress**: Monitor changes in usage patterns over time
- **Optimize workflows**: Identify opportunities to reduce context switching

## Contributing

This project is designed for personal productivity analysis. Feel free to extend functionality:

- Add Windows/Linux support
- Implement data export features
- Create detailed analytics and reporting
- Add productivity scoring algorithms
- Integrate with time-tracking tools

## Privacy

All data is stored locally on your machine. No information is transmitted to external servers. The `data/` directory contains your personal usage patterns and is excluded from Git tracking.

## License

This project is for educational and personal productivity purposes.
