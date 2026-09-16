package service

import (
	"sort"
	"time"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/idgen"
)

// CreateReader 创建读卡器，校验所属门禁点存在。
func (s *Service) CreateReader(input model.Reader) (*model.Reader, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetAccessPoint(input.AccessPointID); err != nil {
		return nil, model.NewValidationError("access_point_id", "所属门禁点不存在")
	}
	now := time.Now()
	r := &model.Reader{
		ID:            idgen.Hex(),
		Name:          input.Name,
		AccessPointID: input.AccessPointID,
		Model:         input.Model,
		Status:        input.Status,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.store.CreateReader(r); err != nil {
		return nil, err
	}
	s.log.Infof("创建读卡器 %s (%s)", r.Name, r.ID)
	return r, nil
}

func (s *Service) GetReader(id string) (*model.Reader, error) {
	return s.store.GetReader(id)
}

func (s *Service) ListReaders(filter model.ReaderFilter, page, size int) ([]*model.Reader, int, error) {
	all := s.store.ListReaders()
	matched := make([]*model.Reader, 0, len(all))
	for _, r := range all {
		if filter.Match(r) {
			matched = append(matched, r)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	return paginate(matched, page, size), total, nil
}

func (s *Service) UpdateReader(id string, input model.Reader) (*model.Reader, error) {
	r, err := s.store.GetReader(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		r.Name = input.Name
	}
	if input.AccessPointID != "" {
		if _, err := s.store.GetAccessPoint(input.AccessPointID); err != nil {
			return nil, model.NewValidationError("access_point_id", "所属门禁点不存在")
		}
		r.AccessPointID = input.AccessPointID
	}
	if input.Model != "" {
		r.Model = input.Model
	}
	if input.Status != "" {
		r.Status = input.Status
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	r.UpdatedAt = time.Now()
	if err := s.store.UpdateReader(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Service) DeleteReader(id string) error {
	if _, err := s.store.GetReader(id); err != nil {
		return err
	}
	return s.store.DeleteReader(id)
}
