package service

import (
	"sort"

	"accesscontrol/internal/model"
)

// KVStat 表示通用的键值统计项。
type KVStat struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

// DirectionStat 表示通行方向 × 结果的交叉统计。
type DirectionStat struct {
	Direction string `json:"direction"`
	Granted   int    `json:"granted"`
	Denied    int    `json:"denied"`
	Total     int    `json:"total"`
}

// HourlyStat 表示按小时的通行量统计。
type HourlyStat struct {
	Hour  int `json:"hour"`
	Total int `json:"total"`
}

// PointTypeStats 返回门禁点按类型分布。
func (s *Service) PointTypeStats() ([]KVStat, error) {
	points := s.store.ListAccessPoints()
	m := make(map[string]int)
	for _, p := range points {
		m[p.Type]++
	}
	return mapToKV(m), nil
}

// PointStatusStats 返回门禁点按状态分布。
func (s *Service) PointStatusStats() ([]KVStat, error) {
	points := s.store.ListAccessPoints()
	m := make(map[string]int)
	for _, p := range points {
		m[p.Status]++
	}
	return mapToKV(m), nil
}

// ReaderStatusStats 返回读卡器按状态分布。
func (s *Service) ReaderStatusStats() ([]KVStat, error) {
	readers := s.store.ListReaders()
	m := make(map[string]int)
	for _, r := range readers {
		m[r.Status]++
	}
	return mapToKV(m), nil
}

// DirectionStats 返回通行记录按方向 × 结果的交叉统计。
func (s *Service) DirectionStats() ([]DirectionStat, error) {
	logs := s.store.ListAccessLogs()
	byDir := make(map[string]*DirectionStat)
	for _, l := range logs {
		ds, ok := byDir[l.Direction]
		if !ok {
			ds = &DirectionStat{Direction: l.Direction}
			byDir[l.Direction] = ds
		}
		ds.Total++
		if l.Result == model.AccessGranted {
			ds.Granted++
		} else {
			ds.Denied++
		}
	}
	out := make([]DirectionStat, 0, len(byDir))
	for _, ds := range byDir {
		out = append(out, *ds)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Direction < out[j].Direction
	})
	return out, nil
}

// HourlyTrafficStats 返回按小时（0-23）的通行量分布。
func (s *Service) HourlyTrafficStats() ([]HourlyStat, error) {
	logs := s.store.ListAccessLogs()
	counts := make([]int, 24)
	for _, l := range logs {
		h := l.AccessAt.Hour()
		if h >= 0 && h < 24 {
			counts[h]++
		}
	}
	out := make([]HourlyStat, 0, 24)
	for h, c := range counts {
		out = append(out, HourlyStat{Hour: h, Total: c})
	}
	return out, nil
}

// TimeRuleEffectStats 返回时段规则按效果分布。
func (s *Service) TimeRuleEffectStats() ([]KVStat, error) {
	rules := s.store.ListTimeRules()
	m := make(map[string]int)
	for _, r := range rules {
		m[r.Effect]++
	}
	return mapToKV(m), nil
}

// ScheduleStatusStats 返回排班按状态分布。
func (s *Service) ScheduleStatusStats() ([]KVStat, error) {
	schedules := s.store.ListSchedules()
	m := make(map[string]int)
	for _, sc := range schedules {
		m[sc.Status]++
	}
	return mapToKV(m), nil
}

// PersonCredentialStats 返回每个人员持有的凭证数量，按数量倒序。
func (s *Service) PersonCredentialStats() ([]KVStat, error) {
	creds := s.store.ListCredentials()
	persons := s.store.ListPersons()
	nameMap := make(map[string]string, len(persons))
	for _, p := range persons {
		nameMap[p.ID] = p.Name
	}
	m := make(map[string]int)
	for _, c := range creds {
		name := nameMap[c.PersonID]
		if name == "" {
			name = c.PersonID
		}
		m[name]++
	}
	out := mapToKV(m)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Count > out[j].Count
	})
	return out, nil
}

func mapToKV(m map[string]int) []KVStat {
	out := make([]KVStat, 0, len(m))
	for k, v := range m {
		out = append(out, KVStat{Key: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}
