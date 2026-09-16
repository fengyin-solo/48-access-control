package service

import (
	"sort"
	"time"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/idgen"
)

// CreateSchedule 创建排班，校验人员与区域存在。
func (s *Service) CreateSchedule(input model.Schedule) (*model.Schedule, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetPerson(input.PersonID); err != nil {
		return nil, model.NewValidationError("person_id", "排班人员不存在")
	}
	if _, err := s.store.GetZone(input.ZoneID); err != nil {
		return nil, model.NewValidationError("zone_id", "排班区域不存在")
	}
	now := time.Now()
	sc := &model.Schedule{
		ID:        idgen.Hex(),
		PersonID:  input.PersonID,
		ZoneID:    input.ZoneID,
		StartAt:   input.StartAt,
		EndAt:     input.EndAt,
		Status:    input.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.CreateSchedule(sc); err != nil {
		return nil, err
	}
	s.log.Infof("创建排班 %s (%s)", sc.ID, sc.PersonID)
	return sc, nil
}

func (s *Service) GetSchedule(id string) (*model.Schedule, error) {
	return s.store.GetSchedule(id)
}

func (s *Service) ListSchedules(filter model.ScheduleFilter, page, size int) ([]*model.Schedule, int, error) {
	all := s.store.ListSchedules()
	matched := make([]*model.Schedule, 0, len(all))
	for _, sc := range all {
		if filter.Match(sc) {
			matched = append(matched, sc)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].StartAt.After(matched[j].StartAt)
	})
	total := len(matched)
	return paginate(matched, page, size), total, nil
}

func (s *Service) UpdateSchedule(id string, input model.Schedule) (*model.Schedule, error) {
	sc, err := s.store.GetSchedule(id)
	if err != nil {
		return nil, err
	}
	if input.PersonID != "" {
		if _, err := s.store.GetPerson(input.PersonID); err != nil {
			return nil, model.NewValidationError("person_id", "排班人员不存在")
		}
		sc.PersonID = input.PersonID
	}
	if input.ZoneID != "" {
		if _, err := s.store.GetZone(input.ZoneID); err != nil {
			return nil, model.NewValidationError("zone_id", "排班区域不存在")
		}
		sc.ZoneID = input.ZoneID
	}
	if !input.StartAt.IsZero() {
		sc.StartAt = input.StartAt
	}
	if !input.EndAt.IsZero() {
		sc.EndAt = input.EndAt
	}
	if input.Status != "" {
		sc.Status = input.Status
	}
	if err := sc.Validate(); err != nil {
		return nil, err
	}
	sc.UpdatedAt = time.Now()
	if err := s.store.UpdateSchedule(sc); err != nil {
		return nil, err
	}
	return sc, nil
}

func (s *Service) DeleteSchedule(id string) error {
	if _, err := s.store.GetSchedule(id); err != nil {
		return err
	}
	return s.store.DeleteSchedule(id)
}

// TransitionSchedule 执行排班状态流转（active -> ended）。
func (s *Service) TransitionSchedule(id, to string) (*model.Schedule, error) {
	sc, err := s.store.GetSchedule(id)
	if err != nil {
		return nil, err
	}
	if to == "" {
		return nil, model.NewValidationError("status", "目标状态不能为空")
	}
	if !model.ScheduleCanTransition(sc.Status, to) {
		return nil, model.NewValidationError("status",
			"排班状态不允许从 "+sc.Status+" 流转到 "+to)
	}
	sc.Status = to
	sc.UpdatedAt = time.Now()
	if err := s.store.UpdateSchedule(sc); err != nil {
		return nil, err
	}
	s.log.Infof("排班 %s 状态流转 %s -> %s", sc.ID, sc.Status, to)
	return sc, nil
}
