package handler

import (
	"net/http"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/httpx"
)

func (s *Server) registerAccessPointRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/access-points", s.createAccessPoint)
	mux.HandleFunc("GET /api/access-points", s.listAccessPoints)
	mux.HandleFunc("GET /api/access-points/{id}", s.getAccessPoint)
	mux.HandleFunc("PUT /api/access-points/{id}", s.updateAccessPoint)
	mux.HandleFunc("DELETE /api/access-points/{id}", s.deleteAccessPoint)
	mux.HandleFunc("POST /api/access-points/{id}/transition", s.transitionAccessPoint)
}

type accessPointRequest struct {
	Name      string `json:"name"`
	ZoneID    string `json:"zone_id"`
	Type      string `json:"type"`
	Direction string `json:"direction"`
	Status    string `json:"status"`
}

type transitionRequest struct {
	Status string `json:"status"`
}

func (s *Server) createAccessPoint(w http.ResponseWriter, r *http.Request) {
	var req accessPointRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.CreateAccessPoint(model.AccessPoint{
		Name:      req.Name,
		ZoneID:    req.ZoneID,
		Type:      req.Type,
		Direction: req.Direction,
		Status:    req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, a)
}

func (s *Server) listAccessPoints(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.AccessPointFilter{
		ZoneID:  r.URL.Query().Get("zone_id"),
		Type:    r.URL.Query().Get("type"),
		Status:  r.URL.Query().Get("status"),
		Keyword: r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListAccessPoints(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getAccessPoint(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetAccessPoint(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) updateAccessPoint(w http.ResponseWriter, r *http.Request) {
	var req accessPointRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.UpdateAccessPoint(r.PathValue("id"), model.AccessPoint{
		Name:      req.Name,
		ZoneID:    req.ZoneID,
		Type:      req.Type,
		Direction: req.Direction,
		Status:    req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) deleteAccessPoint(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteAccessPoint(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) transitionAccessPoint(w http.ResponseWriter, r *http.Request) {
	var req transitionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.TransitionAccessPoint(r.PathValue("id"), req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}
