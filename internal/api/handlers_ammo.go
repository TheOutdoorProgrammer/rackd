package api

import (
	"errors"
	"net/http"

	"github.com/TheOutdoorProgrammer/boating-accident/internal/db"
)

func (s *Server) handleListAmmo(w http.ResponseWriter, _ *http.Request) {
	items, err := s.store.ListAmmo()
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateAmmo(w http.ResponseWriter, r *http.Request) {
	var a db.Ammo
	if err := decodeJSON(r, &a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := s.store.CreateAmmo(&a); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (s *Server) handleGetAmmo(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	a, err := s.store.GetAmmo(id)
	if errors.Is(err, db.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (s *Server) handleUpdateAmmo(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	existing, err := s.store.GetAmmo(id)
	if errors.Is(err, db.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	var a db.Ammo
	if err := decodeJSON(r, &a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	a.ID = id
	a.CreatedAt = existing.CreatedAt
	if err := s.store.UpdateAmmo(&a); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (s *Server) handleDeleteAmmo(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	s.deletePhotosForOwner("ammo", id)
	switch err := s.store.DeleteAmmo(id); {
	case errors.Is(err, db.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case err != nil:
		serverError(w, err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleAdjustAmmo applies a signed delta to rounds-on-hand (use/refill buttons).
func (s *Server) handleAdjustAmmo(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Delta int64 `json:"delta"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	a, err := s.store.AdjustAmmo(id, body.Delta)
	if errors.Is(err, db.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

// handleListAmmoFirearms lists the firearms an ammo line is linked to.
func (s *Server) handleListAmmoFirearms(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	items, err := s.store.ListFirearmsForAmmo(id)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}
