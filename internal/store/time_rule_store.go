package store

import "accesscontrol/internal/model"

func (s *MemoryStore) CreateTimeRule(t *model.TimeRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.timeRules {
		if exist.Name == t.Name {
			return ErrConflict
		}
	}
	s.timeRules[t.ID] = t
	return nil
}

func (s *MemoryStore) GetTimeRule(id string) (*model.TimeRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.timeRules[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (s *MemoryStore) GetTimeRuleByName(name string) (*model.TimeRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.timeRules {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListTimeRules() []*model.TimeRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.TimeRule, 0, len(s.timeRules))
	for _, t := range s.timeRules {
		list = append(list, t)
	}
	return list
}

func (s *MemoryStore) UpdateTimeRule(t *model.TimeRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.timeRules[t.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.timeRules {
		if exist.ID != t.ID && exist.Name == t.Name {
			return ErrConflict
		}
	}
	s.timeRules[t.ID] = t
	return nil
}

func (s *MemoryStore) DeleteTimeRule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.timeRules[id]; !ok {
		return ErrNotFound
	}
	delete(s.timeRules, id)
	return nil
}

// ListTimeRulesByAccessPoint 返回指定门禁点的全部时段规则。
func (s *MemoryStore) ListTimeRulesByAccessPoint(apID string) []*model.TimeRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.TimeRule, 0)
	for _, t := range s.timeRules {
		if t.AccessPointID == apID {
			list = append(list, t)
		}
	}
	return list
}
