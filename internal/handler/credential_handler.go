package handler

import (
	"net/http"
	"time"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/httpx"
)

func (s *Server) registerCredentialRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/credentials", s.createCredential)
	mux.HandleFunc("GET /api/credentials", s.listCredentials)
	mux.HandleFunc("GET /api/credentials/{id}", s.getCredential)
	mux.HandleFunc("PUT /api/credentials/{id}", s.updateCredential)
	mux.HandleFunc("DELETE /api/credentials/{id}", s.deleteCredential)
	mux.HandleFunc("POST /api/credentials/{id}/transition", s.transitionCredential)
	mux.HandleFunc("POST /api/credentials/{id}/check", s.checkCredential)
	mux.HandleFunc("POST /api/credentials/batch-issue", s.batchIssueCredentials)
	mux.HandleFunc("POST /api/credentials/batch-disable", s.batchDisableCredentials)
}

type credentialRequest struct {
	PersonID   string    `json:"person_id"`
	Type       string    `json:"type"`
	Code       string    `json:"code"`
	ValidFrom  time.Time `json:"valid_from"`
	ValidUntil time.Time `json:"valid_until"`
	Status     string    `json:"status"`
}

func (s *Server) createCredential(w http.ResponseWriter, r *http.Request) {
	var req credentialRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	c, err := s.svc.CreateCredential(model.Credential{
		PersonID:   req.PersonID,
		Type:       req.Type,
		Code:       req.Code,
		ValidFrom:  req.ValidFrom,
		ValidUntil: req.ValidUntil,
		Status:     req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, c)
}

func (s *Server) listCredentials(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.CredentialFilter{
		PersonID: r.URL.Query().Get("person_id"),
		Type:     r.URL.Query().Get("type"),
		Status:   r.URL.Query().Get("status"),
		Keyword:  r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListCredentials(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getCredential(w http.ResponseWriter, r *http.Request) {
	c, err := s.svc.GetCredential(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

func (s *Server) updateCredential(w http.ResponseWriter, r *http.Request) {
	var req credentialRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	c, err := s.svc.UpdateCredential(r.PathValue("id"), model.Credential{
		PersonID:   req.PersonID,
		Type:       req.Type,
		Code:       req.Code,
		ValidFrom:  req.ValidFrom,
		ValidUntil: req.ValidUntil,
		Status:     req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

func (s *Server) deleteCredential(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteCredential(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) transitionCredential(w http.ResponseWriter, r *http.Request) {
	var req transitionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	c, err := s.svc.TransitionCredential(r.PathValue("id"), req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

func (s *Server) checkCredential(w http.ResponseWriter, r *http.Request) {
	c, err := s.svc.CheckValidity(r.PathValue("id"), time.Now())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{
		"id":      c.ID,
		"code":    c.Code,
		"status":  c.Status,
		"valid":   true,
	})
}

type batchIssueRequest struct {
	Credentials []credentialRequest `json:"credentials"`
}

func (s *Server) batchIssueCredentials(w http.ResponseWriter, r *http.Request) {
	var req batchIssueRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	inputs := make([]model.Credential, 0, len(req.Credentials))
	for _, c := range req.Credentials {
		inputs = append(inputs, model.Credential{
			PersonID:   c.PersonID,
			Type:       c.Type,
			Code:       c.Code,
			ValidFrom:  c.ValidFrom,
			ValidUntil: c.ValidUntil,
			Status:     c.Status,
		})
	}
	created, failures, err := s.svc.BatchIssueCredentials(inputs)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{
		"created":  created,
		"failures": failures,
	})
}

type batchDisableRequest struct {
	IDs []string `json:"ids"`
}

func (s *Server) batchDisableCredentials(w http.ResponseWriter, r *http.Request) {
	var req batchDisableRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	updated, failures, err := s.svc.BatchDisableCredentials(req.IDs)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{
		"updated":  updated,
		"failures": failures,
	})
}
