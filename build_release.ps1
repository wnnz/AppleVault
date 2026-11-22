# 果仓助手 (AppleVault) 增量发布构建脚本 (不会清空 build/bin 目录及依赖/数据)
param (
    [switch]$SkipFrontend = $false
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$SourceDir = "$ScriptDir/src"
$FrontendDir = "$SourceDir/frontend"
$ExternalToolsDir = "$ScriptDir/tools"
Set-Location $ScriptDir

# 确保 Go 与 go-winres 环境变量可用（支持便携式 Go）
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    $GoCandidates = @(
        "D:\Development\MyDockerComposes\.vendor-deploy\sub-manager-20261107\tools\go",
        "$env:TEMP\liveagent-go1.27.1\go"
    )
    foreach ($candidate in $GoCandidates) {
        if (Test-Path "$candidate\bin\go.exe") {
            $env:GOROOT = $candidate
            $env:PATH = "$env:GOROOT\bin;$env:USERPROFILE\go\bin;$env:PATH"
            break
        }
    }
}

Write-Host "==> [0/4] 检查版本一致性..." -ForegroundColor Cyan
$UpdateGoContent = Get-Content "$SourceDir/backend/update.go" -Raw
if ($UpdateGoContent -notmatch 'AppVersion\s*=\s*"([^"]+)"') {
    throw "无法从 src/backend/update.go 解析 AppVersion"
}
$UpdateGoVersion = $Matches[1]
$PkgVersion = (Get-Content "$FrontendDir/package.json" -Raw | ConvertFrom-Json).version
$WailsVersion = (Get-Content "$ScriptDir/wails.json" -Raw | ConvertFrom-Json).info.productVersion
if (($UpdateGoVersion -ne $PkgVersion) -or ($UpdateGoVersion -ne $WailsVersion)) {
    throw "版本号不一致: update.go=$UpdateGoVersion, package.json=$PkgVersion, wails.json=$WailsVersion，请运行 scripts/set-version.ps1 -Version x.y.z 统一"
}
$AppVersion = $UpdateGoVersion
Write-Host "    当前版本: $AppVersion" -ForegroundColor DarkGray

Write-Host "==> [1/4] 检查前端构建..." -ForegroundColor Cyan
if (-not $SkipFrontend) {
    Set-Location $FrontendDir
    npm run build
    if ($LASTEXITCODE -ne 0) {
        throw "前端构建失败"
    }
    Set-Location $ScriptDir
}

Write-Host "==> [2/4] 确保 build/bin 及 tools 目录结构..." -ForegroundColor Cyan
$BinDir = "$ScriptDir/build/bin"
$ToolsDir = "$BinDir/tools"
if (-not (Test-Path $ToolsDir)) {
    New-Item -ItemType Directory -Force $ToolsDir | Out-Null
}

# App Store 能力已编译进主程序，仅需同步设备管理工具。
Get-ChildItem -LiteralPath $ToolsDir -Filter "ipatool*.exe" -File | Remove-Item -Force
if (Test-Path "$ExternalToolsDir/ios.exe") {
    $TargetIos = "$ToolsDir/ios.exe"
    $ShouldCopyIos = -not (Test-Path $TargetIos) -or `
        ((Get-FileHash "$ExternalToolsDir/ios.exe" -Algorithm SHA256).Hash -ne (Get-FileHash $TargetIos -Algorithm SHA256).Hash)
    if ($ShouldCopyIos) {
        Copy-Item "$ExternalToolsDir/ios.exe" $TargetIos -Force
    }
}

Write-Host "==> [3/4] 检查 Windows Logo 图标与应用资源嵌入..." -ForegroundColor Cyan
$SysoFile = "$SourceDir/rsrc_windows_amd64.syso"
$IconFile = "$ScriptDir/build/appicon.png"
$WinresStamp = "$BinDir/.winres-version"
$CurrentWinresVersion = if (Test-Path $WinresStamp) { (Get-Content $WinresStamp -Raw).Trim() } else { "" }
$ShouldRegenerateResources = -not (Test-Path $SysoFile) -or `
    ($CurrentWinresVersion -ne $AppVersion) -or `
    ((Test-Path $IconFile) -and ((Get-Item $IconFile).LastWriteTimeUtc -gt (Get-Item $SysoFile).LastWriteTimeUtc)) -or `
    ((Get-Item $MyInvocation.MyCommand.Definition).LastWriteTimeUtc -gt (Get-Item $SysoFile).LastWriteTimeUtc)
if ($ShouldRegenerateResources) {
    $WinresCmd = Get-Command go-winres -ErrorAction SilentlyContinue
    if (-not $WinresCmd -and (Test-Path "$env:USERPROFILE/go/bin/go-winres.exe")) {
        $WinresCmd = "$env:USERPROFILE/go/bin/go-winres.exe"
    }
    if ($WinresCmd) {
        $winresArgs = @(
            "simply",
            "--icon", $IconFile,
            "--manifest", "gui",
            "--product-name", "果仓助手",
            "--file-description", "果仓助手 (AppleVault)",
            "--product-version", $AppVersion,
            "--file-version", $AppVersion,
            "--copyright", "AppleVault (果仓助手)",
            "--arch", "amd64"
        )
        Push-Location $SourceDir
        try {
            & $WinresCmd @winresArgs
            if ($LASTEXITCODE -ne 0) {
                throw "Windows 应用资源生成失败"
            }
        } finally {
            Pop-Location
        }
        if (-not (Test-Path $BinDir)) {
            New-Item -ItemType Directory -Force $BinDir | Out-Null
        }
        [IO.File]::WriteAllText($WinresStamp, $AppVersion, [System.Text.UTF8Encoding]::new($false))
    } else {
        throw "需要 go-winres 生成 Windows 应用资源"
    }
}

Write-Host "==> [4/4] 编译主程序 AppleVault.exe (增量覆盖输出，保留所有依赖与数据)..." -ForegroundColor Cyan
go build -tags "desktop,production" -ldflags "-w -s -H windowsgui" -o "$BinDir/AppleVault.exe" ./src
if ($LASTEXITCODE -ne 0) {
    throw "Go 主程序构建失败"
}

Write-Host "`n✅ 发布构建完成！主程序输出于: $BinDir/AppleVault.exe" -ForegroundColor Green
Write-Host "   目录结构："
Write-Host "   ├── AppleVault.exe        (主程序)"
Write-Host "   ├── tools/                (外部依赖目录: ios.exe)"
Write-Host "   └── data/                 (程序数据目录: 自动生成 settings.json, tasks.json, 登录凭据)"
