package store

import "accesscontrol/internal/model"

func (s *MemoryStore) CreateZone(z *model.Zone) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.zones {
		if exist.Name == z.Name {
			return ErrConflict
		}
	}
	s.zones[z.ID] = z
	return nil
}

func (s *MemoryStore) GetZone(id string) (*model.Zone, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	z, ok := s.zones[id]
	if !ok {
		return nil, ErrNotFound
	}
	return z, nil
}

func (s *MemoryStore) GetZoneByName(name string) (*model.Zone, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, z := range s.zones {
		if z.Name == name {
			return z, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListZones() []*model.Zone {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Zone, 0, len(s.zones))
	for _, z := range s.zones {
		list = append(list, z)
	}
	return list
}

func (s *MemoryStore) UpdateZone(z *model.Zone) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.zones[z.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.zones {
		if exist.ID != z.ID && exist.Name == z.Name {
			return ErrConflict
		}
	}
	s.zones[z.ID] = z
	return nil
}

func (s *MemoryStore) DeleteZone(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.zones[id]; !ok {
		return ErrNotFound
	}
	delete(s.zones, id)
	return nil
}
