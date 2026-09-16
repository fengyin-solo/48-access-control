package model

import (
	"fmt"
	"strings"
	"time"
)

// 时段规则状态常量。
const (
	TimeRuleEnabled  = "enabled"
	TimeRuleDisabled = "disabled"
)

// 时段规则效果常量。
const (
	EffectAllow = "allow"
	EffectDeny  = "deny"
)

// TimeRule 表示门禁点在特定星期/时段内的通行策略。
type TimeRule struct {
	ID            string    `json:"id"`
	AccessPointID string    `json:"access_point_id"`
	Name          string    `json:"name"`
	Weekdays      []int     `json:"weekdays"`
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	Effect        string    `json:"effect"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (t *TimeRule) Validate() error {
	t.AccessPointID = strings.TrimSpace(t.AccessPointID)
	t.Name = strings.TrimSpace(t.Name)
	t.StartTime = strings.TrimSpace(t.StartTime)
	t.EndTime = strings.TrimSpace(t.EndTime)
	t.Effect = strings.TrimSpace(t.Effect)
	if t.AccessPointID == "" {
		return NewValidationError("access_point_id", "所属门禁点不能为空")
	}
	if t.Name == "" {
		return NewValidationError("name", "时段规则名称不能为空")
	}
	if len(t.Weekdays) == 0 {
		return NewValidationError("weekdays", "至少选择一个星期")
	}
	seen := make(map[int]bool, len(t.Weekdays))
	for _, d := range t.Weekdays {
		if d < 1 || d > 7 {
			return NewValidationError("weekdays", fmt.Sprintf("星期取值必须在 1-7 之间，收到 %d", d))
		}
		if seen[d] {
			return NewValidationError("weekdays", fmt.Sprintf("星期 %d 重复", d))
		}
		seen[d] = true
	}
	if t.StartTime == "" {
		return NewValidationError("start_time", "开始时间不能为空")
	}
	if t.EndTime == "" {
		return NewValidationError("end_time", "结束时间不能为空")
	}
	if !validTime(t.StartTime) {
		return NewValidationError("start_time", "开始时间格式须为 HH:MM")
	}
	if !validTime(t.EndTime) {
		return NewValidationError("end_time", "结束时间格式须为 HH:MM")
	}
	if t.Effect == "" {
		t.Effect = EffectAllow
	}
	if t.Effect != EffectAllow && t.Effect != EffectDeny {
		return NewValidationError("effect", "时段规则效果不合法")
	}
	if t.Status == "" {
		t.Status = TimeRuleEnabled
	}
	if t.Status != TimeRuleEnabled && t.Status != TimeRuleDisabled {
		return NewValidationError("status", "时段规则状态不合法")
	}
	return nil
}

// validTime 校验 "HH:MM" 格式与取值范围。
func validTime(s string) bool {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) != 2 || len(parts[1]) != 2 {
		return false
	}
	var h, m int
	if _, err := fmt.Sscanf(parts[0], "%d", &h); err != nil {
		return false
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &m); err != nil {
		return false
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return false
	}
	return true
}

// Covers 判断该规则是否覆盖指定星期与时刻。
func (t *TimeRule) Covers(weekday int, hour, minute int) bool {
	for _, d := range t.Weekdays {
		if d != weekday {
			continue
		}
		var sh, sm, eh, em int
		if _, err := fmt.Sscanf(t.StartTime, "%d:%d", &sh, &sm); err != nil {
			return false
		}
		if _, err := fmt.Sscanf(t.EndTime, "%d:%d", &eh, &em); err != nil {
			return false
		}
		cur := hour*60 + minute
		start := sh*60 + sm
		end := eh*60 + em
		if start <= end {
			return cur >= start && cur <= end
		}
		// 跨午夜时段
		return cur >= start || cur <= end
	}
	return false
}

// TimeRuleFilter 时段规则列表多条件筛选。
type TimeRuleFilter struct {
	AccessPointID string
	Effect        string
	Status        string
	Keyword       string
}

func (f TimeRuleFilter) Match(t *TimeRule) bool {
	if f.AccessPointID != "" && t.AccessPointID != f.AccessPointID {
		return false
	}
	if f.Effect != "" && t.Effect != f.Effect {
		return false
	}
	if f.Status != "" && t.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(t.Name), k) {
			return false
		}
	}
	return true
}
