package api

import "net/http"

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.pool.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error", "database": "down"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "up"})
}
