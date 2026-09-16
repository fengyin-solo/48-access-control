package service

import (
	"sort"
	"time"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/idgen"
)

// CreateAccessPoint 创建门禁点，校验所属区域存在。
func (s *Service) CreateAccessPoint(input model.AccessPoint) (*model.AccessPoint, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetZone(input.ZoneID); err != nil {
		return nil, model.NewValidationError("zone_id", "所属区域不存在")
	}
	now := time.Now()
	a := &model.AccessPoint{
		ID:        idgen.Hex(),
		Name:      input.Name,
		ZoneID:    input.ZoneID,
		Type:      input.Type,
		Direction: input.Direction,
		Status:    input.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.CreateAccessPoint(a); err != nil {
		return nil, err
	}
	s.log.Infof("创建门禁点 %s (%s)", a.Name, a.ID)
	return a, nil
}

func (s *Service) GetAccessPoint(id string) (*model.AccessPoint, error) {
	return s.store.GetAccessPoint(id)
}

func (s *Service) ListAccessPoints(filter model.AccessPointFilter, page, size int) ([]*model.AccessPoint, int, error) {
	all := s.store.ListAccessPoints()
	matched := make([]*model.AccessPoint, 0, len(all))
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

func (s *Service) UpdateAccessPoint(id string, input model.AccessPoint) (*model.AccessPoint, error) {
	a, err := s.store.GetAccessPoint(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		a.Name = input.Name
	}
	if input.ZoneID != "" {
		if _, err := s.store.GetZone(input.ZoneID); err != nil {
			return nil, model.NewValidationError("zone_id", "所属区域不存在")
		}
		a.ZoneID = input.ZoneID
	}
	if input.Type != "" {
		a.Type = input.Type
	}
	if input.Direction != "" {
		a.Direction = input.Direction
	}
	if input.Status != "" {
		a.Status = input.Status
	}
	if err := a.Validate(); err != nil {
		return nil, err
	}
	a.UpdatedAt = time.Now()
	if err := s.store.UpdateAccessPoint(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) DeleteAccessPoint(id string) error {
	if _, err := s.store.GetAccessPoint(id); err != nil {
		return err
	}
	return s.store.DeleteAccessPoint(id)
}

// TransitionAccessPoint 执行门禁点状态流转（状态机校验）。
func (s *Service) TransitionAccessPoint(id, to string) (*model.AccessPoint, error) {
	a, err := s.store.GetAccessPoint(id)
	if err != nil {
		return nil, err
	}
	if to == "" {
		return nil, model.NewValidationError("status", "目标状态不能为空")
	}
	if !model.AccessPointCanTransition(a.Status, to) {
		return nil, model.NewValidationError("status",
			"门禁点状态不允许从 "+a.Status+" 流转到 "+to)
	}
	a.Status = to
	a.UpdatedAt = time.Now()
	if err := s.store.UpdateAccessPoint(a); err != nil {
		return nil, err
	}
	s.log.Infof("门禁点 %s 状态流转 %s -> %s", a.Name, a.Status, to)
	return a, nil
}
