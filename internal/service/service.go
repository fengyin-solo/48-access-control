// Package service 实现业务逻辑层。
package service

import (
	"accesscontrol/internal/config"
	"accesscontrol/internal/store"
	"accesscontrol/pkg/logger"
)

// Service 聚合存储、日志与配置，承载全部业务逻辑。
type Service struct {
	store store.Store
	log   *logger.Logger
	cfg   *config.Config
}

func New(st store.Store, log *logger.Logger, cfg *config.Config) *Service {
	return &Service{store: st, log: log, cfg: cfg}
}

// Store 暴露底层存储，供统计等内部模块使用。
func (s *Service) Store() store.Store { return s.store }
