package api

import (
	"net/http"
	"strings"

	"github.com/TheOutdoorProgrammer/boating-accident/internal/db"
)

type summaryResponse struct {
	Counts          map[string]int `json:"counts"`
	TotalValueCents int64          `json:"totalValueCents"`
	LowStockAmmo    int            `json:"lowStockAmmo"`
}

func (s *Server) handleSummary(w http.ResponseWriter, _ *http.Request) {
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
	accessories, err := s.store.ListAccessories()
	if err != nil {
		serverError(w, err)
		return
	}

	var total int64
	for _, f := range firearms {
		total += f.AcquiredPriceCents
	}
	for _, a := range ammo {
		total += a.AcquiredPriceCents
	}
	for _, k := range knives {
		total += k.AcquiredPriceCents
	}
	for _, a := range accessories {
		total += a.ValueCents
	}

	lowAmmo := 0
	for _, a := range ammo {
		if a.LowStockThreshold > 0 && a.QuantityOnHand <= a.LowStockThreshold {
			lowAmmo++
		}
	}

	writeJSON(w, http.StatusOK, summaryResponse{
		Counts: map[string]int{
			"firearms":    len(firearms),
			"ammo":        len(ammo),
			"knives":      len(knives),
			"accessories": len(accessories),
		},
		TotalValueCents: total,
		LowStockAmmo:    lowAmmo,
	})
}

type searchResponse struct {
	Firearms    []db.Firearm   `json:"firearms"`
	Ammo        []db.Ammo      `json:"ammo"`
	Knives      []db.Knife     `json:"knives"`
	Accessories []db.Accessory `json:"accessories"`
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

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
	accessories, err := s.store.ListAccessories()
	if err != nil {
		serverError(w, err)
		return
	}

	out := searchResponse{Firearms: []db.Firearm{}, Ammo: []db.Ammo{}, Knives: []db.Knife{}, Accessories: []db.Accessory{}}
	if q == "" {
		out.Firearms, out.Ammo, out.Knives, out.Accessories = firearms, ammo, knives, accessories
		writeJSON(w, http.StatusOK, out)
		return
	}

	for _, f := range firearms {
		if matches(q, f.Nickname, f.Manufacturer, f.Model, f.Caliber, f.SerialNumber, f.Notes) {
			out.Firearms = append(out.Firearms, f)
		}
	}
	for _, a := range ammo {
		if matches(q, a.Name, a.Brand, a.Caliber, a.BulletType, a.LotNumber, a.Notes) {
			out.Ammo = append(out.Ammo, a)
		}
	}
	for _, k := range knives {
		if matches(q, k.Nickname, k.Manufacturer, k.Model, k.BladeSteel, k.SerialNumber, k.Notes) {
			out.Knives = append(out.Knives, k)
		}
	}
	for _, a := range accessories {
		if matches(q, a.Name, a.Category, a.Manufacturer, a.Model, a.SerialNumber, a.Notes) {
			out.Accessories = append(out.Accessories, a)
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func matches(q string, fields ...string) bool {
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), q) {
			return true
		}
	}
	return false
}
