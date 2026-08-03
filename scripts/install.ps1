# p2t CLI Installer for Windows (PowerShell)
$ErrorActionPreference = "Stop"

$InstallDir = "$env:USERPROFILE\.p2t\bin"
$BinaryName = "p2t.exe"
$TargetPath = Join-Path $InstallDir $BinaryName
$SourcePath = "dist\p2t-windows-amd64.exe"

Write-Host "🍬 Instalando p2t CLI para Windows..." -ForegroundColor Cyan

If (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

If (Test-Path $SourcePath) {
    Copy-Item -Path $SourcePath -Destination $TargetPath -Force
    Write-Host " └─ Binário copiado de $SourcePath para: $TargetPath" -ForegroundColor Green
} ElseIf (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Host " └─ Executável $SourcePath não encontrado. Compilando via 'go build'..." -ForegroundColor Yellow
    go build -o $TargetPath ./cmd/p2t
    Write-Host " └─ Binário compilado e instalado em: $TargetPath" -ForegroundColor Green
} Else {
    Write-Host " └─ Executável $SourcePath não encontrado e Go não está instalado." -ForegroundColor Red
    Write-Host "   Execute 'scripts/build.sh' ou instale o Go." -ForegroundColor Yellow
    Exit 1
}

$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
If ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$InstallDir", "User")
    Write-Host " └─ Adicionado $InstallDir ao PATH de Usuário." -ForegroundColor Green
}

If ($env:Path -notlike "*$InstallDir*") {
    $env:Path = "$env:Path;$InstallDir"
}

Write-Host "✅ Instalação concluída com sucesso!" -ForegroundColor Green
Write-Host " Use 'p2t --help' para começar." -ForegroundColor Cyan
