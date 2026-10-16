@echo off
setlocal enabledelayedexpansion

echo Building S-UI for Windows...

set "ROOT_DIR=%~dp0.."
cd /d "%ROOT_DIR%"

REM Read the shared build tags from build-tags.sh.
set "BUILD_TAGS="
for /f "tokens=1,* delims==" %%A in ('findstr /b "BASE_TAGS=" "%ROOT_DIR%\build-tags.sh"') do set "BUILD_TAGS=%%B"
set "BUILD_TAGS=!BUILD_TAGS:"=!"
if not defined BUILD_TAGS (
    echo Error: Could not read BASE_TAGS from build-tags.sh
    pause
    exit /b 1
)

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

REM Check if Node.js is installed
node --version >nul 2>&1
if errorlevel 1 (
    echo Error: Node.js is not installed or not in PATH
    echo Please install Node.js from https://nodejs.org/
    pause
    exit /b 1
)

echo Building frontend...
cd frontend
call npm ci
if errorlevel 1 (
    echo Error: Failed to install frontend dependencies
    pause
    exit /b 1
)

call npm run build
if errorlevel 1 (
    echo Error: Failed to build frontend
    pause
    exit /b 1
)

cd ..

echo Creating web/html directory...
if not exist "web\html" mkdir "web\html"

echo Copying frontend build files...
xcopy "frontend\dist\*" "web\html\" /E /Y /Q

echo Building backend...
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

go build -trimpath -ldflags "-w -s" -tags "!BUILD_TAGS!" -o sui.exe main.go
if errorlevel 1 (
    echo Error: Failed to build backend with CGO enabled
    pause
    exit /b 1
)
echo Built with CGO

echo Build completed successfully!
echo Output: sui.exe
pause
