package store

import "accesscontrol/internal/model"

func (s *MemoryStore) CreateAlert(a *model.Alert) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts[a.ID] = a
	return nil
}

func (s *MemoryStore) GetAlert(id string) (*model.Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.alerts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *MemoryStore) ListAlerts() []*model.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Alert, 0, len(s.alerts))
	for _, a := range s.alerts {
		list = append(list, a)
	}
	return list
}

func (s *MemoryStore) UpdateAlert(a *model.Alert) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.alerts[a.ID]; !ok {
		return ErrNotFound
	}
	s.alerts[a.ID] = a
	return nil
}

func (s *MemoryStore) DeleteAlert(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.alerts[id]; !ok {
		return ErrNotFound
	}
	delete(s.alerts, id)
	return nil
}

// ListAlertsByAccessLog 返回关联指定通行记录的告警。
func (s *MemoryStore) ListAlertsByAccessLog(logID string) []*model.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Alert, 0)
	for _, a := range s.alerts {
		if a.AccessLogID == logID {
			list = append(list, a)
		}
	}
	return list
}
