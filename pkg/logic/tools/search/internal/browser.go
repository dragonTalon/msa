package internal

import (
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

// Browser 浏览器接口
type Browser interface {
	GetPageHTML(ctx context.Context, url string) (string, error)
	GetPageText(ctx context.Context, url string) (string, error)
	Close() error
}

// BrowserManager 浏览器管理器
//
// chromedp 三层上下文模型：
//
//	allocCtx  → 对应一个 Chrome 进程（长期存活，进程级别）
//	browserCtx → 对应一个浏览器实例（长期存活，实例级别）
//	tabCtx    → 对应一个 Tab（每次任务创建，任务完成后关闭）
//
// 浏览器进程和实例只创建一次，每次请求复用，通过独立 Tab 隔离并发任务。
type BrowserManager struct {
	mu            sync.Mutex
	allocCtx      context.Context
	allocCancel   context.CancelFunc
	browserCtx    context.Context
	browserCancel context.CancelFunc
	chromeOK      bool // Chrome 检测结果缓存
}

// NewBrowserManager 创建浏览器管理器（懒加载，不立即启动 Chrome）
func NewBrowserManager() *BrowserManager {
	return &BrowserManager{}
}

// ensureBrowser 确保浏览器进程已启动（线程安全，只创建一次）
// 若浏览器上下文已失效则自动重建
func (b *BrowserManager) ensureBrowser() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 检查现有浏览器是否仍然存活
	if b.browserCtx != nil {
		select {
		case <-b.browserCtx.Done():
			// 浏览器上下文已失效，清理后重建
			log.Warn("[Browser] 浏览器上下文已失效，正在重建...")
			b.destroyLocked()
		default:
			// 浏览器仍然存活，直接返回
			return nil
		}
	}

	// 首次或重建时检测 Chrome（只在需要启动时检测）
	if !b.chromeOK {
		if err := checkChrome(); err != nil {
			return err
		}
		b.chromeOK = true
	}

	// 启动 Chrome 进程（allocCtx 生命周期 = Chrome 进程生命周期）
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-background-networking", true),
		// 关键：禁用 headless 特征，避免被反爬虫检测
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("exclude-switches", "enable-automation"),
		chromedp.Flag("useAutomationExtension", false),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// 创建浏览器实例（browserCtx 生命周期 = 浏览器实例生命周期）
	browserCtx, browserCancel := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(func(format string, args ...interface{}) {
			log.Debugf("[chromedp] "+format, args...)
		}),
	)

	// 预热：触发浏览器真正启动（避免第一次请求时才启动导致超时）
	if err := chromedp.Run(browserCtx); err != nil {
		browserCancel()
		allocCancel()
		return fmt.Errorf("浏览器启动失败: %w", err)
	}

	// 注入反检测脚本：消除 navigator.webdriver 特征
	// 在浏览器级别注入，对所有后续 Tab 生效
	if err := chromedp.Run(browserCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(
				`Object.defineProperty(navigator, 'webdriver', {get: () => undefined});` +
					`window.chrome = { runtime: {} };` +
					`Object.defineProperty(navigator, 'plugins', {get: () => [1,2,3,4,5]});` +
					`Object.defineProperty(navigator, 'languages', {get: () => ['zh-CN','zh','en']});`,
			).Do(ctx)
			return err
		}),
	); err != nil {
		log.Warnf("[Browser] 注入反检测脚本失败（不影响使用）: %v", err)
	}

	b.allocCtx = allocCtx
	b.allocCancel = allocCancel
	b.browserCtx = browserCtx
	b.browserCancel = browserCancel

	log.Info("[Browser] Chrome 浏览器已启动，后续请求将复用此实例")
	return nil
}

// newTab 从浏览器实例派生一个新 Tab 上下文
// 每次任务独占一个 Tab，任务完成后调用返回的 cancel 关闭 Tab
func (b *BrowserManager) newTab(outerCtx context.Context, timeout time.Duration) (context.Context, context.CancelFunc, error) {
	if err := b.ensureBrowser(); err != nil {
		return nil, nil, err
	}

	b.mu.Lock()
	browserCtx := b.browserCtx
	b.mu.Unlock()

	// 从 browserCtx 派生 tabCtx，每个 Tab 独立，互不干扰
	tabCtx, tabCancel := chromedp.NewContext(browserCtx)

	// 叠加超时限制
	tabCtx, timeoutCancel := context.WithTimeout(tabCtx, timeout)

	// 合并 cancel：关闭 Tab + 取消超时
	combinedCancel := func() {
		timeoutCancel()
		tabCancel()
	}

	// 监听外部 ctx，一旦取消则同步关闭 Tab
	go func() {
		select {
		case <-outerCtx.Done():
			log.Warnf("[Browser] 外部 ctx 已取消，关闭 Tab: %v", outerCtx.Err())
			combinedCancel()
		case <-tabCtx.Done():
			// Tab 已自然结束（超时或任务完成），退出监听
		}
	}()

	return tabCtx, combinedCancel, nil
}

