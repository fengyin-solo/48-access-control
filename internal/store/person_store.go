package store

import "accesscontrol/internal/model"

func (s *MemoryStore) CreatePerson(p *model.Person) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.persons {
		if exist.Phone == p.Phone {
			return ErrConflict
		}
	}
	s.persons[p.ID] = p
	return nil
}

func (s *MemoryStore) GetPerson(id string) (*model.Person, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.persons[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *MemoryStore) ListPersons() []*model.Person {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Person, 0, len(s.persons))
	for _, p := range s.persons {
		list = append(list, p)
	}
	return list
}

func (s *MemoryStore) UpdatePerson(p *model.Person) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.persons[p.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.persons {
		if exist.ID != p.ID && exist.Phone == p.Phone {
			return ErrConflict
		}
	}
	s.persons[p.ID] = p
	return nil
}

func (s *MemoryStore) DeletePerson(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.persons[id]; !ok {
		return ErrNotFound
	}
	delete(s.persons, id)
	return nil
}
