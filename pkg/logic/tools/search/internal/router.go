package internal

import (
	"context"
	"strings"
	"time"

	searchmsa_model "msa/pkg/model"
)

// ErrorType 错误类型
type ErrorType int

const (
	ErrorTypePermanent ErrorType = iota // CAPTCHA, 403, 404
	ErrorTypeTemporary                  // timeout, 5xx
	ErrorTypeNetwork                    // connection refused
)

// SearchResult 搜索结果（包含状态）
type SearchResult struct {
	Results    []searchmsa_model.SearchResultItem
	Status     string // "success", "failed", "captcha"
	Query      string
	Error      string
	Message    string
	RequestID  string
	UsedEngine string
	Duration   time.Duration
}

// SearchConfig 搜索配置
type SearchConfig struct {
	Engines         []string
	Timeouts        map[string]time.Duration
	FailoverEnabled bool
	FailoverDelay   time.Duration
	MaxRetries      int
}

// DefaultConfig 返回默认配置
func DefaultConfig() *SearchConfig {
	return &SearchConfig{
		Engines:         []string{"google", "bing"},
		Timeouts:        map[string]time.Duration{"google": 60 * time.Second, "bing": 30 * time.Second},
		FailoverEnabled: true,
		FailoverDelay:   2 * time.Second,
		MaxRetries:      2,
	}
}

// SearchRouter 搜索引擎路由器
type SearchRouter struct {
	engines []SearchEngine
	tracker *EngineTracker
	logger  *SearchLogger
	config  *SearchConfig
}

// NewSearchRouter 创建搜索路由器
func NewSearchRouter(browser Browser, logger *SearchLogger) *SearchRouter {
	config := DefaultConfig()

	return &SearchRouter{
		engines: []SearchEngine{
			NewGoogleSearchEngine(browser),
			NewBingSearchEngine(browser),
		},
		tracker: NewEngineTracker(),
		logger:  logger,
		config:  config,
	}
}

// Search 执行搜索，自动降级
func (r *SearchRouter) Search(ctx context.Context, query string, numResults int) *SearchResult {
	requestID := r.logger.GenerateRequestID()
	startTime := time.Now()

	result := &SearchResult{
		RequestID: requestID,
		Query:     query,
	}

	var failedEngines = make(map[string]string)

	// 遍历所有引擎
	for i, engine := range r.engines {
		engineName := r.getEngineName(engine)

		// 检查引擎是否可用
		if !r.tracker.ShouldTry(engineName) {
			r.logger.WithField("request_id", requestID).WithField("engine", engineName).
				WithField("reason", "cooldown").Warn("Skipping engine (in cooldown)")
			failedEngines[engineName] = "cooldown"
			continue
		}

		// 记录开始
		r.logger.LogStart(requestID, query, engineName)

		// 如果不是第一个引擎，等待一段时间
		if i > 0 && r.config.FailoverDelay > 0 {
			time.Sleep(r.config.FailoverDelay)
		}

		// 执行搜索
		searchStart := time.Now()
		results, err := r.searchWithRetry(ctx, engine, query, numResults, requestID)
		duration := time.Since(searchStart)

		if err == nil {
			// 成功
			r.tracker.RecordSuccess(engineName)
			r.logger.LogSuccess(requestID, engineName, duration, len(results))

			result.Results = results
			result.Status = "success"
			result.UsedEngine = engineName
			result.Duration = time.Since(startTime)
			return result
		}

		// 失败
		r.tracker.RecordFailure(engineName, err)

		// 记录失败
		errorType := ClassifyError(err)
		errorMsg := err.Error()
		failedEngines[engineName] = errorMsg

		if errorType == ErrorTypePermanent {
			if IsCaptchaError(err) {
				r.logger.LogCaptcha(requestID, engineName, []string{errorMsg})
			} else {
				r.logger.LogEngineError(requestID, engineName, "permanent", errorMsg)
			}
		} else {
			r.logger.LogEngineError(requestID, engineName, "temporary", errorMsg)
		}

		// 记录降级
		if i < len(r.engines)-1 {
			nextEngineName := r.getEngineName(r.engines[i+1])
			r.logger.LogFailover(requestID, engineName, nextEngineName, errorMsg)
		}
	}

	// 所有引擎都失败
	r.logger.LogAllEnginesFailed(requestID, failedEngines)

	result.Status = "failed"
	result.Error = "all_engines_failed"
	result.Message = "搜索暂时不可用，请稍后再试"
	result.Duration = time.Since(startTime)

	// 检查是否有 CAPTCHA
	for _, errMsg := range failedEngines {
		if strings.Contains(strings.ToLower(errMsg), "captcha") {
			result.Status = "captcha"
			result.Error = "detected_automation"
			result.Message = "搜索暂时不可用，可能是由于访问频率限制或检测到自动化行为。请稍后再试，或提供更多信息让我基于已有知识回答。"
			break
		}
	}

	return result
}

// searchWithRetry 带重试的搜索
func (r *SearchRouter) searchWithRetry(ctx context.Context, engine SearchEngine, query string, numResults int, requestID string) ([]searchmsa_model.SearchResultItem, error) {
	engineName := r.getEngineName(engine)
	maxRetries := r.config.MaxRetries

	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		// 检查 context 是否已取消
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		results, err := engine.Search(ctx, query, numResults)
		if err == nil {
			return results, nil
		}

		lastErr = err

		// 永久错误不重试
		if ClassifyError(err) == ErrorTypePermanent {
			return nil, err
		}

		// 记录重试
		if i < maxRetries {
			r.logger.LogEngineRetry(requestID, engineName, i+1, maxRetries)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(1 * time.Second):
			}
		}
	}

	return nil, lastErr
}

// getEngineName 获取引擎名称
func (r *SearchRouter) getEngineName(engine SearchEngine) string {
	// 尝试类型断言
	if bing, ok := engine.(*BingSearchEngine); ok {
		return bing.Name()
	}
	if _, ok := engine.(*GoogleSearchEngine); ok {
		return "google" // Google 没有 Name() 方法，直接返回
	}
	return "unknown"
}

// GetTracker 获取追踪器（用于测试）
func (r *SearchRouter) GetTracker() *EngineTracker {
	return r.tracker
}

// SetLogger 设置日志记录器
func (r *SearchRouter) SetLogger(logger *SearchLogger) {
	r.logger = logger
}

// SetConfig 设置配置
func (r *SearchRouter) SetConfig(config *SearchConfig) {
	r.config = config
}

// ClassifyError 分类错误
func ClassifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeTemporary
	}

	errStr := strings.ToLower(err.Error())

	// 永久错误
	permanentErrors := []string{"captcha", "403", "404", "forbidden"}
	for _, pe := range permanentErrors {
		if strings.Contains(errStr, pe) {
			return ErrorTypePermanent
		}
	}

	// 临时错误
	temporaryErrors := []string{"timeout", "5", "deadline exceeded"}
	for _, te := range temporaryErrors {
		if strings.Contains(errStr, te) {
			return ErrorTypeTemporary
		}
	}

	return ErrorTypeNetwork
}

// IsCaptchaError 判断是否为 CAPTCHA 错误
func IsCaptchaError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	captchaIndicators := []string{"captcha", "unusual traffic"}
	for _, indicator := range captchaIndicators {
		if strings.Contains(errStr, indicator) {
			return true
		}
	}
	return false
}
