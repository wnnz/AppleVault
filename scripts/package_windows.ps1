# 果仓助手 (AppleVault) Windows 干净发布压缩包构建脚本 (绝不包含 build/bin/data)
param (
    [switch]$SkipFrontend = $false
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$RepoRoot = Split-Path -Parent $ScriptDir
Set-Location $RepoRoot

& "$RepoRoot/build_release.ps1" -SkipFrontend:$SkipFrontend
if ($LASTEXITCODE -ne 0) {
    throw "Windows 发布构建失败"
}

$WailsConfig = Get-Content "$RepoRoot/wails.json" -Raw | ConvertFrom-Json
$Version = [string]$WailsConfig.info.productVersion
if (-not $Version) {
    throw "无法从 wails.json 读取版本号"
}

$ExePath = "$RepoRoot/build/bin/AppleVault.exe"
if (-not (Test-Path $ExePath)) {
    throw "未找到构建产物: $ExePath"
}

$IosToolPath = "$RepoRoot/build/bin/tools/ios.exe"
if (-not (Test-Path $IosToolPath)) {
    $IosToolPath = "$RepoRoot/tools/ios.exe"
}
if (-not (Test-Path $IosToolPath)) {
    throw "未找到设备安装工具 ios.exe"
}

$ReleaseDir = "$RepoRoot/build/release"
if (-not (Test-Path $ReleaseDir)) {
    New-Item -ItemType Directory -Force $ReleaseDir | Out-Null
}

$ZipName = "AppleVault-v$Version-windows-amd64.zip"
$ZipPath = "$ReleaseDir/$ZipName"
$ShaPath = "$ZipPath.sha256"

$StageRoot = Join-Path ([IO.Path]::GetTempPath()) ("applevault_winpkg_" + [guid]::NewGuid().ToString("N"))
$AppStage = Join-Path $StageRoot "AppleVault"
try {
    New-Item -ItemType Directory -Force (Join-Path $AppStage "tools") | Out-Null
    Copy-Item $ExePath (Join-Path $AppStage "AppleVault.exe") -Force
    Copy-Item $IosToolPath (Join-Path $AppStage "tools\ios.exe") -Force

    if (Test-Path "$RepoRoot/README.md") {
        Copy-Item "$RepoRoot/README.md" (Join-Path $AppStage "README.md") -Force
    }
    if (Test-Path "$RepoRoot/LICENSE") {
        Copy-Item "$RepoRoot/LICENSE" (Join-Path $AppStage "LICENSE") -Force
    }
    if (Test-Path "$RepoRoot/third_party/ipatool-LICENSE.txt") {
        New-Item -ItemType Directory -Force (Join-Path $AppStage "third_party") | Out-Null
        Copy-Item "$RepoRoot/third_party/ipatool-LICENSE.txt" (Join-Path $AppStage "third_party\ipatool-LICENSE.txt") -Force
    }

    if (Test-Path $ZipPath) {
        Remove-Item $ZipPath -Force
    }
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    [System.IO.Compression.ZipFile]::CreateFromDirectory(
        $AppStage,
        $ZipPath,
        [System.IO.Compression.CompressionLevel]::Optimal,
        $true
    )

    $Hash = (Get-FileHash -Path $ZipPath -Algorithm SHA256).Hash.ToLowerInvariant()
    [IO.File]::WriteAllText($ShaPath, "$Hash  $ZipName`r`n", [System.Text.UTF8Encoding]::new($false))
} finally {
    Remove-Item -Recurse -Force $StageRoot -ErrorAction SilentlyContinue
}

Write-Host "`nWindows 发布压缩包已生成: $ZipPath" -ForegroundColor Green
Write-Host "SHA256 ($Hash) 已写入: $ShaPath" -ForegroundColor DarkGray
