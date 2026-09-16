package model

import (
	"strings"
	"time"
)

// 人员状态常量。
const (
	PersonActive   = "active"
	PersonDisabled = "disabled"
)

// Person 表示系统内可被授予通行权限的人员。
type Person struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Department string    `json:"department"`
	Phone      string    `json:"phone"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (p *Person) Validate() error {
	p.Name = strings.TrimSpace(p.Name)
	p.Department = strings.TrimSpace(p.Department)
	p.Phone = strings.TrimSpace(p.Phone)
	if p.Name == "" {
		return NewValidationError("name", "人员姓名不能为空")
	}
	if len(p.Name) > 64 {
		return NewValidationError("name", "人员姓名过长")
	}
	if p.Department == "" {
		return NewValidationError("department", "所属部门不能为空")
	}
	if p.Phone == "" {
		return NewValidationError("phone", "联系电话不能为空")
	}
	if len(p.Phone) > 20 {
		return NewValidationError("phone", "联系电话过长")
	}
	if p.Status == "" {
		p.Status = PersonActive
	}
	if p.Status != PersonActive && p.Status != PersonDisabled {
		return NewValidationError("status", "人员状态不合法")
	}
	return nil
}

// PersonFilter 人员列表多条件筛选。
type PersonFilter struct {
	Department string
	Status     string
	Keyword    string
}

func (f PersonFilter) Match(p *Person) bool {
	if f.Department != "" && p.Department != f.Department {
		return false
	}
	if f.Status != "" && p.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(p.Name), k) &&
			!strings.Contains(strings.ToLower(p.Phone), k) {
			return false
		}
	}
	return true
}
