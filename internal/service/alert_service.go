package service

import (
	"sort"
	"time"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/idgen"
)

// CreateAlert 创建告警，校验关联通行记录存在。
func (s *Service) CreateAlert(input model.Alert) (*model.Alert, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetAccessLog(input.AccessLogID); err != nil {
		return nil, model.NewValidationError("access_log_id", "关联通行记录不存在")
	}
	now := time.Now()
	a := &model.Alert{
		ID:          idgen.Hex(),
		AccessLogID: input.AccessLogID,
		Type:        input.Type,
		Level:       input.Level,
		Status:      input.Status,
		Handler:     input.Handler,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.store.CreateAlert(a); err != nil {
		return nil, err
	}
	s.log.Infof("创建告警 %s (%s)", a.Type, a.ID)
	return a, nil
}

func (s *Service) GetAlert(id string) (*model.Alert, error) {
	return s.store.GetAlert(id)
}

func (s *Service) ListAlerts(filter model.AlertFilter, page, size int) ([]*model.Alert, int, error) {
	all := s.store.ListAlerts()
	matched := make([]*model.Alert, 0, len(all))
	for _, a := range all {
		if filter.Match(a) {
			matched = append(matched, a)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	return paginate(matched, page, size), total, nil
}

func (s *Service) UpdateAlert(id string, input model.Alert) (*model.Alert, error) {
	a, err := s.store.GetAlert(id)
	if err != nil {
		return nil, err
	}
	if input.Type != "" {
		a.Type = input.Type
	}
	if input.Level != "" {
		a.Level = input.Level
	}
	if input.Status != "" {
		a.Status = input.Status
	}
	if input.Handler != "" {
		a.Handler = input.Handler
	}
	if err := a.Validate(); err != nil {
		return nil, err
	}
	a.UpdatedAt = time.Now()
	if err := s.store.UpdateAlert(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) DeleteAlert(id string) error {
	if _, err := s.store.GetAlert(id); err != nil {
		return err
	}
	return s.store.DeleteAlert(id)
}

// TransitionAlert 执行告警状态流转（状态机校验）。
func (s *Service) TransitionAlert(id, to, handler string) (*model.Alert, error) {
	a, err := s.store.GetAlert(id)
	if err != nil {
		return nil, err
	}
	if to == "" {
		return nil, model.NewValidationError("status", "目标状态不能为空")
	}
	if !model.AlertCanTransition(a.Status, to) {
		return nil, model.NewValidationError("status",
			"告警状态不允许从 "+a.Status+" 流转到 "+to)
	}
	a.Status = to
	if handler != "" {
		a.Handler = handler
	}
	a.UpdatedAt = time.Now()
	if err := s.store.UpdateAlert(a); err != nil {
		return nil, err
	}
	s.log.Infof("告警 %s 状态流转 %s -> %s", a.ID, a.Status, to)
	return a, nil
}

// BatchUpdateAlertStatus 批量更新告警状态，返回成功与失败详情。
func (s *Service) BatchUpdateAlertStatus(ids []string, to, handler string) ([]*model.Alert, []BatchFailure, error) {
	if len(ids) == 0 {
		return nil, nil, model.NewValidationError("ids", "告警 ID 列表不能为空")
	}
	var updated []*model.Alert
	var failures []BatchFailure
	for i, id := range ids {
		a, err := s.TransitionAlert(id, to, handler)
		if err != nil {
			failures = append(failures, BatchFailure{Index: i, ID: id, Error: err.Error()})
			continue
		}
		updated = append(updated, a)
	}
	return updated, failures, nil
}
