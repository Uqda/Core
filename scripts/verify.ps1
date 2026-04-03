#Requires -Version 5.1
<#
.SYNOPSIS
  Full local verification (parity with GitHub Actions + recommended extras).
  Run from repo root:  .\scripts\verify.ps1
  Or:                   pwsh -File scripts/verify.ps1 [-Race] [-SkipVuln] [-SkipLint]
#>
param(
    [switch]$Race,
    [switch]$SkipVuln,
    [switch]$SkipLint
)

$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

function Fail($msg) {
    Write-Host $msg -ForegroundColor Red
    exit 1
}

Write-Host "Repository: $Root" -ForegroundColor Green
Write-Host ""

Write-Host "=== go version ===" -ForegroundColor Cyan
go version
if ($LASTEXITCODE -ne 0) { Fail "go not found; install Go 1.25.8+ (see go.mod toolchain)." }

Write-Host "`n=== go mod verify ===" -ForegroundColor Cyan
go mod verify
if ($LASTEXITCODE -ne 0) { Fail "go mod verify failed" }

if (-not $SkipLint) {
    $lint = Get-Command golangci-lint -ErrorAction SilentlyContinue
    if ($lint) {
        Write-Host "`n=== golangci-lint run ===" -ForegroundColor Cyan
        golangci-lint run --timeout=5m
        if ($LASTEXITCODE -ne 0) { Fail "golangci-lint failed" }
    } else {
        Write-Host "`n=== golangci-lint (skipped: not in PATH) ===" -ForegroundColor Yellow
    }
}

Write-Host "`n=== go vet ./... ===" -ForegroundColor Cyan
go vet ./...
if ($LASTEXITCODE -ne 0) { Fail "go vet failed" }

if (-not $SkipVuln) {
    Write-Host "`n=== govulncheck ./... ===" -ForegroundColor Cyan
    go install golang.org/x/vuln/cmd/govulncheck@latest
    if ($LASTEXITCODE -ne 0) { Fail "go install govulncheck failed" }
    govulncheck ./...
    if ($LASTEXITCODE -ne 0) { Fail "govulncheck failed" }
}

Write-Host "`n=== go build ./... ===" -ForegroundColor Cyan
go build -v ./...
if ($LASTEXITCODE -ne 0) { Fail "go build ./... failed" }

Write-Host "`n=== go build cmd (explicit) ===" -ForegroundColor Cyan
go build -v ./cmd/uqda
if ($LASTEXITCODE -ne 0) { Fail "go build uqda failed" }
go build -v ./cmd/uqdactl
if ($LASTEXITCODE -ne 0) { Fail "go build uqdactl failed" }

if ($Race) {
    Write-Host "`n=== go test -race ./... (slower) ===" -ForegroundColor Cyan
    go test ./... -count=1 -race -timeout=120s
} else {
    Write-Host "`n=== go test ./... ===" -ForegroundColor Cyan
    go test ./... -count=1 -timeout=120s
}
if ($LASTEXITCODE -ne 0) { Fail "go test failed" }

Write-Host "`nAll checks passed." -ForegroundColor Green
