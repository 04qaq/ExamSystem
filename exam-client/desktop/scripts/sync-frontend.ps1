$ErrorActionPreference = "Stop"
$desktopDir = Split-Path -Parent $PSScriptRoot
$examClientDir = Split-Path -Parent $desktopDir
$src = Join-Path $examClientDir "frontend\dist"
$dest = Join-Path $desktopDir "embed\dist"

if (-not (Test-Path $src)) {
    Write-Error "找不到 $src ，请先在 exam-client\frontend 执行 npm run build"
}

if (Test-Path $dest) {
    Remove-Item $dest -Recurse -Force
}
New-Item -ItemType Directory -Path $dest -Force | Out-Null
Copy-Item -Path (Join-Path $src "*") -Destination $dest -Recurse -Force
Write-Host "已同步前端产物 -> $dest"
