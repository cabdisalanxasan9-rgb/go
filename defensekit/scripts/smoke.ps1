param(
    [string]$Target = "scanme.nmap.org",
    [int]$WebPort = 8081,
    [string]$Password = "StrongP@ssw0rd123!"
)

$ErrorActionPreference = "Stop"

function Write-Step {
    param([string]$Message)
    Write-Host "[SMOKE] $Message" -ForegroundColor Cyan
}

$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $projectRoot

Write-Step "Running password mode test"
go run ./cmd -mode password -password $Password -output password-smoke.json -format json | Out-Host

Write-Step "Starting web server on port $WebPort"
$job = Start-Job -ScriptBlock {
    param($root, $port)
    Set-Location $root
    go run ./cmd -serve -addr (":" + $port)
} -ArgumentList $projectRoot, $WebPort

try {
    Start-Sleep -Seconds 3

    Write-Step "Checking /api/health"
    $health = Invoke-WebRequest -UseBasicParsing ("http://127.0.0.1:{0}/api/health" -f $WebPort)
    $health.Content | Out-Host

    Write-Step "Checking /api/http"
    $http = Invoke-WebRequest -UseBasicParsing ("http://127.0.0.1:{0}/api/http?target={1}" -f $WebPort, $Target)
    $http.Content | Out-Host
}
finally {
    Write-Step "Stopping web server job"
    Stop-Job $job -ErrorAction SilentlyContinue
    Remove-Job $job -ErrorAction SilentlyContinue
}

Write-Step "Smoke tests completed"
