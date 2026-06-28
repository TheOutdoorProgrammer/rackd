package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/TheOutdoorProgrammer/rackd/internal/db"
)

func (s *Server) handleListAccessories(w http.ResponseWriter, r *http.Request) {
	var firearmID *int64
	if q := r.URL.Query().Get("firearmId"); q != "" {
		if v, err := strconv.ParseInt(q, 10, 64); err == nil {
			firearmID = &v
		}
	}
	items, err := s.store.ListAccessories(firearmID)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateAccessory(w http.ResponseWriter, r *http.Request) {
	var a db.Accessory
	if err := decodeJSON(r, &a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := s.store.CreateAccessory(&a); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (s *Server) handleGetAccessory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	a, err := s.store.GetAccessory(id)
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

func (s *Server) handleUpdateAccessory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	existing, err := s.store.GetAccessory(id)
	if errors.Is(err, db.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	var a db.Accessory
	if err := decodeJSON(r, &a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	a.ID = id
	a.CreatedAt = existing.CreatedAt
	if err := s.store.UpdateAccessory(&a); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (s *Server) handleDeleteAccessory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	s.deletePhotosForOwner("accessories", id)
	switch err := s.store.DeleteAccessory(id); {
	case errors.Is(err, db.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case err != nil:
		serverError(w, err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
