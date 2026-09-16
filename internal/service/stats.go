package service

import (
	"sort"
	"time"

	"accesscontrol/internal/model"
)

// ZoneTraffic 表示单个区域的通行量统计。
type ZoneTraffic struct {
	ZoneID   string `json:"zone_id"`
	ZoneName string `json:"zone_name"`
	Granted  int    `json:"granted"`
	Denied   int    `json:"denied"`
	Total    int    `json:"total"`
}

// PointTraffic 表示单个门禁点的通行量统计。
type PointTraffic struct {
	AccessPointID   string `json:"access_point_id"`
	AccessPointName string `json:"access_point_name"`
	Total           int    `json:"total"`
}

// CredentialStatusStat 表示凭证状态分布项。
type CredentialStatusStat struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// AlertStat 表示告警统计项。
type AlertStat struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

// Overview 表示看板汇总快照。
type Overview struct {
	ZoneCount        int                    `json:"zone_count"`
	AccessPointCount int                    `json:"access_point_count"`
	ReaderCount      int                    `json:"reader_count"`
	PersonCount      int                    `json:"person_count"`
	CredentialCount  int                    `json:"credential_count"`
	AccessLogCount   int                    `json:"access_log_count"`
	GrantedCount     int                    `json:"granted_count"`
	DeniedCount      int                    `json:"denied_count"`
	OpenAlertCount   int                    `json:"open_alert_count"`
	CredentialStats  []CredentialStatusStat `json:"credential_stats"`
	AlertByLevel     []AlertStat            `json:"alert_by_level"`
	AlertByStatus    []AlertStat            `json:"alert_by_status"`
	GeneratedAt      time.Time              `json:"generated_at"`
}

// StatsOverview 返回看板所需的全部汇总指标。
func (s *Service) StatsOverview() (*Overview, error) {
	zones := s.store.ListZones()
	points := s.store.ListAccessPoints()
	readers := s.store.ListReaders()
	persons := s.store.ListPersons()
	creds := s.store.ListCredentials()
	logs := s.store.ListAccessLogs()
	alerts := s.store.ListAlerts()

	ov := &Overview{
		ZoneCount:        len(zones),
		AccessPointCount: len(points),
		ReaderCount:      len(readers),
		PersonCount:      len(persons),
		CredentialCount:  len(creds),
		AccessLogCount:   len(logs),
		GeneratedAt:      time.Now(),
	}

	// 凭证状态分布
	credMap := make(map[string]int)
	for _, c := range creds {
		credMap[c.Status]++
	}
	ov.CredentialStats = make([]CredentialStatusStat, 0, len(credMap))
	for status, count := range credMap {
		ov.CredentialStats = append(ov.CredentialStats, CredentialStatusStat{Status: status, Count: count})
	}
	sort.Slice(ov.CredentialStats, func(i, j int) bool {
		return ov.CredentialStats[i].Status < ov.CredentialStats[j].Status
	})

	// 告警按等级/状态统计
	levelMap := make(map[string]int)
	statusMap := make(map[string]int)
	openCount := 0
	for _, a := range alerts {
		levelMap[a.Level]++
		statusMap[a.Status]++
		if a.Status == model.AlertOpen {
			openCount++
		}
	}
	ov.OpenAlertCount = openCount
	ov.AlertByLevel = mapToSortedStats(levelMap)
	ov.AlertByStatus = mapToSortedStats(statusMap)

	// 通行结果统计
	for _, l := range logs {
		if l.Result == model.AccessGranted {
			ov.GrantedCount++
		} else {
			ov.DeniedCount++
		}
	}

	return ov, nil
}

// ZoneTrafficStats 统计各区域通行量，按总通行量倒序。
func (s *Service) ZoneTrafficStats() ([]ZoneTraffic, error) {
	zones := s.store.ListZones()
	points := s.store.ListAccessPoints()
	logs := s.store.ListAccessLogs()

	pointZone := make(map[string]string, len(points))
	for _, p := range points {
		pointZone[p.ID] = p.ZoneID
	}

	result := make([]ZoneTraffic, 0, len(zones))
	byZone := make(map[string]*ZoneTraffic, len(zones))
	for _, z := range zones {
		zt := &ZoneTraffic{ZoneID: z.ID, ZoneName: z.Name}
		byZone[z.ID] = zt
		result = append(result, *zt)
	}

	for _, l := range logs {
		zoneID, ok := pointZone[l.AccessPointID]
		if !ok {
			continue
		}
		zt, ok := byZone[zoneID]
		if !ok {
			continue
		}
		zt.Total++
		if l.Result == model.AccessGranted {
			zt.Granted++
		} else {
			zt.Denied++
		}
	}

	out := make([]ZoneTraffic, 0, len(byZone))
	for _, zt := range byZone {
		out = append(out, *zt)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Total > out[j].Total
	})
	return out, nil
}

// PointTrafficTopN 返回门禁点通行量 TOP N。
func (s *Service) PointTrafficTopN(n int) ([]PointTraffic, error) {
	if n <= 0 {
		n = 5
	}
	points := s.store.ListAccessPoints()
	logs := s.store.ListAccessLogs()

	nameMap := make(map[string]string, len(points))
	countMap := make(map[string]int, len(points))
	for _, p := range points {
		nameMap[p.ID] = p.Name
	}
	for _, l := range logs {
		countMap[l.AccessPointID]++
	}

	out := make([]PointTraffic, 0, len(countMap))
	for apID, count := range countMap {
		out = append(out, PointTraffic{
			AccessPointID:   apID,
			AccessPointName: nameMap[apID],
			Total:           count,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Total > out[j].Total
	})
	if len(out) > n {
		out = out[:n]
	}
	return out, nil
}

// CredentialStatusStats 返回凭证状态分布。
func (s *Service) CredentialStatusStats() ([]CredentialStatusStat, error) {
	creds := s.store.ListCredentials()
	m := make(map[string]int)
	for _, c := range creds {
		m[c.Status]++
	}
	out := make([]CredentialStatusStat, 0, len(m))
	for status, count := range m {
		out = append(out, CredentialStatusStat{Status: status, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Status < out[j].Status
	})
	return out, nil
}

// AlertLevelStats 返回告警按等级统计。
func (s *Service) AlertLevelStats() ([]AlertStat, error) {
	alerts := s.store.ListAlerts()
	m := make(map[string]int)
	for _, a := range alerts {
		m[a.Level]++
	}
	return mapToSortedStats(m), nil
}

// AlertStatusStats 返回告警按状态统计。
func (s *Service) AlertStatusStats() ([]AlertStat, error) {
	alerts := s.store.ListAlerts()
	m := make(map[string]int)
	for _, a := range alerts {
		m[a.Status]++
	}
	return mapToSortedStats(m), nil
}

func mapToSortedStats(m map[string]int) []AlertStat {
	out := make([]AlertStat, 0, len(m))
	for k, v := range m {
		out = append(out, AlertStat{Key: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}
