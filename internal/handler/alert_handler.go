package handler

import (
	"net/http"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/httpx"
)

func (s *Server) registerAlertRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/alerts", s.createAlert)
	mux.HandleFunc("GET /api/alerts", s.listAlerts)
	mux.HandleFunc("GET /api/alerts/{id}", s.getAlert)
	mux.HandleFunc("PUT /api/alerts/{id}", s.updateAlert)
	mux.HandleFunc("DELETE /api/alerts/{id}", s.deleteAlert)
	mux.HandleFunc("POST /api/alerts/{id}/transition", s.transitionAlert)
	mux.HandleFunc("POST /api/alerts/batch-update", s.batchUpdateAlerts)
}

type alertRequest struct {
	AccessLogID string `json:"access_log_id"`
	Type        string `json:"type"`
	Level       string `json:"level"`
	Status      string `json:"status"`
	Handler     string `json:"handler"`
}

type alertTransitionRequest struct {
	Status  string `json:"status"`
	Handler string `json:"handler"`
}

func (s *Server) createAlert(w http.ResponseWriter, r *http.Request) {
	var req alertRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.CreateAlert(model.Alert{
		AccessLogID: req.AccessLogID,
		Type:        req.Type,
		Level:       req.Level,
		Status:      req.Status,
		Handler:     req.Handler,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, a)
}

func (s *Server) listAlerts(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.AlertFilter{
		Type:    r.URL.Query().Get("type"),
		Level:   r.URL.Query().Get("level"),
		Status:  r.URL.Query().Get("status"),
		Handler: r.URL.Query().Get("handler"),
		Keyword: r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListAlerts(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getAlert(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetAlert(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) updateAlert(w http.ResponseWriter, r *http.Request) {
	var req alertRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.UpdateAlert(r.PathValue("id"), model.Alert{
		Type:    req.Type,
		Level:   req.Level,
		Status:  req.Status,
		Handler: req.Handler,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) deleteAlert(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteAlert(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) transitionAlert(w http.ResponseWriter, r *http.Request) {
	var req alertTransitionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.TransitionAlert(r.PathValue("id"), req.Status, req.Handler)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

type batchUpdateAlertRequest struct {
	IDs     []string `json:"ids"`
	Status  string   `json:"status"`
	Handler string   `json:"handler"`
}

func (s *Server) batchUpdateAlerts(w http.ResponseWriter, r *http.Request) {
	var req batchUpdateAlertRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	updated, failures, err := s.svc.BatchUpdateAlertStatus(req.IDs, req.Status, req.Handler)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{
		"updated":  updated,
		"failures": failures,
	})
}
