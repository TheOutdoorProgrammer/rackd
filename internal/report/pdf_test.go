package report

import (
	"bytes"
	"testing"
	"time"

	"github.com/TheOutdoorProgrammer/boating-accident/internal/db"
)

func TestBuild(t *testing.T) {
	d := Data{
		Firearms:          []db.Firearm{{ID: 1, Nickname: "Truck Gun", Manufacturer: "Ruger", Model: "10/22", Caliber: ".22 LR", SerialNumber: "ABC123", Status: "owned", AcquiredPriceCents: 24999}},
		Ammo:              []db.Ammo{{ID: 1, Name: "Plinking", Caliber: ".22 LR", BulletType: "LRN", QuantityOnHand: 50, LowStockThreshold: 100, AcquiredPriceCents: 1999}},
		Knives:            []db.Knife{{ID: 1, Nickname: "EDC", Type: "folding", Manufacturer: "Benchmade", BladeSteel: "CPM MagnaCut", AcquiredPriceCents: 18000}},
		Accessories:       []db.Accessory{{ID: 1, Name: "Red Dot", Category: "optic", Manufacturer: "Holosun", ValueCents: 30000}},
		AccessoryFirearms: map[int64][]int64{1: {1}},
		Generated:         time.Unix(0, 0).UTC(),
	}
	out, err := Build(d)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF")) {
		t.Fatalf("output is not a PDF (len=%d)", len(out))
	}
}

func TestBuildEmpty(t *testing.T) {
	out, err := Build(Data{Generated: time.Unix(0, 0).UTC()})
	if err != nil {
		t.Fatalf("Build empty: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF")) {
		t.Fatalf("empty output is not a PDF")
	}
}
