package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/TheOutdoorProgrammer/rackd/internal/db"
	"github.com/TheOutdoorProgrammer/rackd/internal/images"
)

const (
	maxUpload = 20 << 20 // 20 MiB per photo
	thumbMax  = 400      // px, longest edge
)

var allowedOwners = map[string]bool{
	"firearms": true, "ammo": true, "knives": true, "accessories": true,
}

func (s *Server) uploadsDir() string { return filepath.Join(s.cfg.DataDir, "uploads") }

// ownerRef reads and validates the ?owner=&id= query pair.
func (s *Server) ownerRef(r *http.Request) (string, int64, bool) {
	owner := r.URL.Query().Get("owner")
	if !allowedOwners[owner] {
		return "", 0, false
	}
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		return "", 0, false
	}
	return owner, id, true
}

func (s *Server) ownerExists(owner string, id int64) bool {
	var err error
	switch owner {
	case "firearms":
		_, err = s.store.GetFirearm(id)
	case "ammo":
		_, err = s.store.GetAmmo(id)
	case "knives":
		_, err = s.store.GetKnife(id)
	case "accessories":
		_, err = s.store.GetAccessory(id)
	default:
		return false
	}
	return err == nil
}

func (s *Server) handleListPhotos(w http.ResponseWriter, r *http.Request) {
	owner, id, ok := s.ownerRef(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid owner or id")
		return
	}
	items, err := s.store.ListAttachments(owner, id)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleUploadPhoto(w http.ResponseWriter, r *http.Request) {
	owner, id, ok := s.ownerRef(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid owner or id")
		return
	}
	if !s.ownerExists(owner, id) {
		writeError(w, http.StatusNotFound, "owner not found")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		writeError(w, http.StatusBadRequest, "file too large or malformed")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		serverError(w, err)
		return
	}
	processed, err := images.Process(raw, thumbMax)
	if err != nil {
		writeError(w, http.StatusBadRequest, "unsupported or invalid image")
		return
	}

	if err := os.MkdirAll(s.uploadsDir(), 0o700); err != nil {
		serverError(w, err)
		return
	}
	name, err := randomName()
	if err != nil {
		serverError(w, err)
		return
	}
	stored := name + ".bin"
	thumb := name + ".thumb.bin"
	if err := s.writeEncrypted(stored, processed.Full); err != nil {
		serverError(w, err)
		return
	}
	if err := s.writeEncrypted(thumb, processed.Thumb); err != nil {
		serverError(w, err)
		return
	}

	a := &db.Attachment{
		OwnerType:   owner,
		OwnerID:     id,
		Kind:        "photo",
		Filename:    filepath.Base(header.Filename),
		ContentType: "image/jpeg",
		SizeBytes:   int64(len(processed.Full)),
		StoredPath:  stored,
		ThumbPath:   thumb,
	}
	if err := s.store.CreateAttachment(a); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (s *Server) handleServePhoto(w http.ResponseWriter, r *http.Request) {
	a := s.loadAttachment(w, r)
	if a == nil {
		return
	}
	s.serveEncrypted(w, a.StoredPath)
}

func (s *Server) handleServeThumb(w http.ResponseWriter, r *http.Request) {
	a := s.loadAttachment(w, r)
	if a == nil {
		return
	}
	path := a.ThumbPath
	if path == "" {
		path = a.StoredPath
	}
	s.serveEncrypted(w, path)
}

func (s *Server) handleDeletePhoto(w http.ResponseWriter, r *http.Request) {
	a := s.loadAttachment(w, r)
	if a == nil {
		return
	}
	s.removeFiles(a)
	if err := s.store.DeleteAttachment(a.ID); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

func (s *Server) loadAttachment(w http.ResponseWriter, r *http.Request) *db.Attachment {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return nil
	}
	a, err := s.store.GetAttachment(id)
	if errors.Is(err, db.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return nil
	}
	if err != nil {
		serverError(w, err)
		return nil
	}
	return a
}

func (s *Server) writeEncrypted(name string, data []byte) error {
	enc, err := s.vault.Encrypt(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.uploadsDir(), name), enc, 0o600)
}

func (s *Server) serveEncrypted(w http.ResponseWriter, name string) {
	enc, err := os.ReadFile(filepath.Join(s.uploadsDir(), name))
	if err != nil {
		serverError(w, err)
		return
	}
	data, err := s.vault.Decrypt(enc)
	if err != nil {
		serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private, max-age=86400")
	_, _ = w.Write(data)
}

func (s *Server) removeFiles(a *db.Attachment) {
	_ = os.Remove(filepath.Join(s.uploadsDir(), a.StoredPath))
	if a.ThumbPath != "" {
		_ = os.Remove(filepath.Join(s.uploadsDir(), a.ThumbPath))
	}
}

// deletePhotosForOwner removes all photos (rows + files) for an item, called
// when the item itself is deleted.
func (s *Server) deletePhotosForOwner(owner string, id int64) {
	items, err := s.store.ListAttachments(owner, id)
	if err != nil {
		return
	}
	for i := range items {
		s.removeFiles(&items[i])
		_ = s.store.DeleteAttachment(items[i].ID)
	}
}

func randomName() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
