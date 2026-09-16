package store

import "accesscontrol/internal/model"

func (s *MemoryStore) CreateAccessLog(l *model.AccessLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessLogs[l.ID] = l
	return nil
}

func (s *MemoryStore) GetAccessLog(id string) (*model.AccessLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.accessLogs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return l, nil
}

func (s *MemoryStore) ListAccessLogs() []*model.AccessLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.AccessLog, 0, len(s.accessLogs))
	for _, l := range s.accessLogs {
		list = append(list, l)
	}
	return list
}

func (s *MemoryStore) UpdateAccessLog(l *model.AccessLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.accessLogs[l.ID]; !ok {
		return ErrNotFound
	}
	s.accessLogs[l.ID] = l
	return nil
}

func (s *MemoryStore) DeleteAccessLog(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.accessLogs[id]; !ok {
		return ErrNotFound
	}
	delete(s.accessLogs, id)
	return nil
}

// DeleteAccessLogs 批量删除指定 ID 集合的通行记录，返回实际删除数量。
func (s *MemoryStore) DeleteAccessLogs(ids []string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, id := range ids {
		if _, ok := s.accessLogs[id]; ok {
			delete(s.accessLogs, id)
			n++
		}
	}
	return n
}

// ListAccessLogsByPoint 返回指定门禁点的全部通行记录。
func (s *MemoryStore) ListAccessLogsByPoint(apID string) []*model.AccessLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.AccessLog, 0)
	for _, l := range s.accessLogs {
		if l.AccessPointID == apID {
			list = append(list, l)
		}
	}
	return list
}
