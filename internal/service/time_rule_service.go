package service

import (
	"sort"
	"time"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/idgen"
)

// CreateTimeRule 创建时段规则，校验所属门禁点存在。
func (s *Service) CreateTimeRule(input model.TimeRule) (*model.TimeRule, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetAccessPoint(input.AccessPointID); err != nil {
		return nil, model.NewValidationError("access_point_id", "所属门禁点不存在")
	}
	now := time.Now()
	t := &model.TimeRule{
		ID:            idgen.Hex(),
		AccessPointID: input.AccessPointID,
		Name:          input.Name,
		Weekdays:      input.Weekdays,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		Effect:        input.Effect,
		Status:        input.Status,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.store.CreateTimeRule(t); err != nil {
		return nil, err
	}
	s.log.Infof("创建时段规则 %s (%s)", t.Name, t.ID)
	return t, nil
}

func (s *Service) GetTimeRule(id string) (*model.TimeRule, error) {
	return s.store.GetTimeRule(id)
}

func (s *Service) ListTimeRules(filter model.TimeRuleFilter, page, size int) ([]*model.TimeRule, int, error) {
	all := s.store.ListTimeRules()
	matched := make([]*model.TimeRule, 0, len(all))
	for _, t := range all {
		if filter.Match(t) {
			matched = append(matched, t)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	return paginate(matched, page, size), total, nil
}

func (s *Service) UpdateTimeRule(id string, input model.TimeRule) (*model.TimeRule, error) {
	t, err := s.store.GetTimeRule(id)
	if err != nil {
		return nil, err
	}
	if input.AccessPointID != "" {
		if _, err := s.store.GetAccessPoint(input.AccessPointID); err != nil {
			return nil, model.NewValidationError("access_point_id", "所属门禁点不存在")
		}
		t.AccessPointID = input.AccessPointID
	}
	if input.Name != "" {
		t.Name = input.Name
	}
	if input.Weekdays != nil {
		t.Weekdays = input.Weekdays
	}
	if input.StartTime != "" {
		t.StartTime = input.StartTime
	}
	if input.EndTime != "" {
		t.EndTime = input.EndTime
	}
	if input.Effect != "" {
		t.Effect = input.Effect
	}
	if input.Status != "" {
		t.Status = input.Status
	}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	t.UpdatedAt = time.Now()
	if err := s.store.UpdateTimeRule(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) DeleteTimeRule(id string) error {
	if _, err := s.store.GetTimeRule(id); err != nil {
		return err
	}
	return s.store.DeleteTimeRule(id)
}
