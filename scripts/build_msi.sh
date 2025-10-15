#!/bin/bash

# ROI Agent - Windows MSI Installer Builder
# ThousandEyes形式に倣ったMSIインストーラー生成
# Note: MSIビルドはWindows環境（WiXツールセット）で実行する必要があります
# このスクリプトはWiX XMLファイルとビルド構造を生成します

set -e

echo "📦 ROI Agent - Windows MSI Installer Builder"
echo "============================================="

# 設定
PROJECT_ROOT="/Users/taktakeu/Local/roi-agent"
BUILD_DIR="$PROJECT_ROOT/build/msi"
VERSION="${1:-1.0.4}"
ARCH="${2:-x64}"  # x64 または x86

# アーキテクチャ設定
if [ "$ARCH" = "x64" ]; then
    ARCH_LABEL="x64"
    PLATFORM="x64"
    PROGRAM_FILES="ProgramFiles64Folder"
elif [ "$ARCH" = "x86" ]; then
    ARCH_LABEL="x86"
    PLATFORM="x86"
    PROGRAM_FILES="ProgramFilesFolder"
else
    echo "❌ Invalid architecture: $ARCH (use 'x64' or 'x86')"
    exit 1
fi

echo "Building for: Windows $ARCH_LABEL"
echo "Version: $VERSION"
echo ""

# クリーンアップ
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# ソースファイル用ディレクトリ
SOURCE_DIR="$BUILD_DIR/source"
mkdir -p "$SOURCE_DIR"

echo "🔨 Building Windows binaries..."

# 1. Agent バイナリをビルド
cd "$PROJECT_ROOT/agent"
if [ "$ARCH" = "x64" ]; then
    GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "$SOURCE_DIR/roi-agent.exe" main.go
else
    GOOS=windows GOARCH=386 go build -ldflags="-s -w" -o "$SOURCE_DIR/roi-agent.exe" main.go
fi
echo "  ✅ roi-agent.exe ($ARCH_LABEL)"

# 2. Data Sender バイナリをビルド
cd "$PROJECT_ROOT/data-sender"
if [ "$ARCH" = "x64" ]; then
    GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "$SOURCE_DIR/data-sender.exe" .
else
    GOOS=windows GOARCH=386 go build -ldflags="-s -w" -o "$SOURCE_DIR/data-sender.exe" .
fi
echo "  ✅ data-sender.exe ($ARCH_LABEL)"

cd "$PROJECT_ROOT"

# 3. Windowsサービスラッパー作成
cat > "$SOURCE_DIR/service-wrapper.exe.txt" << 'EOF'
Note: service-wrapper.exe should be built from service-wrapper/main.go
This is a Windows Service wrapper for roi-agent
EOF

# 4. 設定ファイルテンプレート
cat > "$SOURCE_DIR/.env.template" << 'EOF'
# ROI Agent Configuration
ROI_AGENT_BASE_URL=__BASE_URL__
ROI_AGENT_API_KEY=__API_KEY__
ROI_AGENT_INTERVAL_MINUTES=10
EOF
echo "  ✅ .env.template"

# 5. インストール/アンインストールスクリプト
cat > "$SOURCE_DIR/install-service.bat" << 'EOF'
@echo off
REM ROI Agent - Install Service

echo Installing ROI Agent Service...

sc create "ROI Agent" binPath= "%ProgramFiles%\ROI Agent\service-wrapper.exe" start= auto displayname= "ROI Agent Monitoring Service"

if %errorlevel% == 0 (
    echo Service installed successfully
    sc start "ROI Agent"
    echo Service started
) else (
    echo Failed to install service
)

pause
EOF

cat > "$SOURCE_DIR/uninstall-service.bat" << 'EOF'
@echo off
REM ROI Agent - Uninstall Service

echo Stopping ROI Agent Service...
sc stop "ROI Agent"

echo Uninstalling ROI Agent Service...
sc delete "ROI Agent"

if %errorlevel% == 0 (
    echo Service uninstalled successfully
) else (
    echo Failed to uninstall service
)

pause
EOF

# 6. README
cat > "$SOURCE_DIR/README.txt" << EOF
ROI Agent - Windows Installer
==============================

Version: $VERSION
Architecture: $ARCH_LABEL

INSTALLATION:
This installer will install ROI Agent as a Windows Service:
  C:\Program Files\ROI Agent\

COMPONENTS:
- roi-agent.exe: Main monitoring agent
- data-sender.exe: Data transmission service
- service-wrapper.exe: Windows Service wrapper

USAGE:
After installation, the service starts automatically.
It will run in the background and start on system boot.

Manual control:
  services.msc -> Find "ROI Agent Monitoring Service"

Or command line:
  sc start "ROI Agent"
  sc stop "ROI Agent"

UNINSTALLATION:
Use Windows "Add or Remove Programs"
Or run: C:\Program Files\ROI Agent\uninstall-service.bat

SUPPORT:
Visit: https://roi-dashboard-607617540267.asia-northeast1.run.app
EOF
echo "  ✅ README.txt"

