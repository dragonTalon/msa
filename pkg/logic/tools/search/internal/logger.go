package internal

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// SearchLogger 结构化搜索日志
type SearchLogger struct {
	*logrus.Logger
}

// NewSearchLogger 创建搜索日志记录器
func NewSearchLogger(baseLogger *logrus.Logger) *SearchLogger {
	return &SearchLogger{
		Logger: baseLogger,
	}
}

// GenerateRequestID 生成唯一的请求 ID
// 格式：SRCH-{timestamp}-{random}
func (l *SearchLogger) GenerateRequestID() string {
	timestamp := time.Now().UnixMilli()
	random := rand.Intn(10000)
	return fmt.Sprintf("SRCH-%d-%04d", timestamp, random)
}

// LogStart 记录搜索开始
func (l *SearchLogger) LogStart(requestID, query, engine string) {
	l.WithFields(logrus.Fields{
		"request_id": requestID,
		"query":      query,
		"engine":     engine,
		"event":      "search_start",
		"timestamp":  time.Now().UnixMilli(),
	}).Info("Starting search")
}

// LogSuccess 记录搜索成功
func (l *SearchLogger) LogSuccess(requestID, engine string, duration time.Duration, resultCount int) {
	l.WithFields(logrus.Fields{
		"request_id":   requestID,
		"engine":       engine,
		"event":        "search_success",
		"duration_ms":  duration.Milliseconds(),
		"result_count": resultCount,
		"timestamp":    time.Now().UnixMilli(),
	}).Info("Search completed successfully")
}

// LogFailover 记录引擎降级
func (l *SearchLogger) LogFailover(requestID, from, to, reason string) {
	l.WithFields(logrus.Fields{
		"request_id":  requestID,
		"from_engine": from,
		"to_engine":   to,
		"reason":      reason,
		"event":       "failover",
		"timestamp":   time.Now().UnixMilli(),
	}).Warn("Engine failover")
}

// LogCaptcha 记录 CAPTCHA 检测
func (l *SearchLogger) LogCaptcha(requestID, engine string, indicators []string) {
	l.WithFields(logrus.Fields{
		"request_id": requestID,
		"engine":     engine,
		"error_type": "captcha",
		"indicators": indicators,
		"event":      "captcha_detected",
		"timestamp":  time.Now().UnixMilli(),
	}).Warn("CAPTCHA detected")
}

// LogAllEnginesFailed 记录所有引擎失败
func (l *SearchLogger) LogAllEnginesFailed(requestID string, failedEngines map[string]string) {
	l.WithFields(logrus.Fields{
		"request_id":     requestID,
		"failed_engines": failedEngines,
		"total_attempts": len(failedEngines),
		"event":          "all_engines_failed",
		"status":         "failed",
		"timestamp":      time.Now().UnixMilli(),
	}).Error("All search engines failed")
}

// LogEngineError 记录引擎错误
func (l *SearchLogger) LogEngineError(requestID, engine, errorType, errorMessage string) {
	l.WithFields(logrus.Fields{
		"request_id":    requestID,
		"engine":        engine,
		"error_type":    errorType,
		"error_message": errorMessage,
		"event":         "engine_error",
		"timestamp":     time.Now().UnixMilli(),
	}).Error("Search engine error")
}

// LogEngineRetry 记录引擎重试
func (l *SearchLogger) LogEngineRetry(requestID, engine string, retryCount int, maxRetries int) {
	l.WithFields(logrus.Fields{
		"request_id":  requestID,
		"engine":      engine,
		"retry_count": retryCount,
		"max_retries": maxRetries,
		"event":       "engine_retry",
		"timestamp":   time.Now().UnixMilli(),
	}).Info("Retrying search engine")
}
