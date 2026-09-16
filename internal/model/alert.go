package model

import (
	"strings"
	"time"
)

// 告警等级常量。
const (
	AlertLevelInfo     = "info"
	AlertLevelWarn     = "warn"
	AlertLevelCritical = "critical"
)

// 告警状态常量。
const (
	AlertOpen         = "open"
	AlertAcknowledged = "acknowledged"
	AlertResolved     = "resolved"
)

// alertTransitions 定义告警状态机合法流转。
var alertTransitions = map[string]map[string]bool{
	AlertOpen:         {AlertAcknowledged: true, AlertResolved: true},
	AlertAcknowledged: {AlertResolved: true},
	AlertResolved:     {},
}

// AlertCanTransition 校验告警状态流转是否合法。
func AlertCanTransition(from, to string) bool {
	if m, ok := alertTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Alert 表示通行异常触发的告警。
type Alert struct {
	ID          string    `json:"id"`
	AccessLogID string    `json:"access_log_id"`
	Type        string    `json:"type"`
	Level       string    `json:"level"`
	Status      string    `json:"status"`
	Handler     string    `json:"handler"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (a *Alert) Validate() error {
	a.AccessLogID = strings.TrimSpace(a.AccessLogID)
	a.Type = strings.TrimSpace(a.Type)
	a.Level = strings.TrimSpace(a.Level)
	a.Handler = strings.TrimSpace(a.Handler)
	if a.AccessLogID == "" {
		return NewValidationError("access_log_id", "关联通行记录不能为空")
	}
	if a.Type == "" {
		return NewValidationError("type", "告警类型不能为空")
	}
	if a.Level == "" {
		a.Level = AlertLevelInfo
	}
	if a.Level != AlertLevelInfo && a.Level != AlertLevelWarn && a.Level != AlertLevelCritical {
		return NewValidationError("level", "告警等级不合法")
	}
	if a.Status == "" {
		a.Status = AlertOpen
	}
	if a.Status != AlertOpen && a.Status != AlertAcknowledged && a.Status != AlertResolved {
		return NewValidationError("status", "告警状态不合法")
	}
	return nil
}

// AlertFilter 告警列表多条件筛选。
type AlertFilter struct {
	Type     string
	Level    string
	Status   string
	Handler  string
	Keyword  string
}

func (f AlertFilter) Match(a *Alert) bool {
	if f.Type != "" && a.Type != f.Type {
		return false
	}
	if f.Level != "" && a.Level != f.Level {
		return false
	}
	if f.Status != "" && a.Status != f.Status {
		return false
	}
	if f.Handler != "" && a.Handler != f.Handler {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(a.Type), k) {
			return false
		}
	}
	return true
}
