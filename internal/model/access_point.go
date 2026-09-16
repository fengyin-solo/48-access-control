package model

import (
	"strings"
	"time"
)

// 门禁点状态常量。
const (
	AccessPointOnline   = "online"
	AccessPointOffline  = "offline"
	AccessPointDisabled = "disabled"
)

// 门禁点类型常量。
const (
	AccessPointDoor     = "door"
	AccessPointGate     = "gate"
	AccessPointElevator = "elevator"
)

// 门禁点方向常量。
const (
	DirectionIn   = "in"
	DirectionOut  = "out"
	DirectionBoth = "both"
)

// accessPointTransitions 定义门禁点状态机合法流转。
var accessPointTransitions = map[string]map[string]bool{
	AccessPointOnline:   {AccessPointOffline: true, AccessPointDisabled: true},
	AccessPointOffline:  {AccessPointOnline: true, AccessPointDisabled: true},
	AccessPointDisabled: {},
}

// AccessPointCanTransition 校验门禁点状态流转是否合法。
func AccessPointCanTransition(from, to string) bool {
	if m, ok := accessPointTransitions[from]; ok {
		return m[to]
	}
	return false
}

// AccessPoint 表示一个可通行的门禁控制点。
type AccessPoint struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ZoneID    string    `json:"zone_id"`
	Type      string    `json:"type"`
	Direction string    `json:"direction"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *AccessPoint) Validate() error {
	a.Name = strings.TrimSpace(a.Name)
	a.ZoneID = strings.TrimSpace(a.ZoneID)
	a.Type = strings.TrimSpace(a.Type)
	a.Direction = strings.TrimSpace(a.Direction)
	if a.Name == "" {
		return NewValidationError("name", "门禁点名称不能为空")
	}
	if len(a.Name) > 64 {
		return NewValidationError("name", "门禁点名称过长")
	}
	if a.ZoneID == "" {
		return NewValidationError("zone_id", "所属区域不能为空")
	}
	if a.Type == "" {
		a.Type = AccessPointDoor
	}
	if a.Type != AccessPointDoor && a.Type != AccessPointGate && a.Type != AccessPointElevator {
		return NewValidationError("type", "门禁点类型不合法")
	}
	if a.Direction == "" {
		a.Direction = DirectionBoth
	}
	if a.Direction != DirectionIn && a.Direction != DirectionOut && a.Direction != DirectionBoth {
		return NewValidationError("direction", "通行方向不合法")
	}
	if a.Status == "" {
		a.Status = AccessPointOnline
	}
	if a.Status != AccessPointOnline && a.Status != AccessPointOffline && a.Status != AccessPointDisabled {
		return NewValidationError("status", "门禁点状态不合法")
	}
	return nil
}

// AccessPointFilter 门禁点列表多条件筛选。
type AccessPointFilter struct {
	ZoneID string
	Type   string
	Status string
	Keyword string
}

func (f AccessPointFilter) Match(a *AccessPoint) bool {
	if f.ZoneID != "" && a.ZoneID != f.ZoneID {
		return false
	}
	if f.Type != "" && a.Type != f.Type {
		return false
	}
	if f.Status != "" && a.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(a.Name), k) {
			return false
		}
	}
	return true
}
