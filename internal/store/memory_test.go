package store

import (
	"testing"
	"time"

	"accesscontrol/internal/model"
)

func newTestZone(id, name string) *model.Zone {
	now := time.Now()
	return &model.Zone{ID: id, Name: name, Level: 1, Status: model.ZoneActive, CreatedAt: now, UpdatedAt: now}
}

func newTestAccessPoint(id, name, zoneID string) *model.AccessPoint {
	now := time.Now()
	return &model.AccessPoint{ID: id, Name: name, ZoneID: zoneID, Type: model.AccessPointDoor, Direction: model.DirectionBoth, Status: model.AccessPointOnline, CreatedAt: now, UpdatedAt: now}
}

func newTestReader(id, name, apID string) *model.Reader {
	now := time.Now()
	return &model.Reader{ID: id, Name: name, AccessPointID: apID, Model: "M100", Status: model.ReaderOnline, CreatedAt: now, UpdatedAt: now}
}

func newTestPerson(id, name, phone string) *model.Person {
	now := time.Now()
	return &model.Person{ID: id, Name: name, Department: "安保", Phone: phone, Status: model.PersonActive, CreatedAt: now, UpdatedAt: now}
}

func newTestCredential(id, code, personID string) *model.Credential {
	now := time.Now()
	return &model.Credential{ID: id, PersonID: personID, Type: model.CredentialCard, Code: code, ValidFrom: now, ValidUntil: now.Add(24 * time.Hour), Status: model.CredentialActive, CreatedAt: now, UpdatedAt: now}
}

func newTestAccessLog(id, credID, apID string) *model.AccessLog {
	now := time.Now()
	return &model.AccessLog{ID: id, CredentialID: credID, AccessPointID: apID, AccessAt: now, Direction: model.DirectionIn, Result: model.AccessGranted, Reason: "ok", CreatedAt: now}
}

func newTestTimeRule(id, name, apID string) *model.TimeRule {
	now := time.Now()
	return &model.TimeRule{ID: id, AccessPointID: apID, Name: name, Weekdays: []int{1, 2, 3, 4, 5}, StartTime: "08:00", EndTime: "18:00", Effect: model.EffectAllow, Status: model.TimeRuleEnabled, CreatedAt: now, UpdatedAt: now}
}

func newTestSchedule(id, personID, zoneID string) *model.Schedule {
	now := time.Now()
	return &model.Schedule{ID: id, PersonID: personID, ZoneID: zoneID, StartAt: now, EndAt: now.Add(8 * time.Hour), Status: model.ScheduleActive, CreatedAt: now, UpdatedAt: now}
}

func newTestAlert(id, logID string) *model.Alert {
	now := time.Now()
	return &model.Alert{ID: id, AccessLogID: logID, Type: "access_denied", Level: model.AlertLevelWarn, Status: model.AlertOpen, CreatedAt: now, UpdatedAt: now}
}

func TestZoneCRUD(t *testing.T) {
	s := NewMemoryStore()
	z := newTestZone("z1", "办公区")
	if err := s.CreateZone(z); err != nil {
		t.Fatalf("create zone: %v", err)
	}
	// 冲突
	if err := s.CreateZone(newTestZone("z2", "办公区")); err != ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	got, err := s.GetZone("z1")
	if err != nil || got.Name != "办公区" {
		t.Fatalf("get zone failed: %v", err)
	}
	if len(s.ListZones()) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(s.ListZones()))
	}
	z.Name = "研发区"
	if err := s.UpdateZone(z); err != nil {
		t.Fatalf("update zone: %v", err)
	}
	if err := s.DeleteZone("z1"); err != nil {
		t.Fatalf("delete zone: %v", err)
	}
	if _, err := s.GetZone("z1"); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
	if err := s.UpdateZone(z); err != ErrNotFound {
		t.Fatalf("expected not found on update, got %v", err)
	}
	if err := s.DeleteZone("z1"); err != ErrNotFound {
		t.Fatalf("expected not found on delete, got %v", err)
	}
}

