package handler

import (
	"net/http"

	"accesscontrol/pkg/httpx"
)

func (s *Server) registerExportRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/export", s.export)
}

func (s *Server) export(w http.ResponseWriter, r *http.Request) {
	snap, err := s.svc.BuildExportSnapshot()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, snap)
}
