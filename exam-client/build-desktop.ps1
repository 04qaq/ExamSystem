$ErrorActionPreference = "Stop"
$here = Split-Path -Parent $MyInvocation.MyCommand.Path

Push-Location (Join-Path $here "frontend")
npm run build
Pop-Location

Push-Location (Join-Path $here "desktop")
powershell -ExecutionPolicy Bypass -File ".\scripts\sync-frontend.ps1"
if (-not $env:GOPROXY) { $env:GOPROXY = "https://goproxy.cn,direct" }
$c = Get-Command wails3.exe -ErrorAction SilentlyContinue
$wails3 = if ($c) { $c.Source } else { Join-Path $env:USERPROFILE "go\bin\wails3.exe" }
& $wails3 build
Pop-Location

Write-Host "完成：exe 位于 exam-client\desktop\bin\ExamSystem.exe"
