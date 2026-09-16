package service

import (
	"sort"
	"time"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/idgen"
)

// AccessRequest 表示一次通行鉴权请求。
type AccessRequest struct {
	CredentialID  string    `json:"credential_id"`
	CredentialCode string   `json:"credential_code"`
	AccessPointID string    `json:"access_point_id"`
	Direction     string    `json:"direction"`
	AccessAt      time.Time `json:"access_at"`
}

// AccessResult 表示通行鉴权结果。
type AccessResult struct {
	Granted bool           `json:"granted"`
	Reason  string         `json:"reason"`
	Log     *model.AccessLog `json:"log"`
	Alert   *model.Alert   `json:"alert,omitempty"`
}

// AuthorizeAccess 执行通行授权判定：综合凭证状态、有效期与门禁点时段规则，
// 判定 granted/denied，写入通行记录，并在拒绝时生成告警。
func (s *Service) AuthorizeAccess(req AccessRequest) (*AccessResult, error) {
	if req.AccessPointID == "" {
		return nil, model.NewValidationError("access_point_id", "门禁点不能为空")
	}
	now := req.AccessAt
	if now.IsZero() {
		now = time.Now()
	}

	// 1. 解析凭证
	var cred *model.Credential
	var err error
	if req.CredentialID != "" {
		cred, err = s.store.GetCredential(req.CredentialID)
	} else if req.CredentialCode != "" {
		cred, err = s.store.GetCredentialByCode(req.CredentialCode)
	} else {
		return nil, model.NewValidationError("credential_id", "凭证 ID 或编码不能为空")
	}
	if err != nil {
		return nil, model.NewValidationError("credential_id", "凭证不存在")
	}

	// 2. 校验门禁点存在与状态
	ap, err := s.store.GetAccessPoint(req.AccessPointID)
	if err != nil {
		return nil, model.NewValidationError("access_point_id", "门禁点不存在")
	}

	result := &AccessResult{}
	reason := ""

	// 3. 逐项判定拒绝原因
	switch {
	case cred.Status != model.CredentialActive:
		reason = "凭证状态为 " + cred.Status
	case cred.IsExpired(now):
		// 过期自动置 expired
		cred.Status = model.CredentialExpired
		cred.UpdatedAt = now
		_ = s.store.UpdateCredential(cred)
		reason = "凭证已过期"
	case !cred.IsValid(now):
		reason = "凭证不在有效期内"
	case ap.Status == model.AccessPointDisabled:
		reason = "门禁点已禁用"
	case ap.Status == model.AccessPointOffline:
		reason = "门禁点离线"
	default:
		reason = s.evaluateTimeRules(ap.ID, now)
	}

	if reason == "" {
		result.Granted = true
		reason = "授权通过"
	} else {
		result.Granted = false
	}
	result.Reason = reason

	// 4. 写入通行记录
	direction := req.Direction
	if direction == "" {
		direction = model.DirectionIn
	}
	logEntry := &model.AccessLog{
		ID:            idgen.Hex(),
		CredentialID:  cred.ID,
		AccessPointID: ap.ID,
		AccessAt:      now,
		Direction:     direction,
		Result:        model.AccessDenied,
		Reason:        reason,
		CreatedAt:     now,
	}
	if result.Granted {
		logEntry.Result = model.AccessGranted
	}
	if err := s.store.CreateAccessLog(logEntry); err != nil {
		return nil, err
	}
	result.Log = logEntry

	// 5. 拒绝时生成告警
	if !result.Granted {
		alert := s.createAlertForDeny(logEntry)
		result.Alert = alert
	}

	return result, nil
}

// evaluateTimeRules 检查门禁点时段规则，返回非空拒绝原因；空串表示允许。
func (s *Service) evaluateTimeRules(apID string, now time.Time) string {
	rules := s.store.ListTimeRulesByAccessPoint(apID)
	if len(rules) == 0 {
		return ""
	}
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	var denyReason string
	matched := false
	for _, rule := range rules {
		if rule.Status != model.TimeRuleEnabled {
			continue
		}
		if !rule.Covers(weekday, now.Hour(), now.Minute()) {
			continue
		}
		matched = true
		if rule.Effect == model.EffectDeny {
			denyReason = "命中拒绝时段规则: " + rule.Name
		}
	}
	if denyReason != "" {
		return denyReason
	}
	if matched {
		return ""
	}
	return "当前时段不在任何允许规则内"
}

// createAlertForDeny 为拒绝的通行记录生成告警。
func (s *Service) createAlertForDeny(logEntry *model.AccessLog) *model.Alert {
	now := time.Now()
	alert := &model.Alert{
		ID:          idgen.Hex(),
		AccessLogID: logEntry.ID,
		Type:        "access_denied",
		Level:       model.AlertLevelWarn,
		Status:      model.AlertOpen,
		Handler:     "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.store.CreateAlert(alert); err != nil {
		s.log.Errorf("创建告警失败: %v", err)
		return nil
	}
	s.log.Warnf("生成告警 %s 关联通行记录 %s", alert.ID, logEntry.ID)
	return alert
}

// CreateAccessLog 直接写入一条通行记录（手工补录），校验外键存在。
func (s *Service) CreateAccessLog(input model.AccessLog) (*model.AccessLog, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetCredential(input.CredentialID); err != nil {
		return nil, model.NewValidationError("credential_id", "凭证不存在")
	}
	if _, err := s.store.GetAccessPoint(input.AccessPointID); err != nil {
		return nil, model.NewValidationError("access_point_id", "门禁点不存在")
	}
	l := &model.AccessLog{
		ID:            idgen.Hex(),
		CredentialID:  input.CredentialID,
		AccessPointID: input.AccessPointID,
		AccessAt:      input.AccessAt,
		Direction:     input.Direction,
		Result:        input.Result,
		Reason:        input.Reason,
		CreatedAt:     time.Now(),
	}
	if err := s.store.CreateAccessLog(l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *Service) GetAccessLog(id string) (*model.AccessLog, error) {
	return s.store.GetAccessLog(id)
}

func (s *Service) ListAccessLogs(filter model.AccessLogFilter, page, size int) ([]*model.AccessLog, int, error) {
	all := s.store.ListAccessLogs()
	matched := make([]*model.AccessLog, 0, len(all))
	for _, l := range all {
		if filter.Match(l) {
			matched = append(matched, l)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].AccessAt.After(matched[j].AccessAt)
	})
	total := len(matched)
	return paginate(matched, page, size), total, nil
}

// DeleteAccessLog 删除单条通行记录。
func (s *Service) DeleteAccessLog(id string) error {
	if _, err := s.store.GetAccessLog(id); err != nil {
		return err
	}
	return s.store.DeleteAccessLog(id)
}

// BatchDeleteAccessLogs 批量删除通行记录，返回删除数量。
func (s *Service) BatchDeleteAccessLogs(ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, model.NewValidationError("ids", "通行记录 ID 列表不能为空")
	}
	n := s.store.DeleteAccessLogs(ids)
	s.log.Infof("批量删除通行记录 %d 条", n)
	return n, nil
}
