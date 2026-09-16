package model

import (
	"strings"
	"time"
)

// 凭证状态常量。
const (
	CredentialActive   = "active"
	CredentialLost     = "lost"
	CredentialExpired  = "expired"
	CredentialDisabled = "disabled"
)

// 凭证类型常量。
const (
	CredentialCard       = "card"
	CredentialPIN        = "pin"
	CredentialBiometric  = "biometric"
)

// credentialTransitions 定义凭证状态机合法流转。
var credentialTransitions = map[string]map[string]bool{
	CredentialActive:   {CredentialLost: true, CredentialExpired: true, CredentialDisabled: true},
	CredentialLost:     {CredentialDisabled: true},
	CredentialExpired:  {CredentialDisabled: true},
	CredentialDisabled: {},
}

// CredentialCanTransition 校验凭证状态流转是否合法。
func CredentialCanTransition(from, to string) bool {
	if m, ok := credentialTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Credential 表示人员持有的门禁凭证（卡/密码/生物特征）。
type Credential struct {
	ID         string    `json:"id"`
	PersonID   string    `json:"person_id"`
	Type       string    `json:"type"`
	Code       string    `json:"code"`
	ValidFrom  time.Time `json:"valid_from"`
	ValidUntil time.Time `json:"valid_until"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (c *Credential) Validate() error {
	c.PersonID = strings.TrimSpace(c.PersonID)
	c.Type = strings.TrimSpace(c.Type)
	c.Code = strings.TrimSpace(c.Code)
	if c.PersonID == "" {
		return NewValidationError("person_id", "持证人员不能为空")
	}
	if c.Type == "" {
		c.Type = CredentialCard
	}
	if c.Type != CredentialCard && c.Type != CredentialPIN && c.Type != CredentialBiometric {
		return NewValidationError("type", "凭证类型不合法")
	}
	if c.Code == "" {
		return NewValidationError("code", "凭证编码不能为空")
	}
	if len(c.Code) > 128 {
		return NewValidationError("code", "凭证编码过长")
	}
	if c.ValidFrom.IsZero() {
		return NewValidationError("valid_from", "生效时间不能为空")
	}
	if c.ValidUntil.IsZero() {
		return NewValidationError("valid_until", "失效时间不能为空")
	}
	if !c.ValidUntil.After(c.ValidFrom) {
		return NewValidationError("valid_until", "失效时间必须晚于生效时间")
	}
	if c.Status == "" {
		c.Status = CredentialActive
	}
	if c.Status != CredentialActive && c.Status != CredentialLost &&
		c.Status != CredentialExpired && c.Status != CredentialDisabled {
		return NewValidationError("status", "凭证状态不合法")
	}
	return nil
}

// IsExpired 判断凭证当前是否已过有效期。
func (c *Credential) IsExpired(now time.Time) bool {
	return now.After(c.ValidUntil)
}

// IsValid 判断凭证在给定时刻是否在有效期内。
func (c *Credential) IsValid(now time.Time) bool {
	return !now.Before(c.ValidFrom) && !now.After(c.ValidUntil)
}

// CredentialFilter 凭证列表多条件筛选。
type CredentialFilter struct {
	PersonID string
	Type     string
	Status   string
	Keyword  string
}

func (f CredentialFilter) Match(c *Credential) bool {
	if f.PersonID != "" && c.PersonID != f.PersonID {
		return false
	}
	if f.Type != "" && c.Type != f.Type {
		return false
	}
	if f.Status != "" && c.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(c.Code), k) {
			return false
		}
	}
	return true
}
