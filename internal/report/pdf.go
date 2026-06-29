// Package report renders a printable PDF snapshot of the whole inventory.
package report

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/TheOutdoorProgrammer/boating-accident/internal/db"
)

// Data is everything the report needs, already decrypted by the caller.
type Data struct {
	Firearms          []db.Firearm
	Ammo              []db.Ammo
	Knives            []db.Knife
	Accessories       []db.Accessory
	AccessoryFirearms map[int64][]int64 // accessory id -> firearm ids it's mounted on
	Generated         time.Time
}

// Build renders the inventory to a PDF document.
func Build(d Data) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "Letter", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("") // UTF-8 -> cp1252 for core fonts
	pdf.SetTitle("Boating Accident inventory report", false)
	pdf.SetMargins(12, 14, 12)
	pdf.SetAutoPageBreak(true, 14)
	pdf.AddPage()

	// fit truncates a translated string to a column width, adding an ellipsis.
	fit := func(s string, w float64) string {
		s = tr(s)
		if pdf.GetStringWidth(s) <= w-2 {
			return s
		}
		r := []rune(s)
		for len(r) > 1 && pdf.GetStringWidth(string(r)+"...") > w-2 {
			r = r[:len(r)-1]
		}
		return string(r) + "..."
	}
	header := func(cols []string, widths []float64) {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(238, 238, 240)
		pdf.SetTextColor(40, 42, 54)
		for i, c := range cols {
			pdf.CellFormat(widths[i], 7, c, "1", 0, "L", true, 0, "")
		}
		pdf.Ln(-1)
	}
	row := func(cells []string, widths []float64) {
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(0, 0, 0)
		for i, c := range cells {
			pdf.CellFormat(widths[i], 6, fit(c, widths[i]), "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}
	section := func(title string) {
		pdf.Ln(4)
		pdf.SetFont("Helvetica", "B", 13)
		pdf.SetTextColor(40, 42, 54)
		pdf.CellFormat(0, 8, title, "", 1, "L", false, 0, "")
	}

	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(40, 42, 54)
	pdf.CellFormat(0, 10, tr("Boating Accident — Inventory Report"), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(110, 110, 120)
	pdf.CellFormat(0, 6, tr("Generated "+d.Generated.Format("Jan 2, 2006 3:04 PM")), "", 1, "L", false, 0, "")

	var total int64
	for _, f := range d.Firearms {
		total += f.AcquiredPriceCents
	}
	for _, a := range d.Ammo {
		total += a.AcquiredPriceCents
	}
	for _, k := range d.Knives {
		total += k.AcquiredPriceCents
	}
	for _, a := range d.Accessories {
		total += a.ValueCents
	}
	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 7, fit(fmt.Sprintf("%d firearms  -  %d ammo lines  -  %d knives  -  %d accessories  -  est. value %s",
		len(d.Firearms), len(d.Ammo), len(d.Knives), len(d.Accessories), money(total)), 191), "", 1, "L", false, 0, "")

	byID := make(map[int64]db.Firearm, len(d.Firearms))
	for _, f := range d.Firearms {
		byID[f.ID] = f
	}

	if len(d.Firearms) > 0 {
		section("Firearms")
		w := []float64{40, 45, 25, 35, 20, 26}
		header([]string{"Name", "Make / Model", "Caliber", "Serial", "Status", "Value"}, w)
		for _, f := range d.Firearms {
			name := f.Nickname
			if name == "" {
				name = joinNonEmpty(" ", f.Manufacturer, f.Model)
			}
			row([]string{
				name,
				joinNonEmpty(" ", f.Manufacturer, f.Model),
				f.Caliber,
				f.SerialNumber,
				titleCase(f.Status),
				money(f.AcquiredPriceCents),
			}, w)
		}
	}

	if len(d.Ammo) > 0 {
		section("Ammo")
		w := []float64{50, 28, 26, 24, 17, 26}
		header([]string{"Name", "Caliber", "Bullet", "On hand", "Low?", "Value"}, w)
		for _, a := range d.Ammo {
			name := a.Name
			if name == "" {
				name = a.Caliber
			}
			low := ""
			if a.LowStockThreshold > 0 && a.QuantityOnHand <= a.LowStockThreshold {
				low = "LOW"
			}
			row([]string{
				name,
				a.Caliber,
				a.BulletType,
				fmt.Sprintf("%d", a.QuantityOnHand),
				low,
				money(a.AcquiredPriceCents),
			}, w)
		}
	}

	if len(d.Knives) > 0 {
		section("Knives")
		w := []float64{45, 28, 40, 32, 26}
		header([]string{"Name", "Type", "Maker", "Steel", "Value"}, w)
		for _, k := range d.Knives {
			name := k.Nickname
			if name == "" {
				name = joinNonEmpty(" ", k.Manufacturer, k.Model)
			}
			row([]string{
				name,
				titleCase(k.Type),
				k.Manufacturer,
				k.BladeSteel,
				money(k.AcquiredPriceCents),
			}, w)
		}
	}

	if len(d.Accessories) > 0 {
		section("Accessories")
		w := []float64{46, 28, 36, 35, 26}
		header([]string{"Name", "Category", "Maker", "On guns", "Value"}, w)
		for _, a := range d.Accessories {
			names := make([]string, 0, len(d.AccessoryFirearms[a.ID]))
			for _, fid := range d.AccessoryFirearms[a.ID] {
				if f, ok := byID[fid]; ok {
					names = append(names, firearmName(f))
				} else {
					names = append(names, fmt.Sprintf("#%d", fid))
				}
			}
			row([]string{
				a.Name,
				titleCase(a.Category),
				a.Manufacturer,
				strings.Join(names, ", "),
				money(a.ValueCents),
			}, w)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func firearmName(f db.Firearm) string {
	if f.Nickname != "" {
		return f.Nickname
	}
	if n := joinNonEmpty(" ", f.Manufacturer, f.Model); n != "" {
		return n
	}
	return fmt.Sprintf("#%d", f.ID)
}

func money(cents int64) string {
	neg := ""
	if cents < 0 {
		neg = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s$%d.%02d", neg, cents/100, cents%100)
}

func joinNonEmpty(sep string, parts ...string) string {
	kept := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, sep)
}

func titleCase(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
