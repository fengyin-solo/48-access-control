package handler

import (
	"net/http"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/httpx"
)

func (s *Server) registerReaderRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/readers", s.createReader)
	mux.HandleFunc("GET /api/readers", s.listReaders)
	mux.HandleFunc("GET /api/readers/{id}", s.getReader)
	mux.HandleFunc("PUT /api/readers/{id}", s.updateReader)
	mux.HandleFunc("DELETE /api/readers/{id}", s.deleteReader)
}

type readerRequest struct {
	Name          string `json:"name"`
	AccessPointID string `json:"access_point_id"`
	Model         string `json:"model"`
	Status        string `json:"status"`
}

func (s *Server) createReader(w http.ResponseWriter, r *http.Request) {
	var req readerRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	rd, err := s.svc.CreateReader(model.Reader{
		Name:          req.Name,
		AccessPointID: req.AccessPointID,
		Model:         req.Model,
		Status:        req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, rd)
}

func (s *Server) listReaders(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ReaderFilter{
		AccessPointID: r.URL.Query().Get("access_point_id"),
		Status:        r.URL.Query().Get("status"),
		Keyword:       r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListReaders(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getReader(w http.ResponseWriter, r *http.Request) {
	rd, err := s.svc.GetReader(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, rd)
}

func (s *Server) updateReader(w http.ResponseWriter, r *http.Request) {
	var req readerRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	rd, err := s.svc.UpdateReader(r.PathValue("id"), model.Reader{
		Name:          req.Name,
		AccessPointID: req.AccessPointID,
		Model:         req.Model,
		Status:        req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, rd)
}

func (s *Server) deleteReader(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteReader(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
