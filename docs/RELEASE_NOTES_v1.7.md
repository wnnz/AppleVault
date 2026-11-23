# 果仓助手 (AppleVault) v1.7.0 更新说明

果仓助手 (AppleVault) v1.7.0 重点围绕 **macOS 凭据加密与跨平台安全加固**、**全链路取消响应与下载队列治理**、**设备直装与前端状态一致性优化**，以及 **自动化测试与双平台 CI/CD 保障** 进行了全面升级。

---

## 🌟 核心更新

### 1. 🔐 跨平台安全加固与原子持久化
- **macOS 钥匙串密文保护**：在 macOS 端将 `settings.json` 中的钥匙串主密码通过系统 Keychain 主密钥结合 AES-256-GCM 加密存储，并显式启用应用信任（`KeychainTrustApplication`）、解锁可用（`KeychainAccessibleWhenUnlocked`）及禁止同步至 iCloud（`KeychainNotSynchronizable`）；同时优化账号中心针对不同系统的提示文案。
- **Windows 原子写盘与系统集成修复**：Windows 下改用 `MoveFileEx`（`MOVEFILE_REPLACE_EXISTING | MOVEFILE_WRITE_THROUGH`）配合 `.bak` 回退机制原子替换配置；修复 `explorer.exe /select,"..."` 在路径包含空格时的定位问题及提取窗口图标句柄的资源释放；macOS/Linux 启动外部文件管理器时正确回收子进程。
- **网络代理安全收敛**：默认关闭初始本地代理开关，并对 `http`、`https`、`socks5`、`socks5h` 协议及主机名进行统一校验；在代理测速与版本更新请求中排空响应体并及时释放空闲连接池。

### 2. ⚡ 全链路取消响应、下载队列控制与会话优化
- **全链路 Context 取消与半成品清理**：为搜索、已购、登录、版本查询、下载、MD5/ZIP 完整性校验及 Sinf 注入打通上下文取消信号；退避重试即时响应取消，出现异常或取消时自动截断清理 `.tmp` 及受损半成品，对路径成分做文件名安全消毒并在指定版本下载时附加版本标识以防文件覆盖。
- **下载队列并发与历史治理**：内置最大 2 路并发下载信号量与排队状态提示；历史记录上限裁剪为 200 条且始终保留活跃任务；采用无碰撞任务 ID、完整 `YYYY-MM-DD HH:MM:SS` 创建时间及大小写不敏感的 Bundle ID 去重；完善任务中断、账号移除与取消状态同步。
- **会话复用、令牌刷新与缓存防抖**：账号移除前先中止对应下载任务与请求；令牌过期时对同账号并发刷新请求去重复用；新增脏标记 Cookie 存储器避免无效落盘；仅对 `.ipa` 安装包执行 `ReplicateSinf` 以完整兼容 macOS `.pkg`；支持多币种商店价格格式化展示；版本元数据缓存统一归并为单一规范键，配合 800ms 防抖异步落盘与退出强制刷盘。

### 3. 📲 设备直装与前端交互体验提升
- **独立安装上下文与防重入取消**：将设备直装从商店请求上下文拆分为独立受控会话，提供并发防重入与 30 分钟超时保护，并在设备直装页面新增「取消安装」操作按钮。
- **iPad 兼容模式安装放行**：支持仅声明支持 iPhone 的 IPA（`UIDeviceFamily: 1`）在 iPad 上以兼容模式通过安装前兼容性体检并给出明确提示。
- **启动日志回放与跨视图一致性**：后端新增最近 100 条日志环形缓冲与回放接口，应用启动后状态栏即可显示早期初始化日志；切换、移除账号及调整下载平台时统一同步与重置搜索、已购及历史版本视图状态，拦截过期异步结果覆盖。

### 4. 🧪 自动化测试与 CI/CD 质量保障
- **前后端单元测覆盖**：大幅扩充后端与 `internal/ipatool` 核心逻辑单元测试（加解密校验、Context 取消传播、并发令牌刷新、缓存防抖与规范化、队列限流与安装隔离等），并在前端引入 Vitest 对版本号解析与比较逻辑建立单元测试。
- **双平台 CI 与发布流水线加固**：新增面向 Pull Request 与分支推送的 Windows + macOS 自动化 CI 工作流（macOS 端启用 `-race` 竞态检测）；发布流水线固定 `go-winres@v0.3.3`、在打包前强制执行前后端测试、为 macOS 构建脚本下载 `go-ios` v1.3.2 增加真实 SHA256 与 Mach-O 架构校验，并统一规范化 `SHA256SUMS.txt` 输出。

---

## 📦 安装与签名说明
- **Windows (`AppleVault-v1.7.0-windows-amd64.zip`)**：解压后直接运行 `AppleVault.exe` 即可，内置 `tools/ios.exe` 设备安装工具，个人数据保存在同目录 `data/` 下。
- **macOS (`AppleVault-v1.7.0-macos-universal.dmg` / `.zip`)**：将 `AppleVault.app` 拖入「应用程序 (`Applications`)」文件夹。由于未使用 Apple Developer ID 公证，首次打开如提示「无法验证开发者」或「已损坏」，在终端执行以下命令后重新打开即可：
  ```bash
  sudo xattr -cr /Applications/AppleVault.app
  ```
- 下载后建议对照 `SHA256SUMS.txt` 校验安装包哈希完整性。
