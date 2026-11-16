# AppleVault 开发说明

- 回复用户请使用中文。
- Shell：本机无 WSL/bash，使用 PowerShell。
- 发布构建：在仓库根目录运行 `.\build_release.ps1`，输出到 `build/bin/AppleVault.exe`（包含前端 `npm run build` + Go 编译）。
- Go 不在 PATH 中，可临时使用：`D:\Development\MyDockerComposes\.vendor-deploy\sub-manager-20261107\tools\go\bin\go.exe`（设置 GOROOT 并加入 PATH 后再运行构建脚本）。
- 验证：`go build ./...`、`go test ./src/backend/`、`src/frontend` 下 `npm run build`。
