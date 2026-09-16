package handler

import (
	"net/http"
	"time"

	"accesscontrol/internal/service"
	"accesscontrol/pkg/httpx"
)

func (s *Server) registerAccessRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/access", s.access)
}

type accessRequest struct {
	CredentialID   string    `json:"credential_id"`
	CredentialCode string    `json:"credential_code"`
	AccessPointID  string    `json:"access_point_id"`
	Direction      string    `json:"direction"`
	AccessAt       time.Time `json:"access_at"`
}

func (s *Server) access(w http.ResponseWriter, r *http.Request) {
	var req accessRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	result, err := s.svc.AuthorizeAccess(service.AccessRequest{
		CredentialID:   req.CredentialID,
		CredentialCode: req.CredentialCode,
		AccessPointID:  req.AccessPointID,
		Direction:      req.Direction,
		AccessAt:       req.AccessAt,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, result)
}
