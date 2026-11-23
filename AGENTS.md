# AppleVault 开发说明

- 回复用户请使用中文。
- Shell：本机无 WSL/bash，使用 PowerShell。
- 发布构建：在仓库根目录运行 `.\build_release.ps1`，输出到 `build/bin/AppleVault.exe`（包含前端 `npm run build` + Go 编译）。
- Go 工具链：本机 PATH 中已提供 `go`；若在未配置 PATH 的环境中，也可临时使用 `D:\Development\MyDockerComposes\.vendor-deploy\sub-manager-20261107\tools\go\bin\go.exe`（设置 `GOROOT` 并加入 PATH 后再运行构建脚本）。
- 验证：
  - `go vet -unsafeptr=false ./...`（Windows 下过滤 `internal/ipatool/sap/unicorn/library_prepare_windows*.go` 既有 `unsafeptr` 告警）
  - `go build ./...`
  - `go test -count=1 ./src/backend/...`
  - `src/frontend` 下运行 `pnpm run test` 与 `npm run build`
