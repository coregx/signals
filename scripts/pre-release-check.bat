@echo off
REM Pre-Release Check - Windows Wrapper
REM This script calls the bash version for consistency

REM Check if bash is available
where bash >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: bash not found in PATH
    echo.
    echo Please install Git Bash or WSL to run this script
    echo Download Git for Windows: https://git-scm.com/download/win
    echo.
    pause
    exit /b 1
)

REM Get script directory
set SCRIPT_DIR=%~dp0

REM Run bash script
bash "%SCRIPT_DIR%pre-release-check.sh"
exit /b %errorlevel%
