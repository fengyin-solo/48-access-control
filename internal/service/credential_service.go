package service

import (
	"sort"
	"time"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/idgen"
)

// CreateCredential 创建凭证，校验所属人员存在、编码唯一与有效期。
func (s *Service) CreateCredential(input model.Credential) (*model.Credential, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetPerson(input.PersonID); err != nil {
		return nil, model.NewValidationError("person_id", "持证人员不存在")
	}
	now := time.Now()
	c := &model.Credential{
		ID:         idgen.Hex(),
		PersonID:   input.PersonID,
		Type:       input.Type,
		Code:       input.Code,
		ValidFrom:  input.ValidFrom,
		ValidUntil: input.ValidUntil,
		Status:     input.Status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.store.CreateCredential(c); err != nil {
		return nil, err
	}
	s.log.Infof("创建凭证 %s (%s)", c.Code, c.ID)
	return c, nil
}

func (s *Service) GetCredential(id string) (*model.Credential, error) {
	c, err := s.store.GetCredential(id)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) ListCredentials(filter model.CredentialFilter, page, size int) ([]*model.Credential, int, error) {
	all := s.store.ListCredentials()
	matched := make([]*model.Credential, 0, len(all))
	for _, c := range all {
		if filter.Match(c) {
			matched = append(matched, c)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	return paginate(matched, page, size), total, nil
}

func (s *Service) UpdateCredential(id string, input model.Credential) (*model.Credential, error) {
	c, err := s.store.GetCredential(id)
	if err != nil {
		return nil, err
	}
	if input.PersonID != "" {
		if _, err := s.store.GetPerson(input.PersonID); err != nil {
			return nil, model.NewValidationError("person_id", "持证人员不存在")
		}
		c.PersonID = input.PersonID
	}
	if input.Type != "" {
		c.Type = input.Type
	}
	if input.Code != "" {
		c.Code = input.Code
	}
	if !input.ValidFrom.IsZero() {
		c.ValidFrom = input.ValidFrom
	}
	if !input.ValidUntil.IsZero() {
		c.ValidUntil = input.ValidUntil
	}
	if input.Status != "" {
		c.Status = input.Status
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	c.UpdatedAt = time.Now()
	if err := s.store.UpdateCredential(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) DeleteCredential(id string) error {
	if _, err := s.store.GetCredential(id); err != nil {
		return err
	}
	return s.store.DeleteCredential(id)
}

// TransitionCredential 执行凭证状态流转（状态机校验）。
func (s *Service) TransitionCredential(id, to string) (*model.Credential, error) {
	c, err := s.store.GetCredential(id)
	if err != nil {
		return nil, err
	}
	if to == "" {
		return nil, model.NewValidationError("status", "目标状态不能为空")
	}
	if !model.CredentialCanTransition(c.Status, to) {
		return nil, model.NewValidationError("status",
			"凭证状态不允许从 "+c.Status+" 流转到 "+to)
	}
	c.Status = to
	c.UpdatedAt = time.Now()
	if err := s.store.UpdateCredential(c); err != nil {
		return nil, err
	}
	s.log.Infof("凭证 %s 状态流转 %s -> %s", c.Code, c.Status, to)
	return c, nil
}

// CheckValidity 校验凭证有效性（状态 + 有效期），过期自动置 expired。
func (s *Service) CheckValidity(id string, now time.Time) (*model.Credential, error) {
	c, err := s.store.GetCredential(id)
	if err != nil {
		return nil, err
	}
	if c.Status == model.CredentialActive && c.IsExpired(now) {
		c.Status = model.CredentialExpired
		c.UpdatedAt = now
		if err := s.store.UpdateCredential(c); err != nil {
			return nil, err
		}
		s.log.Warnf("凭证 %s 已过期，自动置为 expired", c.Code)
		return c, model.NewValidationError("status", "凭证已过期")
	}
	if c.Status != model.CredentialActive {
		return c, model.NewValidationError("status", "凭证状态为 "+c.Status)
	}
	if !c.IsValid(now) {
		return c, model.NewValidationError("valid_until", "凭证不在有效期内")
	}
	return c, nil
}

// BatchIssueCredentials 批量发放凭证，返回成功与失败详情。
func (s *Service) BatchIssueCredentials(inputs []model.Credential) ([]*model.Credential, []BatchFailure, error) {
	if len(inputs) == 0 {
		return nil, nil, model.NewValidationError("credentials", "凭证列表不能为空")
	}
	var created []*model.Credential
	var failures []BatchFailure
	for i, input := range inputs {
		c, err := s.CreateCredential(input)
		if err != nil {
			failures = append(failures, BatchFailure{Index: i, Error: err.Error()})
			continue
		}
		created = append(created, c)
	}
	return created, failures, nil
}

// BatchDisableCredentials 批量停用凭证，返回成功停用的凭证。
func (s *Service) BatchDisableCredentials(ids []string) ([]*model.Credential, []BatchFailure, error) {
	if len(ids) == 0 {
		return nil, nil, model.NewValidationError("ids", "凭证 ID 列表不能为空")
	}
	var updated []*model.Credential
	var failures []BatchFailure
	for i, id := range ids {
		c, err := s.TransitionCredential(id, model.CredentialDisabled)
		if err != nil {
			failures = append(failures, BatchFailure{Index: i, ID: id, Error: err.Error()})
			continue
		}
		updated = append(updated, c)
	}
	return updated, failures, nil
}

// BatchFailure 描述批量操作中单条的失败信息。
type BatchFailure struct {
	Index int    `json:"index"`
	ID    string `json:"id,omitempty"`
	Error string `json:"error"`
}
