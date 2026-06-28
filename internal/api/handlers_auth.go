package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/TheOutdoorProgrammer/rackd/internal/vault"
)

type statusResponse struct {
	Initialized bool `json:"initialized"`
	Unlocked    bool `json:"unlocked"`
}

type pinRequest struct {
	PIN string `json:"pin"`
}

// handleStatus reports whether a vault exists and whether this session has it
// unlocked. It is unauthenticated so the SPA can decide which screen to show.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	initialized, err := s.vault.Initialized()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "status check failed")
		return
	}
	authed := s.sessions.GetBool(r.Context(), sessionAuthKey)
	writeJSON(w, http.StatusOK, statusResponse{
		Initialized: initialized,
		Unlocked:    authed && s.vault.Unlocked(),
	})
}

// handleSetup creates the vault on first run and authenticates the session.
func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	var req pinRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if !validPIN(req.PIN) {
		writeError(w, http.StatusBadRequest, "pin must be exactly 6 digits")
		return
	}
	switch err := s.vault.Setup(req.PIN); {
	case errors.Is(err, vault.ErrAlreadySetup):
		writeError(w, http.StatusConflict, "already initialized")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "setup failed")
		return
	}
	s.grantSession(r)
	writeJSON(w, http.StatusOK, statusResponse{Initialized: true, Unlocked: true})
}

// handleUnlock verifies the PIN, unlocks the vault, and authenticates the
// session. Repeated failures are throttled by the lockout.
func (s *Server) handleUnlock(w http.ResponseWriter, r *http.Request) {
	if wait := s.lockout.Retry(); wait > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
		writeError(w, http.StatusTooManyRequests, "too many attempts, try again later")
		return
	}
	var req pinRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	switch err := s.vault.Unlock(req.PIN); {
	case errors.Is(err, vault.ErrNotInitialized):
		writeError(w, http.StatusConflict, "not initialized")
		return
	case errors.Is(err, vault.ErrBadPIN):
		s.lockout.Fail()
		writeError(w, http.StatusUnauthorized, "incorrect pin")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "unlock failed")
		return
	}
	s.lockout.Reset()
	s.grantSession(r)
	writeJSON(w, http.StatusOK, statusResponse{Initialized: true, Unlocked: true})
}

// handleLock locks the vault and destroys the session.
func (s *Server) handleLock(w http.ResponseWriter, r *http.Request) {
	s.vault.Lock()
	_ = s.sessions.Destroy(r.Context())
	writeJSON(w, http.StatusOK, statusResponse{Initialized: true, Unlocked: false})
}

// handleMe is a lightweight authenticated probe used by the SPA.
func (s *Server) handleMe(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, statusResponse{Initialized: true, Unlocked: true})
}

// grantSession marks the session authenticated and rotates the token to prevent
// session fixation.
func (s *Server) grantSession(r *http.Request) {
	_ = s.sessions.RenewToken(r.Context())
	s.sessions.Put(r.Context(), sessionAuthKey, true)
}

// validPIN enforces exactly six digits.
func validPIN(pin string) bool {
	if len(pin) != 6 {
		return false
	}
	for _, c := range pin {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
