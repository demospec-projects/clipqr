$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot
$env:GOPATH = Join-Path $PSScriptRoot '.go'
$env:GOCACHE = Join-Path $PSScriptRoot '.cache'
go run github.com/akavel/rsrc -manifest app.manifest -o resource_windows_amd64.syso -arch amd64
if ($LASTEXITCODE -ne 0) { throw 'Génération des ressources impossible.' }
go test ./...
if ($LASTEXITCODE -ne 0) { throw 'Échec des tests.' }
New-Item -ItemType Directory -Force dist | Out-Null
go build -trimpath -ldflags '-H windowsgui -s -w' -o dist/ClipQR.exe .
if ($LASTEXITCODE -ne 0) { throw 'Compilation impossible.' }
Write-Host 'Application créée : dist\ClipQR.exe'
