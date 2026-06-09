<#
.SYNOPSIS
eino-stock Build and Dev Script for Windows
.DESCRIPTION
Build, run, and generate code for eino-stock. Uses absolute paths based on script location.
.PARAMETER Target
  frontend      Build frontend (vite build + copy to embed dir)
  backend       Build backend binary
  wire          Regenerate wire dependency injection
  dev           Start backend + frontend dev servers
  run           Start backend production binary
  run-frontend  Start Vite dev server only
  run-backend   Start backend (go run) only
  test          Run all tests
  clean         Clean build artifacts
  all           Build frontend + backend (default)
#>
param([string]$Target = "all")

$ProjectRoot = Split-Path $PSScriptRoot -Parent
$Version = (Get-Date -Format "yyyy.MM.dd.HHmm")
$Module = "eino-stock"
$BuildDir = Join-Path $ProjectRoot "build"
$EmbedDir = Join-Path $ProjectRoot "internal/server/web"
$FrontendDir = Join-Path $ProjectRoot "frontend"

function Write-Step { param($Msg) Write-Host "==> $Msg" -ForegroundColor Cyan }

switch ($Target.ToLower()) {
    "clean" {
        Write-Step "Cleaning..."
        if (Test-Path $BuildDir) { Remove-Item $BuildDir -Recurse -Force }
        Write-Host "  Cleaned $BuildDir/"
    }
    "frontend" {
        Write-Step "Building frontend..."
        Push-Location $FrontendDir
        npm install 2>$null
        npm run build
        Pop-Location
        if (Test-Path $EmbedDir) { Remove-Item $EmbedDir -Recurse -Force }
        Copy-Item (Join-Path $FrontendDir "dist") -Destination $EmbedDir -Recurse
        Write-Host "  Frontend built -> $EmbedDir/"
    }
    "backend" {
        Write-Step "Building backend (version: $Version)..."
        $env:GOOS = "windows"; $env:GOARCH = "amd64"
        if (!(Test-Path $BuildDir)) { New-Item -ItemType Directory -Path $BuildDir | Out-Null }
        go build -ldflags "-X main.Version=$Version -X main.Name=$Module" -o (Join-Path $BuildDir "${Module}.exe") (Join-Path $ProjectRoot "cmd/${Module}/")
        Write-Host "  Built: $(Join-Path $BuildDir "${Module}.exe")"
    }
    "wire" {
        Write-Step "Regenerating Wire DI..."
        Push-Location $ProjectRoot
        go run github.com/google/wire/cmd/wire ./cmd/${Module}/
        Pop-Location
        Write-Host "  Wire generated."
    }
    "run-backend" {
        Write-Step "Starting backend (go run)..."
        Push-Location $ProjectRoot
        go run ./cmd/${Module}/ -conf configs/
        Pop-Location
    }
    "run-frontend" {
        Write-Step "Starting frontend dev server..."
        Push-Location $FrontendDir
        npm run dev
        Pop-Location
    }
    "dev" {
        Write-Step "Starting backend + frontend dev servers..."
        Push-Location $ProjectRoot
        $logFile = Join-Path $BuildDir "backend.log"
        if (!(Test-Path $BuildDir)) { New-Item -ItemType Directory -Path $BuildDir | Out-Null }
        Start-Process powershell -WindowStyle Hidden -ArgumentList "-Command", "cd '$ProjectRoot'; go run ./cmd/${Module}/ -conf configs/ 2>&1 | Out-File '$logFile'"
        Write-Host "  Backend starting... (log: $logFile)"
        Pop-Location
        Start-Sleep -Seconds 4
        try {
            $r = Invoke-WebRequest "http://localhost:8000/" -UseBasicParsing -TimeoutSec 3 -ErrorAction Stop
            Write-Host "  Backend OK: $($r.StatusCode)"
        } catch { Write-Host "  Backend not ready yet, check logs." }
        Push-Location $FrontendDir
        Write-Host "  Frontend dev server starting..."
        npm run dev
        Pop-Location
    }
    "run" {
        Write-Step "Starting production binary..."
        $binary = Join-Path $BuildDir "${Module}.exe"
        if (!(Test-Path $binary)) {
            Write-Host "  Binary not found, building first..."
            & $MyInvocation.MyCommand.ScriptBlock -Target "backend"
        }
        & $binary -conf (Join-Path $ProjectRoot "configs/")
    }
    "test" {
        Write-Step "Running tests..."
        Push-Location $ProjectRoot
        go test ./... -v -count=1
        Pop-Location
    }
    default {
        Write-Step "Building all..."
        & $MyInvocation.MyCommand.ScriptBlock -Target "frontend"
        & $MyInvocation.MyCommand.ScriptBlock -Target "backend"
        Write-Step "Build complete!"
    }
}
