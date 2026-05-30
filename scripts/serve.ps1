#!/usr/bin/env pwsh
# Launch forge-agent behind a local OpenAI-compatible endpoint.
#
# Reads OPENAI_API_KEY from your environment (your pwsh profile). Runs in the
# foreground so you see logs; press Ctrl-C to stop.
#
#   ./scripts/serve.ps1                       # :8799, model gpt-5.4-nano
#   ./scripts/serve.ps1 -Addr :9000           # custom port
#   ./scripts/serve.ps1 -BaseUrl https://api.x.ai/v1   # point upstream at xAI
#
# Then, in another terminal:  ./scripts/demo.sh
param(
  [string]$Addr = ':8787',
  [string]$Model = 'gpt-5.4-nano',
  [string]$BaseUrl = ''
)
$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)

if (-not $env:OPENAI_API_KEY) {
  Write-Error 'OPENAI_API_KEY is not set in this environment.'
  exit 1
}

Write-Host "forge-agent -> http://localhost$Addr/v1   (upstream model $Model)"
Write-Host "verify with:  ./scripts/demo.sh   (or: curl http://localhost$Addr/v1/models)"
Write-Host ''

$goArgs = @('run', './cmd/forge-agent', '--addr', $Addr, '--model', $Model)
if ($BaseUrl) { $goArgs += @('--base-url', $BaseUrl) }
& go @goArgs
