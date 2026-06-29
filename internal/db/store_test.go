package db

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"github.com/TheOutdoorProgrammer/boating-accident/internal/vault"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	database, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	v := vault.New(database, vault.Argon2Params{MemoryKiB: 8 * 1024, Time: 1, Threads: 1})
	if err := v.Setup("123456"); err != nil {
		t.Fatal(err)
	}
	return NewStore(database, v)
}

func TestFirearmCRUDEncrypted(t *testing.T) {
	s := newTestStore(t)

	f := &Firearm{Nickname: "Carry", Manufacturer: "Glock", Model: "19", SerialNumber: "ABC123XYZ"}
	if err := s.CreateFirearm(f); err != nil {
		t.Fatalf("create: %v", err)
	}
	if f.ID == 0 {
		t.Fatal("expected an assigned id")
	}

	// The serial must not appear in the raw on-disk blob.
	var blob []byte
	if err := s.db.QueryRow(`SELECT data FROM firearms WHERE id = ?`, f.ID).Scan(&blob); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(blob, []byte("ABC123XYZ")) {
		t.Fatal("serial number found in plaintext in the stored blob")
	}

	got, err := s.GetFirearm(f.ID)
	if err != nil || got.SerialNumber != "ABC123XYZ" {
		t.Fatalf("get: %v / %+v", err, got)
	}

	got.Nickname = "Nightstand"
	if err := s.UpdateFirearm(got); err != nil {
		t.Fatalf("update: %v", err)
	}
	list, err := s.ListFirearms()
	if err != nil || len(list) != 1 || list[0].Nickname != "Nightstand" {
		t.Fatalf("list: %v / %+v", err, list)
	}

	if err := s.DeleteFirearm(f.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetFirearm(f.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAmmoLinking(t *testing.T) {
	s := newTestStore(t)

	f := &Firearm{Nickname: "AR", Caliber: "5.56"}
	if err := s.CreateFirearm(f); err != nil {
		t.Fatal(err)
	}
	a := &Ammo{Name: "PMC X-TAC", Caliber: "5.56", QuantityOnHand: 200}
	if err := s.CreateAmmo(a); err != nil {
		t.Fatal(err)
	}
	if err := s.LinkAmmo(f.ID, a.ID, "zeroed load"); err != nil {
		t.Fatalf("link: %v", err)
	}

	links, err := s.ListAmmoForFirearm(f.ID)
	if err != nil || len(links) != 1 || links[0].Note != "zeroed load" || links[0].Ammo.Name != "PMC X-TAC" {
		t.Fatalf("links: %v / %+v", err, links)
	}

	if err := s.UnlinkAmmo(f.ID, a.ID); err != nil {
		t.Fatalf("unlink: %v", err)
	}
	if links, _ := s.ListAmmoForFirearm(f.ID); len(links) != 0 {
		t.Fatalf("expected no links, got %d", len(links))
	}
}

func TestAccessoryLinking(t *testing.T) {
	s := newTestStore(t)

	f1 := &Firearm{Nickname: "AR"}
	f2 := &Firearm{Nickname: "AK"}
	if err := s.CreateFirearm(f1); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateFirearm(f2); err != nil {
		t.Fatal(err)
	}
	// Quantity 2 → mountable on two guns.
	acc := &Accessory{Name: "Aimpoint", Category: "optic", Quantity: 2}
	if err := s.CreateAccessory(acc); err != nil {
		t.Fatal(err)
	}

	if err := s.LinkAccessory(f1.ID, acc.ID); err != nil {
		t.Fatalf("link f1: %v", err)
	}
	if err := s.LinkAccessory(f2.ID, acc.ID); err != nil {
		t.Fatalf("link f2: %v", err)
	}
	// Re-linking the same gun is a no-op, not a capacity error.
	if err := s.LinkAccessory(f1.ID, acc.ID); err != nil {
		t.Fatalf("relink f1: %v", err)
	}

	firearms, err := s.ListFirearmsForAccessory(acc.ID)
	if err != nil || len(firearms) != 2 {
		t.Fatalf("firearms for accessory: %v / %+v", err, firearms)
	}
	accs, err := s.ListAccessoriesForFirearm(f1.ID)
	if err != nil || len(accs) != 1 || accs[0].Name != "Aimpoint" {
		t.Fatalf("accessories for firearm: %v / %+v", err, accs)
	}

	// A third gun exceeds the quantity of 2 → ErrAtCapacity.
	f3 := &Firearm{Nickname: "Shotgun"}
	if err := s.CreateFirearm(f3); err != nil {
		t.Fatal(err)
	}
	if err := s.LinkAccessory(f3.ID, acc.ID); !errors.Is(err, ErrAtCapacity) {
		t.Fatalf("expected ErrAtCapacity, got %v", err)
	}

	// Free a slot, then the third fits.
	if err := s.UnlinkAccessory(f1.ID, acc.ID); err != nil {
		t.Fatalf("unlink: %v", err)
	}
	if err := s.LinkAccessory(f3.ID, acc.ID); err != nil {
		t.Fatalf("link f3 after freeing a slot: %v", err)
	}
}
