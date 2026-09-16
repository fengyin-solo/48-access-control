package model

import (
	"strings"
	"time"
)

// 排班状态常量。
const (
	ScheduleActive = "active"
	ScheduleEnded  = "ended"
)

// scheduleTransitions 定义排班状态机合法流转。
var scheduleTransitions = map[string]map[string]bool{
	ScheduleActive: {ScheduleEnded: true},
	ScheduleEnded:  {},
}

// ScheduleCanTransition 校验排班状态流转是否合法。
func ScheduleCanTransition(from, to string) bool {
	if m, ok := scheduleTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Schedule 表示人员在指定区域内的工作排班。
type Schedule struct {
	ID        string    `json:"id"`
	PersonID  string    `json:"person_id"`
	ZoneID    string    `json:"zone_id"`
	StartAt   time.Time `json:"start_at"`
	EndAt     time.Time `json:"end_at"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Schedule) Validate() error {
	s.PersonID = strings.TrimSpace(s.PersonID)
	s.ZoneID = strings.TrimSpace(s.ZoneID)
	if s.PersonID == "" {
		return NewValidationError("person_id", "排班人员不能为空")
	}
	if s.ZoneID == "" {
		return NewValidationError("zone_id", "排班区域不能为空")
	}
	if s.StartAt.IsZero() {
		return NewValidationError("start_at", "开始时间不能为空")
	}
	if s.EndAt.IsZero() {
		return NewValidationError("end_at", "结束时间不能为空")
	}
	if !s.EndAt.After(s.StartAt) {
		return NewValidationError("end_at", "结束时间必须晚于开始时间")
	}
	if s.Status == "" {
		s.Status = ScheduleActive
	}
	if s.Status != ScheduleActive && s.Status != ScheduleEnded {
		return NewValidationError("status", "排班状态不合法")
	}
	return nil
}

// ScheduleFilter 排班列表多条件筛选。
type ScheduleFilter struct {
	PersonID string
	ZoneID   string
	Status   string
}

func (f ScheduleFilter) Match(s *Schedule) bool {
	if f.PersonID != "" && s.PersonID != f.PersonID {
		return false
	}
	if f.ZoneID != "" && s.ZoneID != f.ZoneID {
		return false
	}
	if f.Status != "" && s.Status != f.Status {
		return false
	}
	return true
}
