package db

import (
	"database/sql"
	"errors"
)

func (s *Store) CreateAmmo(a *Ammo) error {
	a.CreatedAt = nowStamp()
	a.UpdatedAt = a.CreatedAt
	blob, err := s.encrypt(a)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`INSERT INTO ammo (data) VALUES (?)`, blob)
	if err != nil {
		return err
	}
	a.ID, err = res.LastInsertId()
	return err
}

func (s *Store) ListAmmo() ([]Ammo, error) {
	rows, err := s.db.Query(`SELECT id, data FROM ammo ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Ammo{}
	for rows.Next() {
		var id int64
		var blob []byte
		if err := rows.Scan(&id, &blob); err != nil {
			return nil, err
		}
		var a Ammo
		if err := s.decrypt(blob, &a); err != nil {
			return nil, err
		}
		a.ID = id
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) GetAmmo(id int64) (*Ammo, error) {
	var blob []byte
	err := s.db.QueryRow(`SELECT data FROM ammo WHERE id = ?`, id).Scan(&blob)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var a Ammo
	if err := s.decrypt(blob, &a); err != nil {
		return nil, err
	}
	a.ID = id
	return &a, nil
}

func (s *Store) UpdateAmmo(a *Ammo) error {
	a.UpdatedAt = nowStamp()
	blob, err := s.encrypt(a)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE ammo SET data = ? WHERE id = ?`, blob, a.ID)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func (s *Store) DeleteAmmo(id int64) error {
	res, err := s.db.Exec(`DELETE FROM ammo WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

// AdjustAmmo changes an ammo line's rounds-on-hand by delta (clamped at zero)
// and returns the updated record. Backs the quick use/refill buttons.
func (s *Store) AdjustAmmo(id, delta int64) (*Ammo, error) {
	a, err := s.GetAmmo(id)
	if err != nil {
		return nil, err
	}
	a.QuantityOnHand += delta
	if a.QuantityOnHand < 0 {
		a.QuantityOnHand = 0
	}
	if err := s.UpdateAmmo(a); err != nil {
		return nil, err
	}
	return a, nil
}

// LinkAmmo associates an ammo line with a firearm (upsert), with an optional note.
func (s *Store) LinkAmmo(firearmID, ammoID int64, note string) error {
	blob, err := s.encrypt(map[string]string{"note": note})
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO firearm_ammo (firearm_id, ammo_id, data) VALUES (?, ?, ?)
		 ON CONFLICT(firearm_id, ammo_id) DO UPDATE SET data = excluded.data`,
		firearmID, ammoID, blob,
	)
	return err
}

func (s *Store) UnlinkAmmo(firearmID, ammoID int64) error {
	res, err := s.db.Exec(`DELETE FROM firearm_ammo WHERE firearm_id = ? AND ammo_id = ?`, firearmID, ammoID)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

// ListAmmoForFirearm returns the ammo lines linked to a firearm, with notes.
func (s *Store) ListAmmoForFirearm(firearmID int64) ([]AmmoLink, error) {
	rows, err := s.db.Query(
		`SELECT a.id, a.data, fa.data
		 FROM firearm_ammo fa JOIN ammo a ON a.id = fa.ammo_id
		 WHERE fa.firearm_id = ? ORDER BY a.id`,
		firearmID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AmmoLink{}
	for rows.Next() {
		var id int64
		var ammoBlob, linkBlob []byte
		if err := rows.Scan(&id, &ammoBlob, &linkBlob); err != nil {
			return nil, err
		}
		var a Ammo
		if err := s.decrypt(ammoBlob, &a); err != nil {
			return nil, err
		}
		a.ID = id
		note := ""
		if len(linkBlob) > 0 {
			var m map[string]string
			if err := s.decrypt(linkBlob, &m); err != nil {
				return nil, err
			}
			note = m["note"]
		}
		out = append(out, AmmoLink{Ammo: a, Note: note})
	}
	return out, rows.Err()
}

// ListFirearmsForAmmo returns the firearms an ammo line is linked to — the
// reverse of ListAmmoForFirearm, so one ammo can serve many guns.
func (s *Store) ListFirearmsForAmmo(ammoID int64) ([]Firearm, error) {
	rows, err := s.db.Query(
		`SELECT f.id, f.data FROM firearm_ammo fa JOIN firearms f ON f.id = fa.firearm_id
		 WHERE fa.ammo_id = ? ORDER BY f.id`,
		ammoID,
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
