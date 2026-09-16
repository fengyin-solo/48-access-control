package handler

import (
	"net/http"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/httpx"
)

func (s *Server) registerPersonRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/persons", s.createPerson)
	mux.HandleFunc("GET /api/persons", s.listPersons)
	mux.HandleFunc("GET /api/persons/{id}", s.getPerson)
	mux.HandleFunc("PUT /api/persons/{id}", s.updatePerson)
	mux.HandleFunc("DELETE /api/persons/{id}", s.deletePerson)
}

type personRequest struct {
	Name       string `json:"name"`
	Department string `json:"department"`
	Phone      string `json:"phone"`
	Status     string `json:"status"`
}

func (s *Server) createPerson(w http.ResponseWriter, r *http.Request) {
	var req personRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.CreatePerson(model.Person{
		Name:       req.Name,
		Department: req.Department,
		Phone:      req.Phone,
		Status:     req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, p)
}

func (s *Server) listPersons(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.PersonFilter{
		Department: r.URL.Query().Get("department"),
		Status:     r.URL.Query().Get("status"),
		Keyword:    r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListPersons(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getPerson(w http.ResponseWriter, r *http.Request) {
	p, err := s.svc.GetPerson(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

func (s *Server) updatePerson(w http.ResponseWriter, r *http.Request) {
	var req personRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.UpdatePerson(r.PathValue("id"), model.Person{
		Name:       req.Name,
		Department: req.Department,
		Phone:      req.Phone,
		Status:     req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

func (s *Server) deletePerson(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeletePerson(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
