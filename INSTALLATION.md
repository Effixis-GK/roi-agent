# ROI Agent - Download & Installation Guide

## 🚀 Quick Start (Automated)

### Option 1: One-Command Build & Install
```bash
cd /Users/taktakeu/Local/GitHub/roi-agent
chmod +x quick_setup.sh
./quick_setup.sh
```

### Option 2: Manual Step-by-Step Build

```bash
# 1. Build Go agent
./build_agent.sh

# 2. Create macOS app
./build_app.sh

# 3. Create DMG installer
./create_dmg.sh

# 4. Install to Applications
cp -R "build/ROI Agent.app" /Applications/
```

## 📦 Distribution Methods

### Method 1: DMG Installer (Recommended)
1. **Create DMG**: `./create_dmg.sh`
2. **Share**: Send `build/ROI-Agent-Installer.dmg` to users
3. **Install**: Users double-click DMG and drag app to Applications

### Method 2: Direct App Bundle
1. **Build**: `./build_app.sh`
2. **Zip**: `cd build && zip -r "ROI-Agent.zip" "ROI Agent.app"`
3. **Share**: Send ZIP file to users
4. **Install**: Users unzip and move to Applications

### Method 3: GitHub Release
1. **Tag release**: `git tag v1.0.0 && git push origin v1.0.0`
2. **Upload DMG**: Attach DMG to GitHub release
3. **Users download**: From GitHub releases page

## 🔽 Download Instructions for End Users

### For Users Receiving the App:

#### If you receive a DMG file:
1. **Double-click** `ROI-Agent-Installer.dmg`
2. **Drag** "ROI Agent" to Applications folder
3. **Eject** the DMG
4. **Launch** ROI Agent from Applications
5. **Grant permissions** when prompted

#### If you receive a ZIP file:
1. **Double-click** the ZIP to extract
2. **Move** "ROI Agent.app" to Applications folder
3. **Launch** ROI Agent from Applications
4. **Grant permissions** when prompted

## 🛠 Manual Installation

### Prerequisites
- macOS 10.14 or later
- Administrator access for first-time permissions

### Installation Steps
1. **Download** or build the app
2. **Move to Applications**:
   ```bash
   mv "ROI Agent.app" /Applications/
   ```
3. **First Launch**:
   ```bash
   open "/Applications/ROI Agent.app"
   ```
4. **Grant Permissions**:
   - System Preferences → Security & Privacy → Privacy
   - Select "Accessibility" → Add ROI Agent

## 🎯 Post-Installation Usage

### Starting ROI Agent
- **GUI**: Double-click in Applications
- **Terminal**: `/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent`

### Accessing Dashboard
- **Auto-open**: Dashboard opens when launching app
- **Manual**: Visit http://localhost:5002
- **Command**: `/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent dashboard`

### Background Operations
- **Start monitoring**: `/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent start`
- **Stop monitoring**: `/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent stop`
- **Check status**: `/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent status`

## 📊 Data Storage
- **User data**: `~/.roiagent/data/`
- **Logs**: `~/.roiagent/logs/`
- **Daily files**: `usage_YYYY-MM-DD.json`

## 🔧 Troubleshooting

### Common Issues

#### "ROI Agent can't be opened"
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine "/Applications/ROI Agent.app"
```

#### Permission denied
```bash
# Fix permissions
chmod +x "/Applications/ROI Agent.app/Contents/MacOS/roi-agent"
chmod +x "/Applications/ROI Agent.app/Contents/MacOS/monitor"
```

#### No data collected
1. Check accessibility permissions in System Preferences
2. Restart the app after granting permissions
3. Check status: `/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent status`

#### Web interface not accessible
1. Check if services are running: `roi-agent status`
2. Restart services: `roi-agent restart`
3. Check port 5002: `lsof -i :5002`

### Debug Commands
```bash
# Full system check
cd /Users/taktakeu/Local/GitHub/roi-agent
python debug_tools.py full

# Generate test data
python debug_tools.py testdata

# Check processes
ps aux | grep roi-agent
ps aux | grep monitor
```

## 🗑 Uninstallation

### Complete Removal
```bash
# Stop services
/Applications/ROI\ Agent.app/Contents/MacOS/roi-agent stop

# Remove app
rm -rf "/Applications/ROI Agent.app"

# Remove data (optional)
rm -rf ~/.roiagent
```

### Keep Data
```bash
# Remove only the app
rm -rf "/Applications/ROI Agent.app"
# Data remains in ~/.roiagent for future use
```

## 📋 System Requirements
- **OS**: macOS 10.14 (Mojave) or later
- **Architecture**: Intel or Apple Silicon (Universal binary)
- **RAM**: 50MB minimum
- **Storage**: 100MB for app + data
- **Permissions**: Accessibility access required

## 🔐 Security & Privacy
- **Local data only**: No data transmitted externally
- **Accessibility required**: To monitor application usage
- **Open source**: Code available for review
- **No network access**: Runs completely offline

## 📞 Support
- **Documentation**: README.md
- **Issues**: GitHub Issues page
- **Debug tools**: Built-in diagnostic commands
- **Logs**: `~/.roiagent/logs/` for troubleshooting
