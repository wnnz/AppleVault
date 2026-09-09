# 果仓助手 (AppleVault) 增量发布构建脚本 (不会清空 build/bin 目录及依赖/数据)
param (
    [switch]$SkipFrontend = $false
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
Set-Location $ScriptDir

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
if (-not (Test-Path $SysoFile)) {
    $RsrcCmd = Get-Command rsrc -ErrorAction SilentlyContinue
    if (-not $RsrcCmd -and (Test-Path "$env:USERPROFILE/go/bin/rsrc.exe")) {
        $RsrcCmd = "$env:USERPROFILE/go/bin/rsrc.exe"
    }
    $IcoFile = "$ScriptDir/build/windows/icon.ico"
    $ManifestFile = "$ScriptDir/build/windows/wails.exe.manifest"
    if ($RsrcCmd -and (Test-Path $IcoFile) -and (Test-Path $ManifestFile)) {
        & $RsrcCmd -manifest "$ManifestFile" -ico "$IcoFile" -arch "amd64" -o "$SysoFile"
    }
}

Write-Host "==> [4/4] 编译主程序 AppleVault.exe (增量覆盖输出，保留所有依赖与数据)..." -ForegroundColor Cyan
go build -tags "desktop,production" -ldflags "-w -s -H windowsgui" -o "$BinDir/AppleVault.exe" .

Write-Host "`n✅ 发布构建完成！主程序输出于: $BinDir/AppleVault.exe" -ForegroundColor Green
Write-Host "   目录结构："
Write-Host "   ├── AppleVault.exe        (主程序)"
Write-Host "   ├── tools/                (外部依赖目录: ipatool.exe, ios.exe)"
Write-Host "   └── data/                 (程序数据目录: 自动生成 settings.json, tasks.json, 登录凭据)"
