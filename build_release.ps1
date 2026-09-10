# 果仓助手 (AppleVault) 增量发布构建脚本 (不会清空 build/bin 目录及依赖/数据)
param (
    [switch]$SkipFrontend = $false
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
Set-Location $ScriptDir

# 确保 Go 与 go-winres 环境变量可用（支持便携式 Go）
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    if (Test-Path "$env:TEMP\liveagent-go1.27.1\go\bin\go.exe") {
        $env:GOROOT = "$env:TEMP\liveagent-go1.27.1\go"
        $env:PATH = "$env:GOROOT\bin;$env:USERPROFILE\go\bin;$env:PATH"
    }
}

Write-Host "==> [1/3] 检查前端构建..." -ForegroundColor Cyan
if (-not $SkipFrontend) {
    Set-Location "$ScriptDir/frontend"
    npm run build
    Set-Location $ScriptDir
}

Write-Host "==> [2/4] 确保 build/bin 及 tools 目录结构..." -ForegroundColor Cyan
$BinDir = "$ScriptDir/build/bin"
$ToolsDir = "$BinDir/tools"
if (-not (Test-Path $ToolsDir)) {
    New-Item -ItemType Directory -Force $ToolsDir | Out-Null
}

# 若根目录下有外部依赖，自动补充同步至 tools/ 目录（不覆盖已有）
if ((Test-Path "$ScriptDir/ipatool.exe") -and -not (Test-Path "$ToolsDir/ipatool.exe")) {
    Copy-Item "$ScriptDir/ipatool.exe" "$ToolsDir/" -Force
}
if ((Test-Path "$ScriptDir/ios.exe") -and -not (Test-Path "$ToolsDir/ios.exe")) {
    Copy-Item "$ScriptDir/ios.exe" "$ToolsDir/" -Force
}

Write-Host "==> [3/4] 检查 Windows Logo 图标与应用资源嵌入..." -ForegroundColor Cyan
$SysoFile = "$ScriptDir/rsrc_windows_amd64.syso"
$IconFile = "$ScriptDir/build/appicon.png"
$ShouldRegenerateResources = -not (Test-Path $SysoFile) -or `
    ((Test-Path $IconFile) -and ((Get-Item $IconFile).LastWriteTimeUtc -gt (Get-Item $SysoFile).LastWriteTimeUtc))
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
            "--product-version", "1.1.0",
            "--file-version", "1.1.0",
            "--copyright", "AppleVault (果仓助手)",
            "--arch", "amd64"
        )
        & $WinresCmd @winresArgs
    }
}

Write-Host "==> [4/4] 编译主程序 AppleVault.exe (增量覆盖输出，保留所有依赖与数据)..." -ForegroundColor Cyan
go build -tags "desktop,production" -ldflags "-w -s -H windowsgui" -o "$BinDir/AppleVault.exe" .

Write-Host "`n✅ 发布构建完成！主程序输出于: $BinDir/AppleVault.exe" -ForegroundColor Green
Write-Host "   目录结构："
Write-Host "   ├── AppleVault.exe        (主程序)"
Write-Host "   ├── tools/                (外部依赖目录: ipatool.exe, ios.exe)"
Write-Host "   └── data/                 (程序数据目录: 自动生成 settings.json, tasks.json, 登录凭据)"
