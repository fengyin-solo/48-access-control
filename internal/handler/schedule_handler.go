package handler

import (
	"net/http"
	"time"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/httpx"
)

func (s *Server) registerScheduleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/schedules", s.createSchedule)
	mux.HandleFunc("GET /api/schedules", s.listSchedules)
	mux.HandleFunc("GET /api/schedules/{id}", s.getSchedule)
	mux.HandleFunc("PUT /api/schedules/{id}", s.updateSchedule)
	mux.HandleFunc("DELETE /api/schedules/{id}", s.deleteSchedule)
	mux.HandleFunc("POST /api/schedules/{id}/transition", s.transitionSchedule)
}

type scheduleRequest struct {
	PersonID string    `json:"person_id"`
	ZoneID   string    `json:"zone_id"`
	StartAt  time.Time `json:"start_at"`
	EndAt    time.Time `json:"end_at"`
	Status   string    `json:"status"`
}

func (s *Server) createSchedule(w http.ResponseWriter, r *http.Request) {
	var req scheduleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	sc, err := s.svc.CreateSchedule(model.Schedule{
		PersonID: req.PersonID,
		ZoneID:   req.ZoneID,
		StartAt:  req.StartAt,
		EndAt:    req.EndAt,
		Status:   req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, sc)
}

func (s *Server) listSchedules(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ScheduleFilter{
		PersonID: r.URL.Query().Get("person_id"),
		ZoneID:   r.URL.Query().Get("zone_id"),
		Status:   r.URL.Query().Get("status"),
	}
	items, total, err := s.svc.ListSchedules(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getSchedule(w http.ResponseWriter, r *http.Request) {
	sc, err := s.svc.GetSchedule(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, sc)
}

func (s *Server) updateSchedule(w http.ResponseWriter, r *http.Request) {
	var req scheduleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	sc, err := s.svc.UpdateSchedule(r.PathValue("id"), model.Schedule{
		PersonID: req.PersonID,
		ZoneID:   req.ZoneID,
		StartAt:  req.StartAt,
		EndAt:    req.EndAt,
		Status:   req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, sc)
}

func (s *Server) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteSchedule(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) transitionSchedule(w http.ResponseWriter, r *http.Request) {
	var req transitionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	sc, err := s.svc.TransitionSchedule(r.PathValue("id"), req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, sc)
}
