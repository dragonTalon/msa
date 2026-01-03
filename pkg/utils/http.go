package utils

import (
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	restyClient *resty.Client
	restyOnce   sync.Once
)

// GetRestyClient 获取单例的 Resty 客户端
func GetRestyClient() *resty.Client {
	restyOnce.Do(func() {
		restyClient = resty.New().
			SetTimeout(60 * time.Second).
			SetRetryCount(3).
			SetRetryWaitTime(100).
			SetRetryMaxWaitTime(5000)
	})
	return restyClient
}

// ResetRestyClient 重置单例客户端（主要用于测试）
func ResetRestyClient() {
	restyClient = nil
	restyOnce = sync.Once{}
}
