package service

import (
	"sort"
	"time"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/idgen"
)

// CreatePerson 创建人员，校验联系电话唯一性。
func (s *Service) CreatePerson(input model.Person) (*model.Person, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	now := time.Now()
	p := &model.Person{
		ID:         idgen.Hex(),
		Name:       input.Name,
		Department: input.Department,
		Phone:      input.Phone,
		Status:     input.Status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.store.CreatePerson(p); err != nil {
		return nil, err
	}
	s.log.Infof("创建人员 %s (%s)", p.Name, p.ID)
	return p, nil
}

func (s *Service) GetPerson(id string) (*model.Person, error) {
	return s.store.GetPerson(id)
}

func (s *Service) ListPersons(filter model.PersonFilter, page, size int) ([]*model.Person, int, error) {
	all := s.store.ListPersons()
	matched := make([]*model.Person, 0, len(all))
	for _, p := range all {
		if filter.Match(p) {
			matched = append(matched, p)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	return paginate(matched, page, size), total, nil
}

func (s *Service) UpdatePerson(id string, input model.Person) (*model.Person, error) {
	p, err := s.store.GetPerson(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		p.Name = input.Name
	}
	if input.Department != "" {
		p.Department = input.Department
	}
	if input.Phone != "" {
		p.Phone = input.Phone
	}
	if input.Status != "" {
		p.Status = input.Status
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	p.UpdatedAt = time.Now()
	if err := s.store.UpdatePerson(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) DeletePerson(id string) error {
	if _, err := s.store.GetPerson(id); err != nil {
		return err
	}
	return s.store.DeletePerson(id)
}
