package api

import (
	"net/http"
	"time"

	"github.com/TheOutdoorProgrammer/boating-accident/internal/report"
)

// handleReport streams a PDF snapshot of the whole inventory. The document is
// decrypted plaintext — it's served over the authenticated session only.
func (s *Server) handleReport(w http.ResponseWriter, _ *http.Request) {
	firearms, err := s.store.ListFirearms()
	if err != nil {
		serverError(w, err)
		return
	}
	ammo, err := s.store.ListAmmo()
	if err != nil {
		serverError(w, err)
		return
	}
	knives, err := s.store.ListKnives()
	if err != nil {
		serverError(w, err)
		return
	}
	accessories, err := s.store.ListAccessories(nil)
	if err != nil {
		serverError(w, err)
		return
	}

	now := time.Now()
	pdf, err := report.Build(report.Data{
		Firearms:    firearms,
		Ammo:        ammo,
		Knives:      knives,
		Accessories: accessories,
		Generated:   now,
	})
	if err != nil {
		serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Cache-Control", "no-store") // always regenerate; never serve a stale download
	w.Header().Set("Content-Disposition", `attachment; filename="boating-accident-report-`+now.Format("2006-01-02-1504")+`.pdf"`)
	_, _ = w.Write(pdf)
}
