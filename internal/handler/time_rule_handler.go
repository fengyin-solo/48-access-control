package handler

import (
	"net/http"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/httpx"
)

func (s *Server) registerTimeRuleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/time-rules", s.createTimeRule)
	mux.HandleFunc("GET /api/time-rules", s.listTimeRules)
	mux.HandleFunc("GET /api/time-rules/{id}", s.getTimeRule)
	mux.HandleFunc("PUT /api/time-rules/{id}", s.updateTimeRule)
	mux.HandleFunc("DELETE /api/time-rules/{id}", s.deleteTimeRule)
}

type timeRuleRequest struct {
	AccessPointID string `json:"access_point_id"`
	Name          string `json:"name"`
	Weekdays      []int  `json:"weekdays"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Effect        string `json:"effect"`
	Status        string `json:"status"`
}

func (s *Server) createTimeRule(w http.ResponseWriter, r *http.Request) {
	var req timeRuleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	t, err := s.svc.CreateTimeRule(model.TimeRule{
		AccessPointID: req.AccessPointID,
		Name:          req.Name,
		Weekdays:      req.Weekdays,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Effect:        req.Effect,
		Status:        req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, t)
}

func (s *Server) listTimeRules(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.TimeRuleFilter{
		AccessPointID: r.URL.Query().Get("access_point_id"),
		Effect:        r.URL.Query().Get("effect"),
		Status:        r.URL.Query().Get("status"),
		Keyword:       r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListTimeRules(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getTimeRule(w http.ResponseWriter, r *http.Request) {
	t, err := s.svc.GetTimeRule(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, t)
}

func (s *Server) updateTimeRule(w http.ResponseWriter, r *http.Request) {
	var req timeRuleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	t, err := s.svc.UpdateTimeRule(r.PathValue("id"), model.TimeRule{
		AccessPointID: req.AccessPointID,
		Name:          req.Name,
		Weekdays:      req.Weekdays,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Effect:        req.Effect,
		Status:        req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, t)
}

func (s *Server) deleteTimeRule(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteTimeRule(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
