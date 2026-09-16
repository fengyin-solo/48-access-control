// Package handler 实现 HTTP 处理器层。
package handler

import (
	"errors"
	"net/http"
	"runtime/debug"
	"time"

	"accesscontrol/internal/config"
	"accesscontrol/internal/model"
	"accesscontrol/internal/service"
	"accesscontrol/internal/store"
	"accesscontrol/pkg/httpx"
	"accesscontrol/pkg/logger"
)

type Server struct {
	svc *service.Service
	log *logger.Logger
	cfg *config.Config
}

func NewServer(svc *service.Service, log *logger.Logger, cfg *config.Config) *Server {
	return &Server{svc: svc, log: log, cfg: cfg}
}

// Routes 注册全部路由并包裹中间件链：请求日志 -> 恢复 -> 鉴权 -> 限流。
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	s.registerZoneRoutes(mux)
	s.registerAccessPointRoutes(mux)
	s.registerReaderRoutes(mux)
	s.registerPersonRoutes(mux)
	s.registerCredentialRoutes(mux)
	s.registerAccessLogRoutes(mux)
	s.registerTimeRuleRoutes(mux)
	s.registerScheduleRoutes(mux)
	s.registerAlertRoutes(mux)
	s.registerStatsRoutes(mux)
	s.registerAccessRoutes(mux)
	s.registerExportRoutes(mux)

	mux.HandleFunc("GET /healthz", s.healthz)

	// 挂前端静态页面：GET / 返回 web/index.html
	mux.Handle("GET /", http.FileServer(http.Dir("web")))

	var h http.Handler = mux
	h = s.authMiddleware(h)
	h = s.rateLimitMiddleware(h)
	h = s.loggingMiddleware(h)
	h = s.recoveryMiddleware(h)
	return h
}

func (s *Server) maxPageSize() int {
	if s.cfg != nil && s.cfg.MaxPageSize > 0 {
		return s.cfg.MaxPageSize
	}
	return 100
}

func (s *Server) apiKey() string {
	if s.cfg != nil && s.cfg.APIKey != "" {
		return s.cfg.APIKey
	}
	return "access-control-secret"
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.log.Infof("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Errorf("panic: %v\n%s", rec, debug.Stack())
				httpx.InternalError(w, "服务器内部错误")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case model.IsValidationError(err):
		httpx.BadRequest(w, err.Error())
	case errors.Is(err, store.ErrNotFound):
		httpx.NotFound(w, err.Error())
	case errors.Is(err, store.ErrConflict):
		httpx.Conflict(w, err.Error())
	default:
		httpx.InternalError(w, err.Error())
	}
}
