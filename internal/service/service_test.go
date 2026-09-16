package service

import (
	"testing"
	"time"

	"accesscontrol/internal/config"
	"accesscontrol/internal/model"
	"accesscontrol/internal/store"
	"accesscontrol/pkg/logger"
)

func newTestService() *Service {
	cfg := &config.Config{MaxPageSize: 100, APIKey: "test-key", RateLimit: 100}
	log := logger.NewLevel(logger.LevelError)
	return New(store.NewMemoryStore(), log, cfg)
}

func setupZoneAndPoint(t *testing.T, s *Service) (string, string) {
	t.Helper()
	zone, err := s.CreateZone(model.Zone{Name: "测试区", Level: 1})
	if err != nil {
		t.Fatalf("create zone: %v", err)
	}
	ap, err := s.CreateAccessPoint(model.AccessPoint{Name: "大门", ZoneID: zone.ID})
	if err != nil {
		t.Fatalf("create access point: %v", err)
	}
	return zone.ID, ap.ID
}

func setupPersonAndCredential(t *testing.T, s *Service, validUntil time.Time) (string, string) {
	t.Helper()
	p, err := s.CreatePerson(model.Person{Name: "张三", Department: "安保", Phone: "13800000001"})
	if err != nil {
		t.Fatalf("create person: %v", err)
	}
	c, err := s.CreateCredential(model.Credential{
		PersonID:   p.ID,
		Type:       model.CredentialCard,
		Code:       "CARD-001",
		ValidFrom:  time.Now().Add(-24 * time.Hour),
		ValidUntil: validUntil,
	})
	if err != nil {
		t.Fatalf("create credential: %v", err)
	}
	return p.ID, c.ID
}

func TestAuthorizeAccessGranted(t *testing.T) {
	s := newTestService()
	_, apID := setupZoneAndPoint(t, s)
	_, credID := setupPersonAndCredential(t, s, time.Now().Add(24*time.Hour))

	result, err := s.AuthorizeAccess(AccessRequest{
		CredentialID:  credID,
		AccessPointID: apID,
		Direction:     model.DirectionIn,
	})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if !result.Granted {
		t.Fatalf("expected granted, reason=%s", result.Reason)
	}
	if result.Log == nil || result.Log.Result != model.AccessGranted {
		t.Fatalf("expected granted log")
	}
	if result.Alert != nil {
		t.Fatalf("expected no alert on granted access")
	}
	if len(s.store.ListAccessLogs()) != 1 {
		t.Fatalf("expected 1 access log")
	}
}

func TestAuthorizeAccessDeniedExpired(t *testing.T) {
	s := newTestService()
	_, apID := setupZoneAndPoint(t, s)
	_, credID := setupPersonAndCredential(t, s, time.Now().Add(-time.Hour))

	result, err := s.AuthorizeAccess(AccessRequest{
		CredentialID:  credID,
		AccessPointID: apID,
		Direction:     model.DirectionIn,
	})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if result.Granted {
		t.Fatalf("expected denied for expired credential")
	}
	if result.Log == nil || result.Log.Result != model.AccessDenied {
		t.Fatalf("expected denied log")
	}
	if result.Alert == nil {
		t.Fatalf("expected alert on denied access")
	}
	// 过期自动置 expired
	c, _ := s.store.GetCredential(credID)
	if c.Status != model.CredentialExpired {
		t.Fatalf("expected credential expired, got %s", c.Status)
	}
}

func TestAuthorizeAccessDeniedDisabledPoint(t *testing.T) {
	s := newTestService()
	_, apID := setupZoneAndPoint(t, s)
	_, credID := setupPersonAndCredential(t, s, time.Now().Add(24*time.Hour))

	if _, err := s.TransitionAccessPoint(apID, model.AccessPointDisabled); err != nil {
		t.Fatalf("transition point: %v", err)
	}
	result, err := s.AuthorizeAccess(AccessRequest{
		CredentialID:  credID,
		AccessPointID: apID,
		Direction:     model.DirectionIn,
	})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if result.Granted {
		t.Fatalf("expected denied for disabled point")
	}
	if result.Alert == nil {
		t.Fatalf("expected alert")
	}
}

func TestAuthorizeAccessTimeRuleDeny(t *testing.T) {
	s := newTestService()
	_, apID := setupZoneAndPoint(t, s)
	_, credID := setupPersonAndCredential(t, s, time.Now().Add(24*time.Hour))

	// 建一条全周拒绝对应时段的规则
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	_, err := s.CreateTimeRule(model.TimeRule{
		AccessPointID: apID,
		Name:          "全周禁止",
		Weekdays:      []int{weekday},
		StartTime:     "00:00",
		EndTime:       "23:59",
		Effect:        model.EffectDeny,
	})
	if err != nil {
		t.Fatalf("create time rule: %v", err)
	}

	result, err := s.AuthorizeAccess(AccessRequest{
		CredentialID:  credID,
		AccessPointID: apID,
		Direction:     model.DirectionIn,
		AccessAt:      now,
	})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if result.Granted {
		t.Fatalf("expected denied by time rule, reason=%s", result.Reason)
	}
}

