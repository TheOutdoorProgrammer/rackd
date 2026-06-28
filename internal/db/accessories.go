package db

import (
	"database/sql"
	"errors"
)

func (s *Store) CreateAccessory(a *Accessory) error {
	a.CreatedAt = nowStamp()
	a.UpdatedAt = a.CreatedAt
	blob, err := s.encrypt(a)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`INSERT INTO accessories (firearm_id, data) VALUES (?, ?)`, nullInt(a.FirearmID), blob)
	if err != nil {
		return err
	}
	a.ID, err = res.LastInsertId()
	return err
}

// ListAccessories returns all accessories, or only those linked to firearmID
// when it is non-nil.
func (s *Store) ListAccessories(firearmID *int64) ([]Accessory, error) {
	query := `SELECT id, firearm_id, data FROM accessories`
	args := []any{}
	if firearmID != nil {
		query += ` WHERE firearm_id = ?`
		args = append(args, *firearmID)
	}
	query += ` ORDER BY id`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Accessory{}
	for rows.Next() {
		var id int64
		var fid sql.NullInt64
		var blob []byte
		if err := rows.Scan(&id, &fid, &blob); err != nil {
			return nil, err
		}
		var a Accessory
		if err := s.decrypt(blob, &a); err != nil {
			return nil, err
		}
		a.ID = id
		a.FirearmID = ptrFromNullInt(fid)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) GetAccessory(id int64) (*Accessory, error) {
	var fid sql.NullInt64
	var blob []byte
	err := s.db.QueryRow(`SELECT firearm_id, data FROM accessories WHERE id = ?`, id).Scan(&fid, &blob)
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
	a.FirearmID = ptrFromNullInt(fid)
	return &a, nil
}

func (s *Store) UpdateAccessory(a *Accessory) error {
	a.UpdatedAt = nowStamp()
	blob, err := s.encrypt(a)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE accessories SET firearm_id = ?, data = ? WHERE id = ?`, nullInt(a.FirearmID), blob, a.ID)
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
