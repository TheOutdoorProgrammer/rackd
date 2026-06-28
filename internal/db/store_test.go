package db

import (
	"bytes"
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
