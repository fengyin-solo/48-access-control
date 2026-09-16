package handler

import (
	"net/http"
	"strconv"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/httpx"
)

func (s *Server) registerZoneRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/zones", s.createZone)
	mux.HandleFunc("GET /api/zones", s.listZones)
	mux.HandleFunc("GET /api/zones/{id}", s.getZone)
	mux.HandleFunc("PUT /api/zones/{id}", s.updateZone)
	mux.HandleFunc("DELETE /api/zones/{id}", s.deleteZone)
}

type zoneRequest struct {
	Name     string `json:"name"`
	Level    int    `json:"level"`
	ParentID string `json:"parent_id"`
	Status   string `json:"status"`
}

func (s *Server) createZone(w http.ResponseWriter, r *http.Request) {
	var req zoneRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	z, err := s.svc.CreateZone(model.Zone{
		Name:     req.Name,
		Level:    req.Level,
		ParentID: req.ParentID,
		Status:   req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, z)
}

func (s *Server) listZones(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	level, _ := strconv.Atoi(r.URL.Query().Get("level"))
	filter := model.ZoneFilter{
		Name:   r.URL.Query().Get("name"),
		Status: r.URL.Query().Get("status"),
		Level:  level,
		Parent: r.URL.Query().Get("parent_id"),
	}
	items, total, err := s.svc.ListZones(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getZone(w http.ResponseWriter, r *http.Request) {
	z, err := s.svc.GetZone(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, z)
}

func (s *Server) updateZone(w http.ResponseWriter, r *http.Request) {
	var req zoneRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	z, err := s.svc.UpdateZone(r.PathValue("id"), model.Zone{
		Name:     req.Name,
		Level:    req.Level,
		ParentID: req.ParentID,
		Status:   req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, z)
}

func (s *Server) deleteZone(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteZone(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
