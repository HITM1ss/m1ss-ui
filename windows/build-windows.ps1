# PowerShell script for building S-UI on Windows
param(
    [string]$Architecture = "amd64",
    [switch]$NoCGO,
    [switch]$Help
)

if ($Help) {
    Write-Host "Usage: .\build-windows.ps1 [-Architecture <arch>] [-Help]"
    Write-Host "Architectures: amd64, 386, arm64"
    Write-Host "Examples:"
    Write-Host "  .\build-windows.ps1                    # Build for amd64 with CGO"
    Write-Host "  .\build-windows.ps1 -Architecture 386 # Build for 32-bit Windows"
    exit 0
}

if ($NoCGO) {
    Write-Host "Error: -NoCGO is not supported because SQLite requires CGO." -ForegroundColor Red
    exit 1
}

Write-Host "Building S-UI for Windows ($Architecture)..." -ForegroundColor Green

# Resolve the repository root from this script's location.
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

# Read the shared build tags from build-tags.sh.
$tagLine = Get-Content (Join-Path $repoRoot "build-tags.sh") | Where-Object { $_ -match '^BASE_TAGS="([^"]+)"$' } | Select-Object -First 1
if ($null -eq $tagLine -or $tagLine -notmatch '^BASE_TAGS="([^"]+)"$') {
    Write-Host "Error: Could not read BASE_TAGS from build-tags.sh" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}
$buildTags = $Matches[1]

# Check if Go is installed
try {
    $goVersion = go version 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Go not found"
    }
    Write-Host "Go version: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "Error: Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go from https://golang.org/dl/" -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

# Check if Node.js is installed
try {
    $nodeVersion = node --version 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Node.js not found"
    }
    Write-Host "Node.js version: $nodeVersion" -ForegroundColor Green
} catch {
    Write-Host "Error: Node.js is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Node.js from https://nodejs.org/" -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

# Build frontend
Write-Host "Building frontend..." -ForegroundColor Yellow
Push-Location frontend

try {
    Write-Host "Installing dependencies..." -ForegroundColor Cyan
    npm ci
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to install frontend dependencies"
    }

    Write-Host "Building frontend..." -ForegroundColor Cyan
    npm run build
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to build frontend"
    }
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    Pop-Location
    Read-Host "Press Enter to exit"
    exit 1
}

Pop-Location

# Create web/html directory
Write-Host "Creating web/html directory..." -ForegroundColor Yellow
if (!(Test-Path "web\html")) {
    New-Item -ItemType Directory -Path "web\html" -Force | Out-Null
}

# Copy frontend build files
Write-Host "Copying frontend build files..." -ForegroundColor Yellow
Copy-Item "frontend\dist\*" "web\html\" -Recurse -Force

# Build backend
Write-Host "Building backend..." -ForegroundColor Yellow

# Set environment variables
$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = $Architecture

Write-Host "Building with CGO..." -ForegroundColor Yellow

# Build command

try {
    & go build -trimpath -ldflags '-w -s' -tags $buildTags -o sui.exe main.go
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to build backend with CGO enabled"
    } else {
        Write-Host "Built successfully with CGO" -ForegroundColor Green
    }
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host "Build completed successfully!" -ForegroundColor Green
Write-Host "Output: sui.exe" -ForegroundColor Green

# Show file info
if (Test-Path "sui.exe") {
    $fileInfo = Get-Item "sui.exe"
    Write-Host "File size: $([math]::Round($fileInfo.Length / 1MB, 2)) MB" -ForegroundColor Cyan
    Write-Host "Created: $($fileInfo.CreationTime)" -ForegroundColor Cyan
}

Read-Host "Press Enter to exit"