# 7. WiX XMLファイル生成
cat > "$BUILD_DIR/ROIAgent.wxs" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">
  <Product Id="*" 
           Name="ROI Agent" 
           Language="1033" 
           Version="$VERSION.0" 
           Manufacturer="Effixis" 
           UpgradeCode="12345678-1234-1234-1234-123456789012">
    
    <Package InstallerVersion="200" 
             Compressed="yes" 
             InstallScope="perMachine" 
             Platform="$PLATFORM" />

    <MajorUpgrade DowngradeErrorMessage="A newer version of ROI Agent is already installed." />
    
    <MediaTemplate EmbedCab="yes" />

    <Feature Id="ProductFeature" Title="ROI Agent" Level="1">
      <ComponentGroupRef Id="ProductComponents" />
      <ComponentRef Id="ServiceComponent" />
    </Feature>

    <Directory Id="TARGETDIR" Name="SourceDir">
      <Directory Id="$PROGRAM_FILES">
        <Directory Id="INSTALLFOLDER" Name="ROI Agent">
          <Directory Id="DataFolder" Name="data" />
          <Directory Id="LogsFolder" Name="logs" />
        </Directory>
      </Directory>
    </Directory>

    <ComponentGroup Id="ProductComponents" Directory="INSTALLFOLDER">
      <Component Id="AgentExe" Guid="*">
        <File Id="AgentExeFile" Source="source/roi-agent.exe" KeyPath="yes" />
      </Component>
      <Component Id="DataSenderExe" Guid="*">
        <File Id="DataSenderExeFile" Source="source/data-sender.exe" KeyPath="yes" />
      </Component>
      <Component Id="EnvTemplate" Guid="*">
        <File Id="EnvTemplateFile" Source="source/.env.template" KeyPath="yes" />
      </Component>
      <Component Id="ReadmeFile" Guid="*">
        <File Id="ReadmeTxt" Source="source/README.txt" KeyPath="yes" />
      </Component>
      <Component Id="InstallScript" Guid="*">
        <File Id="InstallBat" Source="source/install-service.bat" KeyPath="yes" />
      </Component>
      <Component Id="UninstallScript" Guid="*">
        <File Id="UninstallBat" Source="source/uninstall-service.bat" KeyPath="yes" />
      </Component>
    </ComponentGroup>

    <Component Id="ServiceComponent" Directory="INSTALLFOLDER" Guid="*">
      <ServiceInstall Id="ROIAgentService"
                      Type="ownProcess"
                      Name="ROI Agent"
                      DisplayName="ROI Agent Monitoring Service"
                      Description="ROI Agent - Application and Network Monitoring"
                      Start="auto"
                      Account="LocalSystem"
                      ErrorControl="normal"
                      Arguments=""
                      Interactive="no">
        <ServiceDependency Id="Tcpip" />
      </ServiceInstall>
      
      <ServiceControl Id="StartService"
                      Start="install"
                      Stop="both"
                      Remove="uninstall"
                      Name="ROI Agent"
                      Wait="yes" />
    </Component>

  </Product>
</Wix>
EOF
echo "  ✅ WiX XML generated"

# 8. ビルド指示書作成
cat > "$BUILD_DIR/BUILD_INSTRUCTIONS.md" << EOF
# Windows MSI Build Instructions

## Prerequisites
1. Install WiX Toolset 3.11 or later
   Download: https://wixtoolset.org/releases/

2. Add WiX to PATH:
   C:\Program Files (x86)\WiX Toolset v3.11\bin

## Build Steps

### From Windows Command Prompt:

\`\`\`cmd
cd $BUILD_DIR

REM Compile WiX source
candle ROIAgent.wxs -arch $PLATFORM

REM Link and create MSI
light ROIAgent.wixobj -out ROI-Agent-Windows-$ARCH_LABEL-$VERSION.msi

REM Calculate SHA256
certutil -hashfile ROI-Agent-Windows-$ARCH_LABEL-$VERSION.msi SHA256 > ROI-Agent-Windows-$ARCH_LABEL-$VERSION.msi.sha256
\`\`\`

### From PowerShell:

\`\`\`powershell
cd $BUILD_DIR

# Compile
& "C:\Program Files (x86)\WiX Toolset v3.11\bin\candle.exe" ROIAgent.wxs -arch $PLATFORM

# Link
& "C:\Program Files (x86)\WiX Toolset v3.11\bin\light.exe" ROIAgent.wixobj -out ROI-Agent-Windows-$ARCH_LABEL-$VERSION.msi

# SHA256
Get-FileHash ROI-Agent-Windows-$ARCH_LABEL-$VERSION.msi -Algorithm SHA256 | Select-Object Hash | Out-File ROI-Agent-Windows-$ARCH_LABEL-$VERSION.msi.sha256
\`\`\`

## Output Files
- ROI-Agent-Windows-$ARCH_LABEL-$VERSION.msi
- ROI-Agent-Windows-$ARCH_LABEL-$VERSION.msi.sha256

## Upload to Cloud Storage

\`\`\`bash
gsutil cp ROI-Agent-Windows-$ARCH_LABEL-$VERSION.msi gs://roi-agent-releases/latest/windows/
gsutil cp ROI-Agent-Windows-$ARCH_LABEL-$VERSION.msi.sha256 gs://roi-agent-releases/latest/windows/
\`\`\`
EOF

echo ""
echo "✅ Build preparation completed!"
echo ""
echo "📊 Generated files:"
echo "   Source files: $SOURCE_DIR"
echo "   WiX XML: $BUILD_DIR/ROIAgent.wxs"
echo "   Instructions: $BUILD_DIR/BUILD_INSTRUCTIONS.md"
echo ""
echo "⚠️  Note: MSI build requires Windows environment with WiX Toolset"
echo "    Follow instructions in BUILD_INSTRUCTIONS.md"
