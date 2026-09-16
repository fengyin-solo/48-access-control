package model

import (
	"strings"
	"time"
)

// 区域状态常量。
const (
	ZoneActive   = "active"
	ZoneInactive = "inactive"
)

// Zone 表示物理或逻辑上的安防区域。
type Zone struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Level     int       `json:"level"`
	ParentID  string    `json:"parent_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (z *Zone) Validate() error {
	z.Name = strings.TrimSpace(z.Name)
	z.ParentID = strings.TrimSpace(z.ParentID)
	if z.Name == "" {
		return NewValidationError("name", "区域名称不能为空")
	}
	if len(z.Name) > 64 {
		return NewValidationError("name", "区域名称过长")
	}
	if z.Level < 0 {
		return NewValidationError("level", "区域层级不能为负数")
	}
	if z.Level > 10 {
		return NewValidationError("level", "区域层级不能超过 10")
	}
	if z.ParentID != "" && z.ParentID == z.ID {
		return NewValidationError("parent_id", "区域不能以自身为父级")
	}
	if z.Status == "" {
		z.Status = ZoneActive
	}
	if z.Status != ZoneActive && z.Status != ZoneInactive {
		return NewValidationError("status", "区域状态不合法")
	}
	return nil
}

// ZoneFilter 区域列表多条件筛选。
type ZoneFilter struct {
	Name   string
	Status string
	Level  int
	Parent string
}

func (f ZoneFilter) Match(z *Zone) bool {
	if f.Status != "" && z.Status != f.Status {
		return false
	}
	if f.Level > 0 && z.Level != f.Level {
		return false
	}
	if f.Parent != "" && z.ParentID != f.Parent {
		return false
	}
	if f.Name != "" {
		k := strings.ToLower(strings.TrimSpace(f.Name))
		if k != "" && !strings.Contains(strings.ToLower(z.Name), k) {
			return false
		}
	}
	return true
}
