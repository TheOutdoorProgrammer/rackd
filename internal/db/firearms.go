package db

import (
	"database/sql"
	"errors"
)

func (s *Store) CreateFirearm(f *Firearm) error {
	f.CreatedAt = nowStamp()
	f.UpdatedAt = f.CreatedAt
	blob, err := s.encrypt(f)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`INSERT INTO firearms (data) VALUES (?)`, blob)
	if err != nil {
		return err
	}
	f.ID, err = res.LastInsertId()
	return err
}

func (s *Store) ListFirearms() ([]Firearm, error) {
	rows, err := s.db.Query(`SELECT id, data FROM firearms ORDER BY id`)
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

func (s *Store) GetFirearm(id int64) (*Firearm, error) {
	var blob []byte
	err := s.db.QueryRow(`SELECT data FROM firearms WHERE id = ?`, id).Scan(&blob)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var f Firearm
	if err := s.decrypt(blob, &f); err != nil {
		return nil, err
	}
	f.ID = id
	return &f, nil
}

func (s *Store) UpdateFirearm(f *Firearm) error {
	f.UpdatedAt = nowStamp()
	blob, err := s.encrypt(f)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE firearms SET data = ? WHERE id = ?`, blob, f.ID)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func (s *Store) DeleteFirearm(id int64) error {
	res, err := s.db.Exec(`DELETE FROM firearms WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return checkAffected(res)
}
