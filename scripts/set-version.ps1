param (
    [Parameter(Mandatory)]
    [string]$Version
)

$ErrorActionPreference = "Stop"
if ($Version -notmatch '^\d+\.\d+\.\d+$') {
    throw "版本号格式无效: $Version，应为 x.y.z（例如 1.6.0）"
}

$RepoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Definition)
$Utf8NoBom = [System.Text.UTF8Encoding]::new($false)

function Update-RegexMatch {
    param (
        [string]$RelativePath,
        [string]$Pattern,
        [int]$MaxCount = 0
    )
    $FullPath = Join-Path $RepoRoot $RelativePath
    $Content = [IO.File]::ReadAllText($FullPath, $Utf8NoBom)
    $Regex = [regex]$Pattern
    if (-not $Regex.IsMatch($Content)) {
        throw "在 $RelativePath 中未找到匹配模式: $Pattern"
    }
    $Replacement = '${1}' + $Version + '${2}'
    if ($MaxCount -gt 0) {
        $Updated = $Regex.Replace($Content, $Replacement, $MaxCount)
    } else {
        $Updated = $Regex.Replace($Content, $Replacement)
    }
    [IO.File]::WriteAllText($FullPath, $Updated, $Utf8NoBom)
    return $RelativePath
}

$UpdatedFiles = @(
    (Update-RegexMatch -RelativePath "src/backend/update.go" -Pattern '(AppVersion\s*=\s*")[^"]+(")' -MaxCount 1),
    (Update-RegexMatch -RelativePath "src/frontend/package.json" -Pattern '("version"\s*:\s*")[^"]+(")' -MaxCount 1),
    (Update-RegexMatch -RelativePath "wails.json" -Pattern '("productVersion"\s*:\s*")[^"]+(")' -MaxCount 1)
)

Write-Host "已将版本号统一更新为 $Version：" -ForegroundColor Green
foreach ($File in $UpdatedFiles) {
    Write-Host "  - $File"
}
