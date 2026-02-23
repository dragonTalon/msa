package internal

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

// Browser 浏览器接口
type Browser interface {
	GetPageHTML(ctx context.Context, url string) (string, error)
	GetPageText(ctx context.Context, url string) (string, error)
	Close() error
}

// BrowserManager 浏览器管理器 - 使用单例模式管理 Chrome 实例
type BrowserManager struct {
	ctx         context.Context
	cancel      context.CancelFunc
	allocCancel context.CancelFunc
	once        sync.Once
	initErr     error
	mu          sync.Mutex
}

// NewBrowserManager 创建浏览器管理器
func NewBrowserManager() *BrowserManager {
	return &BrowserManager{}
}

// GetContext 获取浏览器上下文（单例模式）
func (b *BrowserManager) GetContext() (context.Context, error) {
	b.once.Do(func() {
		// 检测 Chrome 是否安装
		if err := b.checkChrome(); err != nil {
			b.initErr = err
			return
		}

		// 创建浏览器上下文
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			// 禁用 GPU 加速
			chromedp.Flag("disable-gpu", true),
			// 禁用扩展
			chromedp.Flag("disable-extensions", true),
			// 禁用图片加载（提升性能）
			chromedp.Flag("blink-settings", "imagesEnabled=false"),
			// 设置 User-Agent 模拟真实浏览器
			chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
			//// headless 模式
			//chromedp.Flag("headless", true),
			//// 禁用 web 安全
			//chromedp.Flag("disable-web-security", true),
		)

		allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
		b.allocCancel = allocCancel
		b.ctx, b.cancel = chromedp.NewContext(allocCtx)

		// 验证浏览器上下文是否有效
		if b.ctx == nil {
			b.initErr = fmt.Errorf("创建浏览器上下文失败：ctx 为 nil")
			return
		}

		log.Info("浏览器实例已创建")
	})

	// 检查浏览器上下文是否已被取消
	if b.ctx != nil {
		select {
		case <-b.ctx.Done():
			log.Warnf("浏览器上下文已被取消: %v", b.ctx.Err())
			return nil, fmt.Errorf("浏览器上下文已失效: %w", b.ctx.Err())
		default:
		}
	}

	return b.ctx, b.initErr
}

// checkChrome 检测 Chrome 是否安装
func (b *BrowserManager) checkChrome() error {
	// 尝试常见的 Chrome 路径
	paths := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", // macOS
		"/usr/bin/google-chrome",                                           // Linux
		"/usr/bin/chromium-browser",                                        // Linux
		"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",       // Windows
		"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe", // Windows 32-bit
	}

	for _, path := range paths {
		cmd := exec.Command(path, "--version")
		output, err := cmd.CombinedOutput()
		if err == nil && len(output) > 0 {
			log.Infof("检测到 Chrome: %s", strings.TrimSpace(string(output)))
			return nil
		}
	}

	return fmt.Errorf("未检测到 Chrome 浏览器。请安装 Chrome 或 Chromium: https://www.google.com/chrome/")
}

// GetPageHTML 获取页面 HTML
func (b *BrowserManager) GetPageHTML(ctx context.Context, url string) (string, error) {
	_, err := b.GetContext()
	if err != nil {
		return "", err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// 检查传入的 ctx 是否已经取消
	select {
	case <-ctx.Done():
		log.Warnf("传入的 ctx 已取消: %v", ctx.Err())
		return "", ctx.Err()
	default:
	}

	var html string
	// 从浏览器上下文创建带超时的上下文
	// chromedp.Run 需要 chromedp 上下文数据，必须从 b.ctx 派生
	timeoutCtx, cancel := context.WithTimeout(b.ctx, 60*10*time.Second)

	// 同时监听传入的 ctx 取消信号
	go func() {
		select {
		case <-ctx.Done():
			log.Warnf("传入的 ctx 被取消，取消 timeoutCtx: %v", ctx.Err())
			cancel()
		case <-timeoutCtx.Done():
			// timeoutCtx 自己完成了（超时或被取消），退出 goroutine
		}
	}()

	// 先导航到页面并等待基本加载
	err = chromedp.Run(timeoutCtx,
		chromedp.Navigate(url),
		// 等待 body 标签出现（不等待所有资源加载完成）
		chromedp.WaitReady("body", chromedp.ByQuery),
		// 额外等待一小段时间让 JS 执行
		chromedp.Sleep(500*time.Millisecond),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	)

	// 确保 cancel 被调用（如果还没有被调用）
	cancel()

	if err != nil {
		log.Errorf("获取页面 HTML 失败: %v", err)
		return "", fmt.Errorf("获取页面 HTML 失败: %w", err)
	}

	return html, nil
}

// GetPageText 获取页面文本内容
func (b *BrowserManager) GetPageText(ctx context.Context, url string) (string, error) {
	_, err := b.GetContext()
	if err != nil {
		return "", err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// 检查传入的 ctx 是否已经取消
	select {
	case <-ctx.Done():
		log.Warnf("传入的 ctx 已取消: %v", ctx.Err())
		return "", ctx.Err()
	default:
	}

	var text string
	// 从浏览器上下文创建带超时的上下文
	// chromedp.Run 需要 chromedp 上下文数据，必须从 b.ctx 派生
	timeoutCtx, cancel := context.WithTimeout(b.ctx, 15*time.Second)

	// 同时监听传入的 ctx 取消信号
	go func() {
		select {
		case <-ctx.Done():
			log.Warnf("传入的 ctx 被取消，取消 timeoutCtx: %v", ctx.Err())
			cancel()
		case <-timeoutCtx.Done():
		}
	}()

	err = chromedp.Run(timeoutCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text("body", &text, chromedp.ByQuery),
	)

	// 确保 cancel 被调用
	cancel()

	if err != nil {
		log.Errorf("获取页面文本失败: %v", err)
		return "", fmt.Errorf("获取页面文本失败: %w", err)
	}

	return text, nil
}

// Close 关闭浏览器实例
func (b *BrowserManager) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cancel != nil {
		b.cancel()
	}
	if b.allocCancel != nil {
		b.allocCancel()
	}
	if b.cancel != nil || b.allocCancel != nil {
		log.Info("浏览器实例已关闭")
	}
	return nil
}
