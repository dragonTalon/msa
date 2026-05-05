# local-chrome-detection 规格说明

## Purpose

BrowserManager 在初始化时自动检测并优先连接用户本地 Chrome 浏览器，从根源规避 Google CAPTCHA。

## Requirements

### Requirement: detect-local-chrome

BrowserManager SHALL 在 `ensureBrowser()` 中优先检测本地 Chrome 调试端口（127.0.0.1:9222），若可用则通过 CDP 连接，利用用户真实浏览器 profile 执行搜索。

#### Scenario: 本地 Chrome 已开启远程调试

- **GIVEN** 用户 Chrome 以 `--remote-debugging-port=9222` 启动
- **WHEN** BrowserManager 执行 `ensureBrowser()`
- **THEN** 通过 `chromedp.NewRemoteAllocator` 连接到 `ws://127.0.0.1:9222`
- **AND** 后续搜索复用该连接的浏览器实例
- **AND** 记录日志 "已连接到本地 Chrome 浏览器"

#### Scenario: Chrome 进程未运行

- **GIVEN** Chrome 进程不存在且 :9222 不可达
- **WHEN** BrowserManager 执行 `ensureBrowser()`
- **THEN** 启动 Chrome + 用户默认 profile + `--remote-debugging-port=9222`
- **AND** 然后通过 CDP 连接

#### Scenario: Chrome 运行中但未开启远程调试

- **GIVEN** Chrome 进程在运行但未以 `--remote-debugging-port` 启动
- **WHEN** BrowserManager 执行 `ensureBrowser()`
- **THEN** 降级到现有 headless chromedp 模式
- **AND** 记录日志 "Chrome 运行中但调试端口不可用，使用 headless 模式"

#### Scenario: 所有本地 Chrome 方式均不可用

- **GIVEN** :9222 不可达、无法启动 Chrome、无法连接
- **WHEN** BrowserManager 执行 `ensureBrowser()`
- **THEN** 最终降级到现有 headless chromedp 模式
- **AND** 搜索功能不会因本地 Chrome 不可用而完全失效

---

### Requirement: security-constraints

CDP 连接 SHALL 仅访问本地地址，不连接远程主机。

#### Scenario: 仅连接 localhost

- **GIVEN** 检测到 :9222 端口开放
- **WHEN** 建立 CDP 连接
- **THEN** 仅连接到 `127.0.0.1:9222`
- **AND** 不尝试连接任何远程地址

#### Scenario: 连接前验证

- **GIVEN** 收到 HTTP 响应从 `http://127.0.0.1:9222/json/version`
- **WHEN** 解析响应
- **THEN** 验证 `Browser` 字段包含 "Chrome"
- **AND** 验证通过后才建立 WebSocket 连接
