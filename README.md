# IPATool GUI

基于 C# WPF 开发的现代化 Apple App Store 正版应用与历史版本 IPA 下载管理工具。基于 [ipatool](https://github.com/majd/ipatool) 引擎。

## 🌟 功能特性

- **👤 账号管理**：支持 Apple ID 登录、退出登录、状态常驻显示、一键清理本地密钥缓存。
- **🔐 智能双重认证**：登录时若触发 Apple 2FA，自动弹出模态验证窗口，输入 6 位验证码一键提交。
- **🔍 应用搜索**：支持按关键字搜索 App Store，支持筛选平台（iPhone / iPad / Apple TV / VisionOS）。
- **📜 历史版本下载**：
  - 输入 Bundle ID 一键获取所有历史构建版本（Build ID）。
  - 支持按具体版本号（如 `10.2.96`）或构建 ID 实时过滤筛选。
  - 支持单条或批量查询对应历史版本号与发布日期。
  - 选定版本一键发送到下载中心。
- **⬇️ 下载中心**：支持最新版或指定历史构建 ID 下载，支持自动获取免费购买凭证（`--purchase`）。
- **📦 已购应用**：分页查看当前 Apple ID 已购/已获取的应用列表。
- **🌐 网络代理支持**：支持配置 HTTP / SOCKS5 代理，底层自动注入环境变量，支持一键连通性测试。
- **⚙️ 无感密码注入**：自动管理 `--keychain-passphrase`，彻底解决 Windows 终端卡死与交互 Bug。
- **🖥️ 实时日志**：内置控制台实时滚动显示执行命令与输出日志，支持中途取消。

## 🛠️ 编译与运行

### 环境要求
- Windows 10 / 11
- .NET 8.0 SDK 或更高版本

### 快速启动
```bash
# 克隆仓库
git clone https://git.mossdeck.com/wzheng/ipatool-gui.git
cd ipatool-gui

# 编译并运行
dotnet run
```

### 发布独立版本
```bash
dotnet publish -c Release -r win-x64 --no-self-contained -o ./publish
```
