package handler

import (
	"net/http"
	"time"

	"accesscontrol/internal/model"
	"accesscontrol/pkg/httpx"
)

func (s *Server) registerAccessLogRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/access-logs", s.createAccessLog)
	mux.HandleFunc("GET /api/access-logs", s.listAccessLogs)
	mux.HandleFunc("GET /api/access-logs/{id}", s.getAccessLog)
	mux.HandleFunc("DELETE /api/access-logs/{id}", s.deleteAccessLog)
	mux.HandleFunc("POST /api/access-logs/batch-delete", s.batchDeleteAccessLogs)
}

type accessLogRequest struct {
	CredentialID  string    `json:"credential_id"`
	AccessPointID string    `json:"access_point_id"`
	AccessAt      time.Time `json:"access_at"`
	Direction     string    `json:"direction"`
	Result        string    `json:"result"`
	Reason        string    `json:"reason"`
}

func (s *Server) createAccessLog(w http.ResponseWriter, r *http.Request) {
	var req accessLogRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	l, err := s.svc.CreateAccessLog(model.AccessLog{
		CredentialID:  req.CredentialID,
		AccessPointID: req.AccessPointID,
		AccessAt:      req.AccessAt,
		Direction:     req.Direction,
		Result:        req.Result,
		Reason:        req.Reason,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, l)
}

func (s *Server) listAccessLogs(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.AccessLogFilter{
		CredentialID:  r.URL.Query().Get("credential_id"),
		AccessPointID: r.URL.Query().Get("access_point_id"),
		Result:        r.URL.Query().Get("result"),
		Direction:     r.URL.Query().Get("direction"),
	}
	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.From = t
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.To = t
		}
	}
	items, total, err := s.svc.ListAccessLogs(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getAccessLog(w http.ResponseWriter, r *http.Request) {
	l, err := s.svc.GetAccessLog(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, l)
}

func (s *Server) deleteAccessLog(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteAccessLog(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type batchDeleteRequest struct {
	IDs []string `json:"ids"`
}

func (s *Server) batchDeleteAccessLogs(w http.ResponseWriter, r *http.Request) {
	var req batchDeleteRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	n, err := s.svc.BatchDeleteAccessLogs(req.IDs)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]int{"deleted": n})
}
