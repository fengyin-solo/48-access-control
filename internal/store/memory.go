package store

import (
	"sync"

	"accesscontrol/internal/model"
)

// MemoryStore 以内存 map 保存所有实体，用读写锁保证并发安全。
type MemoryStore struct {
	mu           sync.RWMutex
	zones        map[string]*model.Zone
	accessPoints map[string]*model.AccessPoint
	readers      map[string]*model.Reader
	persons      map[string]*model.Person
	credentials  map[string]*model.Credential
	accessLogs   map[string]*model.AccessLog
	timeRules    map[string]*model.TimeRule
	schedules    map[string]*model.Schedule
	alerts       map[string]*model.Alert
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		zones:        make(map[string]*model.Zone),
		accessPoints: make(map[string]*model.AccessPoint),
		readers:      make(map[string]*model.Reader),
		persons:      make(map[string]*model.Person),
		credentials:  make(map[string]*model.Credential),
		accessLogs:   make(map[string]*model.AccessLog),
		timeRules:    make(map[string]*model.TimeRule),
		schedules:    make(map[string]*model.Schedule),
		alerts:       make(map[string]*model.Alert),
	}
}

var _ Store = (*MemoryStore)(nil)
