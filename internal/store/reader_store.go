package store

import "accesscontrol/internal/model"

func (s *MemoryStore) CreateReader(r *model.Reader) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.readers {
		if exist.Name == r.Name {
			return ErrConflict
		}
	}
	s.readers[r.ID] = r
	return nil
}

func (s *MemoryStore) GetReader(id string) (*model.Reader, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.readers[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *MemoryStore) GetReaderByName(name string) (*model.Reader, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.readers {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListReaders() []*model.Reader {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Reader, 0, len(s.readers))
	for _, r := range s.readers {
		list = append(list, r)
	}
	return list
}

func (s *MemoryStore) UpdateReader(r *model.Reader) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.readers[r.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.readers {
		if exist.ID != r.ID && exist.Name == r.Name {
			return ErrConflict
		}
	}
	s.readers[r.ID] = r
	return nil
}

func (s *MemoryStore) DeleteReader(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.readers[id]; !ok {
		return ErrNotFound
	}
	delete(s.readers, id)
	return nil
}

// ListReadersByAccessPoint 返回指定门禁点下的全部读卡器。
func (s *MemoryStore) ListReadersByAccessPoint(apID string) []*model.Reader {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Reader, 0)
	for _, r := range s.readers {
		if r.AccessPointID == apID {
			list = append(list, r)
		}
	}
	return list
}
