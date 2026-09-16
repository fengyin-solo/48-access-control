// Package store 定义数据访问接口与内存实现。
package store

import (
	"errors"

	"accesscontrol/internal/model"
)

var (
	ErrNotFound = errors.New("记录不存在")
	ErrConflict = errors.New("记录已存在或状态冲突")
)

// Store 聚合全部实体的数据访问方法，便于测试时替换实现。
type Store interface {
	// Zone 区域
	CreateZone(z *model.Zone) error
	GetZone(id string) (*model.Zone, error)
	GetZoneByName(name string) (*model.Zone, error)
	ListZones() []*model.Zone
	UpdateZone(z *model.Zone) error
	DeleteZone(id string) error

	// AccessPoint 门禁点
	CreateAccessPoint(a *model.AccessPoint) error
	GetAccessPoint(id string) (*model.AccessPoint, error)
	GetAccessPointByName(name string) (*model.AccessPoint, error)
	ListAccessPoints() []*model.AccessPoint
	ListAccessPointsByZone(zoneID string) []*model.AccessPoint
	UpdateAccessPoint(a *model.AccessPoint) error
	DeleteAccessPoint(id string) error

	// Reader 读卡器
	CreateReader(r *model.Reader) error
	GetReader(id string) (*model.Reader, error)
	GetReaderByName(name string) (*model.Reader, error)
	ListReaders() []*model.Reader
	ListReadersByAccessPoint(apID string) []*model.Reader
	UpdateReader(r *model.Reader) error
	DeleteReader(id string) error

	// Person 人员
	CreatePerson(p *model.Person) error
	GetPerson(id string) (*model.Person, error)
	ListPersons() []*model.Person
	UpdatePerson(p *model.Person) error
	DeletePerson(id string) error

	// Credential 凭证
	CreateCredential(c *model.Credential) error
	GetCredential(id string) (*model.Credential, error)
	GetCredentialByCode(code string) (*model.Credential, error)
	ListCredentials() []*model.Credential
	ListCredentialsByPerson(personID string) []*model.Credential
	UpdateCredential(c *model.Credential) error
	DeleteCredential(id string) error

	// AccessLog 通行记录
	CreateAccessLog(l *model.AccessLog) error
	GetAccessLog(id string) (*model.AccessLog, error)
	ListAccessLogs() []*model.AccessLog
	ListAccessLogsByPoint(apID string) []*model.AccessLog
	DeleteAccessLogs(ids []string) int
	UpdateAccessLog(l *model.AccessLog) error
	DeleteAccessLog(id string) error

	// TimeRule 时段规则
	CreateTimeRule(t *model.TimeRule) error
	GetTimeRule(id string) (*model.TimeRule, error)
	GetTimeRuleByName(name string) (*model.TimeRule, error)
	ListTimeRules() []*model.TimeRule
	ListTimeRulesByAccessPoint(apID string) []*model.TimeRule
	UpdateTimeRule(t *model.TimeRule) error
	DeleteTimeRule(id string) error

	// Schedule 排班
	CreateSchedule(s *model.Schedule) error
	GetSchedule(id string) (*model.Schedule, error)
	ListSchedules() []*model.Schedule
	ListSchedulesByPerson(personID string) []*model.Schedule
	UpdateSchedule(s *model.Schedule) error
	DeleteSchedule(id string) error

	// Alert 告警
	CreateAlert(a *model.Alert) error
	GetAlert(id string) (*model.Alert, error)
	ListAlerts() []*model.Alert
	ListAlertsByAccessLog(logID string) []*model.Alert
	UpdateAlert(a *model.Alert) error
	DeleteAlert(id string) error
}
