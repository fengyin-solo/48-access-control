package handler

import (
	"net/http"
	"strconv"

	"accesscontrol/pkg/httpx"
)

func (s *Server) registerStatsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/stats/overview", s.statsOverview)
	mux.HandleFunc("GET /api/stats/zone-traffic", s.zoneTraffic)
	mux.HandleFunc("GET /api/stats/point-traffic/top", s.pointTrafficTop)
	mux.HandleFunc("GET /api/stats/credential-status", s.credentialStatus)
	mux.HandleFunc("GET /api/stats/alert-level", s.alertLevel)
	mux.HandleFunc("GET /api/stats/alert-status", s.alertStatus)
	mux.HandleFunc("GET /api/stats/point-type", s.pointType)
	mux.HandleFunc("GET /api/stats/point-status", s.pointStatus)
	mux.HandleFunc("GET /api/stats/reader-status", s.readerStatus)
	mux.HandleFunc("GET /api/stats/direction", s.directionStats)
	mux.HandleFunc("GET /api/stats/hourly", s.hourlyStats)
	mux.HandleFunc("GET /api/stats/time-rule-effect", s.timeRuleEffect)
	mux.HandleFunc("GET /api/stats/schedule-status", s.scheduleStatus)
	mux.HandleFunc("GET /api/stats/person-credential", s.personCredential)
}

func (s *Server) statsOverview(w http.ResponseWriter, r *http.Request) {
	ov, err := s.svc.StatsOverview()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, ov)
}

func (s *Server) zoneTraffic(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.ZoneTrafficStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) pointTrafficTop(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	if n <= 0 {
		n = 5
	}
	items, err := s.svc.PointTrafficTopN(n)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) credentialStatus(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.CredentialStatusStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) alertLevel(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.AlertLevelStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) alertStatus(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.AlertStatusStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) pointType(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.PointTypeStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) pointStatus(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.PointStatusStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) readerStatus(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.ReaderStatusStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) directionStats(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.DirectionStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) hourlyStats(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.HourlyTrafficStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) timeRuleEffect(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.TimeRuleEffectStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) scheduleStatus(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.ScheduleStatusStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) personCredential(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.PersonCredentialStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}
