package api

import (
	"errors"
	"net/http"

	"github.com/TheOutdoorProgrammer/boating-accident/internal/db"
)

func (s *Server) handleListFirearms(w http.ResponseWriter, _ *http.Request) {
	items, err := s.store.ListFirearms()
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateFirearm(w http.ResponseWriter, r *http.Request) {
	var f db.Firearm
	if err := decodeJSON(r, &f); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := s.store.CreateFirearm(&f); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

func (s *Server) handleGetFirearm(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	f, err := s.store.GetFirearm(id)
	if errors.Is(err, db.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (s *Server) handleUpdateFirearm(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	existing, err := s.store.GetFirearm(id)
	if errors.Is(err, db.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	var f db.Firearm
	if err := decodeJSON(r, &f); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	f.ID = id
	f.CreatedAt = existing.CreatedAt
	if err := s.store.UpdateFirearm(&f); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (s *Server) handleDeleteFirearm(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	s.deletePhotosForOwner("firearms", id)
	switch err := s.store.DeleteFirearm(id); {
	case errors.Is(err, db.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case err != nil:
		serverError(w, err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

// --- ammo links (routed under /firearms/{id}/ammo) ---

func (s *Server) handleListFirearmAmmo(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	links, err := s.store.ListAmmoForFirearm(id)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, links)
}

func (s *Server) handleLinkAmmo(w http.ResponseWriter, r *http.Request) {
	fid, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	aid, err := parseID(r, "ammoID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ammo id")
		return
	}
	var body struct {
		Note string `json:"note"`
	}
	_ = decodeJSON(r, &body) // note is optional
	if err := s.store.LinkAmmo(fid, aid, body.Note); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleUnlinkAmmo(w http.ResponseWriter, r *http.Request) {
	fid, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	aid, err := parseID(r, "ammoID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ammo id")
		return
	}
	switch err := s.store.UnlinkAmmo(fid, aid); {
	case errors.Is(err, db.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case err != nil:
		serverError(w, err)
	default:
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

// --- accessory mounts (routed under /firearms/{id}/accessories) ---

func (s *Server) handleListFirearmAccessories(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	items, err := s.store.ListAccessoriesForFirearm(id)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleLinkAccessory(w http.ResponseWriter, r *http.Request) {
	fid, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	aid, err := parseID(r, "accID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid accessory id")
		return
	}
	switch err := s.store.LinkAccessory(fid, aid); {
	case errors.Is(err, db.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, db.ErrAtCapacity):
		writeError(w, http.StatusConflict, "accessory quantity is fully assigned")
	case err != nil:
		serverError(w, err)
	default:
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func (s *Server) handleUnlinkAccessory(w http.ResponseWriter, r *http.Request) {
	fid, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	aid, err := parseID(r, "accID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid accessory id")
		return
	}
	switch err := s.store.UnlinkAccessory(fid, aid); {
	case errors.Is(err, db.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case err != nil:
		serverError(w, err)
	default:
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}
