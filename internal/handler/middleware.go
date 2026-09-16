package handler

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"accesscontrol/pkg/httpx"
)

// authMiddleware 校验 X-API-Key 请求头；静态资源与导出接口同样受保护，
// 仅放行健康检查与静态文件（页面本身不敏感，但 API 需鉴权）。
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			key := r.Header.Get("X-API-Key")
			if key != s.apiKey() {
				httpx.Unauthorized(w, "无效的 API Key")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimiter 基于 IP 的令牌桶限流器。
type rateLimiter struct {
	mu       sync.Mutex
	rate     int
	capacity int
	buckets  map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

func newRateLimiter(rate int) *rateLimiter {
	return &rateLimiter{
		rate:     rate,
		capacity: rate,
		buckets:  make(map[string]*bucket),
	}
}

// allow 判断给定 key 是否允许通过，并消耗一个令牌。
func (r *rateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	b, ok := r.buckets[key]
	if !ok {
		b = &bucket{tokens: float64(r.capacity), last: now}
		r.buckets[key] = b
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * float64(r.rate)
	if b.tokens > float64(r.capacity) {
		b.tokens = float64(r.capacity)
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// clientIP 提取客户端 IP（优先 X-Forwarded-For）。
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

// rateLimitMiddleware 对 /api/ 路径按客户端 IP 限流。
func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	limiter := newRateLimiter(s.rateLimit())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") && !limiter.allow(clientIP(r)) {
			httpx.Error(w, http.StatusTooManyRequests, 429, "请求过于频繁，请稍后再试")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) rateLimit() int {
	if s.cfg != nil && s.cfg.RateLimit > 0 {
		return s.cfg.RateLimit
	}
	return 100
}

// healthz 提供健康检查端点（无需鉴权）。
func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]string{"status": "ok", "time": time.Now().Format(time.RFC3339)})
}
