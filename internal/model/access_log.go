package model

import (
	"strings"
	"time"
)

// 通行结果常量。
const (
	AccessGranted = "granted"
	AccessDenied  = "denied"
)

// AccessLog 表示一次门禁通行尝试的记录。
type AccessLog struct {
	ID            string    `json:"id"`
	CredentialID  string    `json:"credential_id"`
	AccessPointID string    `json:"access_point_id"`
	AccessAt      time.Time `json:"access_at"`
	Direction     string    `json:"direction"`
	Result        string    `json:"result"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
}

func (l *AccessLog) Validate() error {
	l.CredentialID = strings.TrimSpace(l.CredentialID)
	l.AccessPointID = strings.TrimSpace(l.AccessPointID)
	l.Direction = strings.TrimSpace(l.Direction)
	l.Reason = strings.TrimSpace(l.Reason)
	if l.CredentialID == "" {
		return NewValidationError("credential_id", "凭证不能为空")
	}
	if l.AccessPointID == "" {
		return NewValidationError("access_point_id", "门禁点不能为空")
	}
	if l.AccessAt.IsZero() {
		l.AccessAt = time.Now()
	}
	if l.Direction == "" {
		l.Direction = DirectionIn
	}
	if l.Direction != DirectionIn && l.Direction != DirectionOut {
		return NewValidationError("direction", "通行方向不合法")
	}
	if l.Result == "" {
		l.Result = AccessGranted
	}
	if l.Result != AccessGranted && l.Result != AccessDenied {
		return NewValidationError("result", "通行结果不合法")
	}
	return nil
}

// AccessLogFilter 通行记录列表多条件筛选。
type AccessLogFilter struct {
	CredentialID  string
	AccessPointID string
	Result        string
	Direction     string
	From          time.Time
	To            time.Time
}

func (f AccessLogFilter) Match(l *AccessLog) bool {
	if f.CredentialID != "" && l.CredentialID != f.CredentialID {
		return false
	}
	if f.AccessPointID != "" && l.AccessPointID != f.AccessPointID {
		return false
	}
	if f.Result != "" && l.Result != f.Result {
		return false
	}
	if f.Direction != "" && l.Direction != f.Direction {
		return false
	}
	if !f.From.IsZero() && l.AccessAt.Before(f.From) {
		return false
	}
	if !f.To.IsZero() && l.AccessAt.After(f.To) {
		return false
	}
	return true
}
