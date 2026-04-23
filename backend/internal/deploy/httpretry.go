package deploy

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HTTPRetryPolicy 描述一次"带重试的 HTTP 调用"的策略。零值意味着"不重试"。
//
// buildReq 是请求工厂（每次重试都重新构造），文件上传场景必须在工厂里 os.Open 新句柄；
// 不能提前读入 body —— 失败后的 request body 已经被消费。
type HTTPRetryPolicy struct {
	MaxAttempts     int           // 总尝试次数，默认 1
	BaseDelay       time.Duration // 首次退避基准，默认 500ms
	MaxDelay        time.Duration // 单次退避上限，默认 30s
	RetryableStatus []int         // 可重试的 HTTP 状态码；为空时使用默认白名单
}

// defaultRetryableStatus 对幂等写（PUT / DELETE）和 GET 都是可重试的典型集合。
// 408 / 429 / 5xx 意味着"服务端暂时不可用或要求稍后重试"。
var defaultRetryableStatus = []int{
	http.StatusRequestTimeout,     // 408
	http.StatusTooManyRequests,    // 429
	http.StatusInternalServerError,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
}

// DoHTTPWithRetry 执行一次带重试的 HTTP 调用。
//
//   - 每次尝试前调用 buildReq 构造新的 *http.Request（对文件 body 尤其关键）
//   - 仅对"网络错误" 或 "白名单内的 HTTP 状态码"进行重试
//   - 退避：指数 + 随机 jitter，429 响应优先使用 Retry-After 头
//   - 任何时刻 ctx.Done() 返回 ctx.Err() 并不再重试
//   - onRetry 非 nil 时在即将等待重试时调用（供上层打印"第 N 次重试中..."）
//
// 成功（2xx 或不可重试的 4xx）返回 *http.Response；调用方必须负责 Close Body。
func DoHTTPWithRetry(
	ctx context.Context,
	client *http.Client,
	buildReq func() (*http.Request, error),
	policy HTTPRetryPolicy,
	onRetry func(attempt int, waitFor time.Duration, reason string),
) (*http.Response, error) {
	if policy.MaxAttempts < 1 {
		policy.MaxAttempts = 1
	}
	if policy.BaseDelay <= 0 {
		policy.BaseDelay = 500 * time.Millisecond
	}
	if policy.MaxDelay <= 0 {
		policy.MaxDelay = 30 * time.Second
	}
	retryable := policy.RetryableStatus
	if len(retryable) == 0 {
		retryable = defaultRetryableStatus
	}

	var lastErr error
	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		req, err := buildReq()
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)

		if err == nil && !isRetryableStatus(resp.StatusCode, retryable) {
			return resp, nil
		}

		// 失败或非 2xx 可重试：清理当前响应，准备下一轮
		var reason string
		var retryAfter time.Duration
		if err != nil {
			if !isRetryableErr(err) {
				return nil, err
			}
			reason = fmt.Sprintf("网络错误：%v", err)
			lastErr = err
		} else {
			retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
			reason = fmt.Sprintf("HTTP %d", resp.StatusCode)
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			resp.Body.Close()
		}

		if attempt >= policy.MaxAttempts {
			return nil, fmt.Errorf("重试 %d 次仍失败: %w", policy.MaxAttempts, lastErr)
		}

		wait := retryAfter
		if wait <= 0 {
			wait = expoBackoff(attempt, policy)
		}
		if onRetry != nil {
			onRetry(attempt, wait, reason)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
	return nil, fmt.Errorf("重试 %d 次仍失败: %w", policy.MaxAttempts, lastErr)
}

func isRetryableStatus(code int, allow []int) bool {
	for _, c := range allow {
		if c == code {
			return true
		}
	}
	return false
}

// isRetryableErr 粗略判定"值得重试的网络错误"。
// - net.Error.Timeout() / Temporary()
// - connection reset / broken pipe / no such host / EOF 等字符串兜底
func isRetryableErr(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "EOF")
}

// parseRetryAfter 解析 429 响应的 Retry-After 头。
// 支持"秒数"和 HTTP-date 两种格式；无法解析时返回 0（由调用方退化到指数退避）。
func parseRetryAfter(v string) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

// expoBackoff 返回第 attempt 次重试的等待时长：BaseDelay * 2^(attempt-1) + jitter，
// 不超过 MaxDelay。attempt 从 1 起计。
func expoBackoff(attempt int, policy HTTPRetryPolicy) time.Duration {
	base := policy.BaseDelay << (attempt - 1)
	if base > policy.MaxDelay || base < policy.BaseDelay /* 防溢出 */ {
		base = policy.MaxDelay
	}
	jitter := time.Duration(rand.Int64N(int64(base / 4)))
	wait := base + jitter
	if wait > policy.MaxDelay {
		wait = policy.MaxDelay
	}
	return wait
}
