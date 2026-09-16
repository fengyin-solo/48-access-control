package model

import (
	"strings"
	"time"
)

// 读卡器状态常量。
const (
	ReaderOnline  = "online"
	ReaderOffline = "offline"
)

// Reader 表示安装在门禁点上的读卡/识别设备。
type Reader struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	AccessPointID string    `json:"access_point_id"`
	Model         string    `json:"model"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (r *Reader) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.AccessPointID = strings.TrimSpace(r.AccessPointID)
	r.Model = strings.TrimSpace(r.Model)
	if r.Name == "" {
		return NewValidationError("name", "读卡器名称不能为空")
	}
	if len(r.Name) > 64 {
		return NewValidationError("name", "读卡器名称过长")
	}
	if r.AccessPointID == "" {
		return NewValidationError("access_point_id", "所属门禁点不能为空")
	}
	if r.Model == "" {
		return NewValidationError("model", "读卡器型号不能为空")
	}
	if r.Status == "" {
		r.Status = ReaderOnline
	}
	if r.Status != ReaderOnline && r.Status != ReaderOffline {
		return NewValidationError("status", "读卡器状态不合法")
	}
	return nil
}

// ReaderFilter 读卡器列表多条件筛选。
type ReaderFilter struct {
	AccessPointID string
	Status        string
	Keyword       string
}

func (f ReaderFilter) Match(r *Reader) bool {
	if f.AccessPointID != "" && r.AccessPointID != f.AccessPointID {
		return false
	}
	if f.Status != "" && r.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(r.Name), k) &&
			!strings.Contains(strings.ToLower(r.Model), k) {
			return false
		}
	}
	return true
}
