package store

import "accesscontrol/internal/model"

func (s *MemoryStore) CreateCredential(c *model.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.credentials {
		if exist.Code == c.Code {
			return ErrConflict
		}
	}
	s.credentials[c.ID] = c
	return nil
}

func (s *MemoryStore) GetCredential(id string) (*model.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.credentials[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *MemoryStore) GetCredentialByCode(code string) (*model.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.credentials {
		if c.Code == code {
			return c, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListCredentials() []*model.Credential {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Credential, 0, len(s.credentials))
	for _, c := range s.credentials {
		list = append(list, c)
	}
	return list
}

func (s *MemoryStore) UpdateCredential(c *model.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.credentials[c.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.credentials {
		if exist.ID != c.ID && exist.Code == c.Code {
			return ErrConflict
		}
	}
	s.credentials[c.ID] = c
	return nil
}

func (s *MemoryStore) DeleteCredential(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.credentials[id]; !ok {
		return ErrNotFound
	}
	delete(s.credentials, id)
	return nil
}

// ListCredentialsByPerson 返回指定人员的全部凭证。
func (s *MemoryStore) ListCredentialsByPerson(personID string) []*model.Credential {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Credential, 0)
	for _, c := range s.credentials {
		if c.PersonID == personID {
			list = append(list, c)
		}
	}
	return list
}