func TestCheckValidityExpired(t *testing.T) {
	s := newTestService()
	_, credID := setupPersonAndCredential(t, s, time.Now().Add(-time.Minute))

	if _, err := s.CheckValidity(credID, time.Now()); err == nil {
		t.Fatalf("expected validation error for expired credential")
	}
	c, _ := s.store.GetCredential(credID)
	if c.Status != model.CredentialExpired {
		t.Fatalf("expected expired status, got %s", c.Status)
	}
}

func TestCheckValidityNotYetValid(t *testing.T) {
	s := newTestService()
	p, _ := s.CreatePerson(model.Person{Name: "李四", Department: "安保", Phone: "13800000002"})
	c, err := s.CreateCredential(model.Credential{
		PersonID:   p.ID,
		Type:       model.CredentialCard,
		Code:       "CARD-002",
		ValidFrom:  time.Now().Add(time.Hour),
		ValidUntil: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create credential: %v", err)
	}
	if _, err := s.CheckValidity(c.ID, time.Now()); err == nil {
		t.Fatalf("expected error for not-yet-valid credential")
	}
}

func TestCredentialTransition(t *testing.T) {
	s := newTestService()
	_, credID := setupPersonAndCredential(t, s, time.Now().Add(24*time.Hour))

	c, err := s.TransitionCredential(credID, model.CredentialLost)
	if err != nil {
		t.Fatalf("transition to lost: %v", err)
	}
	if c.Status != model.CredentialLost {
		t.Fatalf("expected lost, got %s", c.Status)
	}
	// 非法流转 lost -> active 应失败
	if _, err := s.TransitionCredential(credID, model.CredentialActive); err == nil {
		t.Fatalf("expected error for illegal transition")
	}
}

func TestAlertTransition(t *testing.T) {
	s := newTestService()
	_, apID := setupZoneAndPoint(t, s)
	_, credID := setupPersonAndCredential(t, s, time.Now().Add(-time.Hour))

	result, err := s.AuthorizeAccess(AccessRequest{CredentialID: credID, AccessPointID: apID, Direction: model.DirectionIn})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	alertID := result.Alert.ID

	a, err := s.TransitionAlert(alertID, model.AlertAcknowledged, "王五")
	if err != nil {
		t.Fatalf("transition to acknowledged: %v", err)
	}
	if a.Status != model.AlertAcknowledged || a.Handler != "王五" {
		t.Fatalf("unexpected alert state: %s/%s", a.Status, a.Handler)
	}
	a, err = s.TransitionAlert(alertID, model.AlertResolved, "王五")
	if err != nil {
		t.Fatalf("transition to resolved: %v", err)
	}
	if a.Status != model.AlertResolved {
		t.Fatalf("expected resolved, got %s", a.Status)
	}
	// resolved 之后不能再流转
	if _, err := s.TransitionAlert(alertID, model.AlertOpen, ""); err == nil {
		t.Fatalf("expected error for illegal alert transition")
	}
}

func TestAccessPointTransition(t *testing.T) {
	s := newTestService()
	_, apID := setupZoneAndPoint(t, s)

	if _, err := s.TransitionAccessPoint(apID, model.AccessPointOffline); err != nil {
		t.Fatalf("to offline: %v", err)
	}
	if _, err := s.TransitionAccessPoint(apID, model.AccessPointOnline); err != nil {
		t.Fatalf("back to online: %v", err)
	}
	if _, err := s.TransitionAccessPoint(apID, model.AccessPointDisabled); err != nil {
		t.Fatalf("to disabled: %v", err)
	}
	// disabled 不能流转到 online
	if _, err := s.TransitionAccessPoint(apID, model.AccessPointOnline); err == nil {
		t.Fatalf("expected error for disabled -> online")
	}
}

func TestScheduleTransition(t *testing.T) {
	s := newTestService()
	zoneID, _ := setupZoneAndPoint(t, s)
	p, _ := s.CreatePerson(model.Person{Name: "赵六", Department: "安保", Phone: "13800000003"})

	sc, err := s.CreateSchedule(model.Schedule{
		PersonID: p.ID,
		ZoneID:   zoneID,
		StartAt:  time.Now(),
		EndAt:    time.Now().Add(8 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	sc, err = s.TransitionSchedule(sc.ID, model.ScheduleEnded)
	if err != nil {
		t.Fatalf("transition schedule: %v", err)
	}
	if sc.Status != model.ScheduleEnded {
		t.Fatalf("expected ended, got %s", sc.Status)
	}
	if _, err := s.TransitionSchedule(sc.ID, model.ScheduleActive); err == nil {
		t.Fatalf("expected error for ended -> active")
	}
}

func TestCrossEntityValidation(t *testing.T) {
	s := newTestService()
	// AccessPoint 引用不存在的 zone
	if _, err := s.CreateAccessPoint(model.AccessPoint{Name: "门", ZoneID: "no-such-zone"}); err == nil {
		t.Fatalf("expected error for missing zone")
	}
	// Credential 引用不存在的 person
	if _, err := s.CreateCredential(model.Credential{
		PersonID:   "no-such-person",
		Type:       model.CredentialCard,
		Code:       "CARD-X",
		ValidFrom:  time.Now(),
		ValidUntil: time.Now().Add(time.Hour),
	}); err == nil {
		t.Fatalf("expected error for missing person")
	}
	// Reader 引用不存在的 access point
	if _, err := s.CreateReader(model.Reader{Name: "读卡器", AccessPointID: "no-such-ap", Model: "M1"}); err == nil {
		t.Fatalf("expected error for missing access point")
	}
}

func TestStatsOverviewAndTraffic(t *testing.T) {
	s := newTestService()
	zoneID, apID := setupZoneAndPoint(t, s)
	_, credID := setupPersonAndCredential(t, s, time.Now().Add(24*time.Hour))

	// 两次通行：一次 granted，一次 denied（先禁用门禁点）
	if _, err := s.AuthorizeAccess(AccessRequest{CredentialID: credID, AccessPointID: apID, Direction: model.DirectionIn}); err != nil {
		t.Fatalf("authorize granted: %v", err)
	}
	if _, err := s.TransitionAccessPoint(apID, model.AccessPointDisabled); err != nil {
		t.Fatalf("disable point: %v", err)
	}
	if _, err := s.AuthorizeAccess(AccessRequest{CredentialID: credID, AccessPointID: apID, Direction: model.DirectionIn}); err != nil {
		t.Fatalf("authorize denied: %v", err)
	}

	ov, err := s.StatsOverview()
	if err != nil {
		t.Fatalf("overview: %v", err)
	}
	if ov.AccessLogCount != 2 || ov.GrantedCount != 1 || ov.DeniedCount != 1 {
		t.Fatalf("unexpected overview: %+v", ov)
	}
	if ov.OpenAlertCount != 1 {
		t.Fatalf("expected 1 open alert, got %d", ov.OpenAlertCount)
	}

	traffic, err := s.ZoneTrafficStats()
	if err != nil {
		t.Fatalf("zone traffic: %v", err)
	}
	if len(traffic) != 1 || traffic[0].Total != 2 || traffic[0].ZoneID != zoneID {
		t.Fatalf("unexpected zone traffic: %+v", traffic)
	}

	top, err := s.PointTrafficTopN(5)
	if err != nil {
		t.Fatalf("top points: %v", err)
	}
	if len(top) != 1 || top[0].Total != 2 {
		t.Fatalf("unexpected top points: %+v", top)
	}
}

func TestBatchIssueAndDisableCredentials(t *testing.T) {
	s := newTestService()
	p, _ := s.CreatePerson(model.Person{Name: "钱七", Department: "安保", Phone: "13800000004"})

	inputs := []model.Credential{
		{PersonID: p.ID, Type: model.CredentialCard, Code: "B-001", ValidFrom: time.Now(), ValidUntil: time.Now().Add(time.Hour)},
		{PersonID: p.ID, Type: model.CredentialCard, Code: "B-002", ValidFrom: time.Now(), ValidUntil: time.Now().Add(time.Hour)},
		{PersonID: p.ID, Type: model.CredentialCard, Code: "B-001", ValidFrom: time.Now(), ValidUntil: time.Now().Add(time.Hour)}, // 重复 code
	}
	created, failures, err := s.BatchIssueCredentials(inputs)
	if err != nil {
		t.Fatalf("batch issue: %v", err)
	}
	if len(created) != 2 || len(failures) != 1 {
		t.Fatalf("expected 2 created, 1 failure; got %d/%d", len(created), len(failures))
	}

	ids := []string{created[0].ID, created[1].ID, "missing"}
	updated, failures, err := s.BatchDisableCredentials(ids)
	if err != nil {
		t.Fatalf("batch disable: %v", err)
	}
	if len(updated) != 2 || len(failures) != 1 {
		t.Fatalf("expected 2 updated, 1 failure; got %d/%d", len(updated), len(failures))
	}
}

func TestBatchUpdateAlertStatus(t *testing.T) {
	s := newTestService()
	_, apID := setupZoneAndPoint(t, s)
	_, credID := setupPersonAndCredential(t, s, time.Now().Add(-time.Hour))

	var alertIDs []string
	for i := 0; i < 3; i++ {
		result, err := s.AuthorizeAccess(AccessRequest{CredentialID: credID, AccessPointID: apID, Direction: model.DirectionIn})
		if err != nil {
			t.Fatalf("authorize: %v", err)
		}
		alertIDs = append(alertIDs, result.Alert.ID)
	}
	updated, failures, err := s.BatchUpdateAlertStatus(alertIDs, model.AlertResolved, "孙八")
	if err != nil {
		t.Fatalf("batch update: %v", err)
	}
	if len(updated) != 3 || len(failures) != 0 {
		t.Fatalf("expected 3 updated, got %d/%d", len(updated), len(failures))
	}
}
