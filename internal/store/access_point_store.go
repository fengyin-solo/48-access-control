package store

import "accesscontrol/internal/model"

func (s *MemoryStore) CreateAccessPoint(a *model.AccessPoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.accessPoints {
		if exist.Name == a.Name {
			return ErrConflict
		}
	}
	s.accessPoints[a.ID] = a
	return nil
}

func (s *MemoryStore) GetAccessPoint(id string) (*model.AccessPoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.accessPoints[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *MemoryStore) GetAccessPointByName(name string) (*model.AccessPoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.accessPoints {
		if a.Name == name {
			return a, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListAccessPoints() []*model.AccessPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.AccessPoint, 0, len(s.accessPoints))
	for _, a := range s.accessPoints {
		list = append(list, a)
	}
	return list
}

func (s *MemoryStore) UpdateAccessPoint(a *model.AccessPoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.accessPoints[a.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.accessPoints {
		if exist.ID != a.ID && exist.Name == a.Name {
			return ErrConflict
		}
	}
	s.accessPoints[a.ID] = a
	return nil
}

func (s *MemoryStore) DeleteAccessPoint(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.accessPoints[id]; !ok {
		return ErrNotFound
	}
	delete(s.accessPoints, id)
	return nil
}

// ListAccessPointsByZone 返回指定区域下的全部门禁点。
func (s *MemoryStore) ListAccessPointsByZone(zoneID string) []*model.AccessPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.AccessPoint, 0)
	for _, a := range s.accessPoints {
		if a.ZoneID == zoneID {
			list = append(list, a)
		}
	}
	return list
}
