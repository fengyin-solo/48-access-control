package service

import (
	"time"

	"accesscontrol/internal/model"
)

// ExportSnapshot 表示数据导出的汇总快照。
type ExportSnapshot struct {
	GeneratedAt  time.Time              `json:"generated_at"`
	Overview     *Overview              `json:"overview"`
	Zones        []*model.Zone          `json:"zones"`
	AccessPoints []*model.AccessPoint   `json:"access_points"`
	Readers      []*model.Reader        `json:"readers"`
	Persons      []*model.Person        `json:"persons"`
	Credentials  []*model.Credential    `json:"credentials"`
	AccessLogs   []*model.AccessLog     `json:"access_logs"`
	TimeRules    []*model.TimeRule      `json:"time_rules"`
	Schedules    []*model.Schedule      `json:"schedules"`
	Alerts       []*model.Alert         `json:"alerts"`
	ZoneTraffic  []ZoneTraffic          `json:"zone_traffic"`
	TopPoints    []PointTraffic         `json:"top_points"`
}

// BuildExportSnapshot 构建全量数据导出快照。
func (s *Service) BuildExportSnapshot() (*ExportSnapshot, error) {
	ov, err := s.StatsOverview()
	if err != nil {
		return nil, err
	}
	zoneTraffic, err := s.ZoneTrafficStats()
	if err != nil {
		return nil, err
	}
	topPoints, err := s.PointTrafficTopN(10)
	if err != nil {
		return nil, err
	}

	snap := &ExportSnapshot{
		GeneratedAt:  time.Now(),
		Overview:     ov,
		Zones:        s.store.ListZones(),
		AccessPoints: s.store.ListAccessPoints(),
		Readers:      s.store.ListReaders(),
		Persons:      s.store.ListPersons(),
		Credentials:  s.store.ListCredentials(),
		AccessLogs:   s.store.ListAccessLogs(),
		TimeRules:    s.store.ListTimeRules(),
		Schedules:    s.store.ListSchedules(),
		Alerts:       s.store.ListAlerts(),
		ZoneTraffic:  zoneTraffic,
		TopPoints:    topPoints,
	}
	return snap, nil
}