// runInTab 在独立 Tab 中执行 chromedp 任务
func (b *BrowserManager) runInTab(outerCtx context.Context, timeout time.Duration, actions ...chromedp.Action) error {
	// 检查外部 ctx 是否已取消
	select {
	case <-outerCtx.Done():
		return fmt.Errorf("请求已取消: %w", outerCtx.Err())
	default:
	}

	tabCtx, tabCancel, err := b.newTab(outerCtx, timeout)
	if err != nil {
		return err
	}
	defer tabCancel() // 任务完成后关闭 Tab，释放资源

	if err := chromedp.Run(tabCtx, actions...); err != nil {
		// 区分：是 Tab 超时/取消，还是浏览器进程本身挂了
		b.mu.Lock()
		browserErr := b.browserCtx.Err()
		b.mu.Unlock()

		if browserErr != nil {
			log.Warnf("[Browser] 浏览器进程已失效，下次请求将自动重建: %v", browserErr)
		}
		return err
	}
	return nil
}

// GetPageHTML 获取页面完整 HTML（复用浏览器，每次使用独立 Tab）
func (b *BrowserManager) GetPageHTML(ctx context.Context, url string) (string, error) {
	var html string
	err := b.runInTab(ctx, 30*time.Second,
		chromedp.Navigate(url),
		// 使用 WaitVisible 比 WaitReady 更稳定，确保 body 真正可见
		chromedp.WaitVisible("body", chromedp.ByQuery),
		// 随机延迟 800ms~1.5s，模拟人类浏览行为，降低被检测概率
		chromedp.Sleep(time.Duration(800+rand.Intn(700))*time.Millisecond),
		// 用 JS 获取 HTML，避免 "No node with given id found" 问题
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`document.documentElement.outerHTML`, &html).Do(ctx)
		}),
	)
	if err != nil {
		return "", fmt.Errorf("获取页面 HTML 失败 [%s]: %w", url, err)
	}
	return html, nil
}

// GetPageText 获取页面文本内容（复用浏览器，每次使用独立 Tab）
func (b *BrowserManager) GetPageText(ctx context.Context, url string) (string, error) {
	var text string
	err := b.runInTab(ctx, 30*time.Second,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(time.Duration(800+rand.Intn(700))*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`document.body.innerText`, &text).Do(ctx)
		}),
	)
	if err != nil {
		return "", fmt.Errorf("获取页面文本失败 [%s]: %w", url, err)
	}
	return text, nil
}

// Close 关闭浏览器进程，释放所有资源
func (b *BrowserManager) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.destroyLocked()
	return nil
}

// destroyLocked 销毁浏览器（调用方必须持有锁）
func (b *BrowserManager) destroyLocked() {
	if b.browserCancel != nil {
		b.browserCancel()
		b.browserCancel = nil
	}
	if b.allocCancel != nil {
		b.allocCancel()
		b.allocCancel = nil
	}
	b.browserCtx = nil
	b.allocCtx = nil
	log.Info("[Browser] 浏览器实例已关闭")
}

// checkChrome 检测 Chrome 是否已安装（包级别函数，结果由调用方缓存）
func checkChrome() error {
	candidates := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", // macOS
		"/usr/bin/google-chrome",    // Linux
		"/usr/bin/chromium-browser", // Linux Chromium
		"/usr/bin/chromium",         // Linux Chromium (部分发行版)
		"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",       // Windows
		"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe", // Windows 32-bit
	}

	for _, path := range candidates {
		cmd := exec.Command(path, "--version")
		output, err := cmd.CombinedOutput()
		if err == nil && len(output) > 0 {
			log.Infof("[Browser] 检测到 Chrome: %s", strings.TrimSpace(string(output)))
			return nil
		}
	}

	return fmt.Errorf("未检测到 Chrome 浏览器，请安装后重试: https://www.google.com/chrome/")
}
