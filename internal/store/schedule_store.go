package store

import "accesscontrol/internal/model"

func (s *MemoryStore) CreateSchedule(sc *model.Schedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedules[sc.ID] = sc
	return nil
}

func (s *MemoryStore) GetSchedule(id string) (*model.Schedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok := s.schedules[id]
	if !ok {
		return nil, ErrNotFound
	}
	return sc, nil
}

func (s *MemoryStore) ListSchedules() []*model.Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Schedule, 0, len(s.schedules))
	for _, sc := range s.schedules {
		list = append(list, sc)
	}
	return list
}

func (s *MemoryStore) UpdateSchedule(sc *model.Schedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.schedules[sc.ID]; !ok {
		return ErrNotFound
	}
	s.schedules[sc.ID] = sc
	return nil
}

func (s *MemoryStore) DeleteSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.schedules[id]; !ok {
		return ErrNotFound
	}
	delete(s.schedules, id)
	return nil
}

// ListSchedulesByPerson 返回指定人员的全部排班。
func (s *MemoryStore) ListSchedulesByPerson(personID string) []*model.Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Schedule, 0)
	for _, sc := range s.schedules {
		if sc.PersonID == personID {
			list = append(list, sc)
		}
	}
	return list
}