func TestAccessPointCRUD(t *testing.T) {
	s := NewMemoryStore()
	a := newTestAccessPoint("a1", "大门", "z1")
	if err := s.CreateAccessPoint(a); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateAccessPoint(newTestAccessPoint("a2", "大门", "z2")); err != ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	if _, err := s.GetAccessPoint("a1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(s.ListAccessPointsByZone("z1")) != 1 {
		t.Fatalf("expected 1 point in zone z1")
	}
	if err := s.DeleteAccessPoint("a1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetAccessPoint("a1"); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestReaderCRUD(t *testing.T) {
	s := NewMemoryStore()
	r := newTestReader("r1", "读卡器1", "a1")
	if err := s.CreateReader(r); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateReader(newTestReader("r2", "读卡器1", "a2")); err != ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	if _, err := s.GetReader("r1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(s.ListReadersByAccessPoint("a1")) != 1 {
		t.Fatalf("expected 1 reader on a1")
	}
	if err := s.UpdateReader(r); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.DeleteReader("r1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetReader("r1"); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestPersonCRUD(t *testing.T) {
	s := NewMemoryStore()
	p := newTestPerson("p1", "张三", "13800000001")
	if err := s.CreatePerson(p); err != nil {
		t.Fatalf("create: %v", err)
	}
	// 电话唯一性冲突
	if err := s.CreatePerson(newTestPerson("p2", "李四", "13800000001")); err != ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	if _, err := s.GetPerson("p1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(s.ListPersons()) != 1 {
		t.Fatalf("expected 1 person, got %d", len(s.ListPersons()))
	}
	p.Name = "张三丰"
	if err := s.UpdatePerson(p); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.DeletePerson("p1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetPerson("p1"); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestCredentialCRUDAndConflict(t *testing.T) {
	s := NewMemoryStore()
	c := newTestCredential("c1", "CARD-001", "p1")
	if err := s.CreateCredential(c); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateCredential(newTestCredential("c2", "CARD-001", "p1")); err != ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	if got, err := s.GetCredentialByCode("CARD-001"); err != nil || got.ID != "c1" {
		t.Fatalf("get by code failed: %v", err)
	}
	if len(s.ListCredentialsByPerson("p1")) != 1 {
		t.Fatalf("expected 1 credential for p1")
	}
	if err := s.DeleteCredential("c1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetCredential("c1"); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestAccessLogCRUDAndBatchDelete(t *testing.T) {
	s := NewMemoryStore()
	l1 := newTestAccessLog("l1", "c1", "a1")
	l2 := newTestAccessLog("l2", "c1", "a2")
	if err := s.CreateAccessLog(l1); err != nil {
		t.Fatalf("create l1: %v", err)
	}
	if err := s.CreateAccessLog(l2); err != nil {
		t.Fatalf("create l2: %v", err)
	}
	if len(s.ListAccessLogsByPoint("a1")) != 1 {
		t.Fatalf("expected 1 log for a1")
	}
	n := s.DeleteAccessLogs([]string{"l1", "l2", "missing"})
	if n != 2 {
		t.Fatalf("expected 2 deleted, got %d", n)
	}
	if len(s.ListAccessLogs()) != 0 {
		t.Fatalf("expected empty logs")
	}
	if _, err := s.GetAccessLog("l1"); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestTimeRuleCRUD(t *testing.T) {
	s := NewMemoryStore()
	tr := newTestTimeRule("t1", "工作日", "a1")
	if err := s.CreateTimeRule(tr); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateTimeRule(newTestTimeRule("t2", "工作日", "a2")); err != ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	if len(s.ListTimeRulesByAccessPoint("a1")) != 1 {
		t.Fatalf("expected 1 rule for a1")
	}
	if _, err := s.GetTimeRule("t1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if err := s.DeleteTimeRule("t1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetTimeRule("t1"); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestScheduleCRUD(t *testing.T) {
	s := NewMemoryStore()
	sc := newTestSchedule("s1", "p1", "z1")
	if err := s.CreateSchedule(sc); err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(s.ListSchedulesByPerson("p1")) != 1 {
		t.Fatalf("expected 1 schedule for p1")
	}
	if _, err := s.GetSchedule("s1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if err := s.UpdateSchedule(sc); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.DeleteSchedule("s1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetSchedule("s1"); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestAlertCRUD(t *testing.T) {
	s := NewMemoryStore()
	a := newTestAlert("al1", "l1")
	if err := s.CreateAlert(a); err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(s.ListAlertsByAccessLog("l1")) != 1 {
		t.Fatalf("expected 1 alert for l1")
	}
	if _, err := s.GetAlert("al1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	a.Status = model.AlertAcknowledged
	if err := s.UpdateAlert(a); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.DeleteAlert("al1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetAlert("al1"); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}
