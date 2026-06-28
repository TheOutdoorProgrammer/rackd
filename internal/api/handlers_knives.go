package api

import (
	"errors"
	"net/http"

	"github.com/TheOutdoorProgrammer/boating-accident/internal/db"
)

func (s *Server) handleListKnives(w http.ResponseWriter, _ *http.Request) {
	items, err := s.store.ListKnives()
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateKnife(w http.ResponseWriter, r *http.Request) {
	var k db.Knife
	if err := decodeJSON(r, &k); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := s.store.CreateKnife(&k); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, k)
}

func (s *Server) handleGetKnife(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	k, err := s.store.GetKnife(id)
	if errors.Is(err, db.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, k)
}

func (s *Server) handleUpdateKnife(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	existing, err := s.store.GetKnife(id)
	if errors.Is(err, db.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	var k db.Knife
	if err := decodeJSON(r, &k); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	k.ID = id
	k.CreatedAt = existing.CreatedAt
	if err := s.store.UpdateKnife(&k); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, k)
}

func (s *Server) handleDeleteKnife(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	s.deletePhotosForOwner("knives", id)
	switch err := s.store.DeleteKnife(id); {
	case errors.Is(err, db.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case err != nil:
		serverError(w, err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
