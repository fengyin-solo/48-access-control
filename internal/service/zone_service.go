package service

import (
	"sort"
	"time"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/idgen"
)

// CreateZone 创建区域，校验名称唯一性与父级存在性。
func (s *Service) CreateZone(input model.Zone) (*model.Zone, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if input.ParentID != "" {
		if _, err := s.store.GetZone(input.ParentID); err != nil {
			return nil, model.NewValidationError("parent_id", "父级区域不存在")
		}
	}
	now := time.Now()
	z := &model.Zone{
		ID:        idgen.Hex(),
		Name:      input.Name,
		Level:     input.Level,
		ParentID:  input.ParentID,
		Status:    input.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.CreateZone(z); err != nil {
		return nil, err
	}
	s.log.Infof("创建区域 %s (%s)", z.Name, z.ID)
	return z, nil
}

func (s *Service) GetZone(id string) (*model.Zone, error) {
	return s.store.GetZone(id)
}

// ListZones 按筛选条件过滤、按创建时间倒序、分页返回。
func (s *Service) ListZones(filter model.ZoneFilter, page, size int) ([]*model.Zone, int, error) {
	all := s.store.ListZones()
	matched := make([]*model.Zone, 0, len(all))
	for _, z := range all {
		if filter.Match(z) {
			matched = append(matched, z)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	return paginate(matched, page, size), total, nil
}

func (s *Service) UpdateZone(id string, input model.Zone) (*model.Zone, error) {
	z, err := s.store.GetZone(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		z.Name = input.Name
	}
	if input.Level > 0 {
		z.Level = input.Level
	}
	if input.ParentID != "" {
		if input.ParentID == id {
			return nil, model.NewValidationError("parent_id", "区域不能以自身为父级")
		}
		if _, err := s.store.GetZone(input.ParentID); err != nil {
			return nil, model.NewValidationError("parent_id", "父级区域不存在")
		}
		z.ParentID = input.ParentID
	}
	if input.Status != "" {
		z.Status = input.Status
	}
	if err := z.Validate(); err != nil {
		return nil, err
	}
	z.UpdatedAt = time.Now()
	if err := s.store.UpdateZone(z); err != nil {
		return nil, err
	}
	return z, nil
}

func (s *Service) DeleteZone(id string) error {
	if _, err := s.store.GetZone(id); err != nil {
		return err
	}
	return s.store.DeleteZone(id)
}

// paginate 对已排序切片做分页截取。
func paginate[T any](list []T, page, size int) []T {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	start := (page - 1) * size
	if start >= len(list) {
		return []T{}
	}
	end := start + size
	if end > len(list) {
		end = len(list)
	}
	return list[start:end]
}
