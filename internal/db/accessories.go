package db

import (
	"database/sql"
	"errors"
)

// ErrAtCapacity is returned when assigning an accessory to one more firearm would
// exceed its quantity on hand — you can't mount more physical units than you own.
var ErrAtCapacity = errors.New("db: accessory at capacity")

func (s *Store) CreateAccessory(a *Accessory) error {
	a.CreatedAt = nowStamp()
	a.UpdatedAt = a.CreatedAt
	blob, err := s.encrypt(a)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`INSERT INTO accessories (data) VALUES (?)`, blob)
	if err != nil {
		return err
	}
	a.ID, err = res.LastInsertId()
	return err
}

func (s *Store) ListAccessories() ([]Accessory, error) {
	rows, err := s.db.Query(`SELECT id, data FROM accessories ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Accessory{}
	for rows.Next() {
		var id int64
		var blob []byte
		if err := rows.Scan(&id, &blob); err != nil {
			return nil, err
		}
		var a Accessory
		if err := s.decrypt(blob, &a); err != nil {
			return nil, err
		}
		a.ID = id
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) GetAccessory(id int64) (*Accessory, error) {
	var blob []byte
	err := s.db.QueryRow(`SELECT data FROM accessories WHERE id = ?`, id).Scan(&blob)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var a Accessory
	if err := s.decrypt(blob, &a); err != nil {
		return nil, err
	}
	a.ID = id
	return &a, nil
}

func (s *Store) UpdateAccessory(a *Accessory) error {
	a.UpdatedAt = nowStamp()
	blob, err := s.encrypt(a)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE accessories SET data = ? WHERE id = ?`, blob, a.ID)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func (s *Store) DeleteAccessory(id int64) error {
	res, err := s.db.Exec(`DELETE FROM accessories WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

// --- firearm mounts (many-to-many via firearm_accessory) ---

// LinkAccessory mounts an accessory on a firearm. It is idempotent, and enforces
// the quantity cap: an accessory may be assigned to at most max(1, quantity)
// distinct firearms (one physical unit per gun). Returns ErrAtCapacity when full.
func (s *Store) LinkAccessory(firearmID, accessoryID int64) error {
	acc, err := s.GetAccessory(accessoryID)
	if err != nil {
		return err // ErrNotFound when the accessory is gone
	}

	// Already mounted on this gun → no-op (and must not count against the cap).
	var exists int
	switch err := s.db.QueryRow(
		`SELECT 1 FROM firearm_accessory WHERE firearm_id = ? AND accessory_id = ?`, firearmID, accessoryID,
	).Scan(&exists); {
	case err == nil:
		return nil
	case !errors.Is(err, sql.ErrNoRows):
		return err
	}

	var assigned int64
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM firearm_accessory WHERE accessory_id = ?`, accessoryID,
	).Scan(&assigned); err != nil {
		return err
	}
	capacity := acc.Quantity
	if capacity < 1 {
		capacity = 1
	}
	if assigned >= capacity {
		return ErrAtCapacity
	}

	_, err = s.db.Exec(
		`INSERT INTO firearm_accessory (firearm_id, accessory_id) VALUES (?, ?)`, firearmID, accessoryID,
	)
	return err
}

func (s *Store) UnlinkAccessory(firearmID, accessoryID int64) error {
	res, err := s.db.Exec(`DELETE FROM firearm_accessory WHERE firearm_id = ? AND accessory_id = ?`, firearmID, accessoryID)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

// ListAccessoriesForFirearm returns the accessories mounted on a firearm.
func (s *Store) ListAccessoriesForFirearm(firearmID int64) ([]Accessory, error) {
	rows, err := s.db.Query(
		`SELECT a.id, a.data FROM firearm_accessory fa JOIN accessories a ON a.id = fa.accessory_id
		 WHERE fa.firearm_id = ? ORDER BY a.id`,
		firearmID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Accessory{}
	for rows.Next() {
		var id int64
		var blob []byte
		if err := rows.Scan(&id, &blob); err != nil {
			return nil, err
		}
		var a Accessory
		if err := s.decrypt(blob, &a); err != nil {
			return nil, err
		}
		a.ID = id
		out = append(out, a)
	}
	return out, rows.Err()
}

// ListFirearmsForAccessory returns the firearms an accessory is mounted on — the
// reverse of ListAccessoriesForFirearm, so one accessory can serve many guns.
func (s *Store) ListFirearmsForAccessory(accessoryID int64) ([]Firearm, error) {
	rows, err := s.db.Query(
		`SELECT f.id, f.data FROM firearm_accessory fa JOIN firearms f ON f.id = fa.firearm_id
		 WHERE fa.accessory_id = ? ORDER BY f.id`,
		accessoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Firearm{}
	for rows.Next() {
		var id int64
		var blob []byte
		if err := rows.Scan(&id, &blob); err != nil {
			return nil, err
		}
		var f Firearm
		if err := s.decrypt(blob, &f); err != nil {
			return nil, err
		}
		f.ID = id
		out = append(out, f)
	}
	return out, rows.Err()
}

// AccessoryFirearmLinks returns every accessory→firearm assignment as a map of
// accessory id to the firearm ids it's mounted on. One query for the whole set;
// used by the inventory report.
func (s *Store) AccessoryFirearmLinks() (map[int64][]int64, error) {
	rows, err := s.db.Query(`SELECT accessory_id, firearm_id FROM firearm_accessory ORDER BY accessory_id, firearm_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64][]int64{}
	for rows.Next() {
		var accID, fID int64
		if err := rows.Scan(&accID, &fID); err != nil {
			return nil, err
		}
		out[accID] = append(out[accID], fID)
	}
	return out, rows.Err()
}
