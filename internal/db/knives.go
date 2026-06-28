package db

import (
	"database/sql"
	"errors"
)

func (s *Store) CreateKnife(k *Knife) error {
	k.CreatedAt = nowStamp()
	k.UpdatedAt = k.CreatedAt
	blob, err := s.encrypt(k)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`INSERT INTO knives (data) VALUES (?)`, blob)
	if err != nil {
		return err
	}
	k.ID, err = res.LastInsertId()
	return err
}

func (s *Store) ListKnives() ([]Knife, error) {
	rows, err := s.db.Query(`SELECT id, data FROM knives ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Knife{}
	for rows.Next() {
		var id int64
		var blob []byte
		if err := rows.Scan(&id, &blob); err != nil {
			return nil, err
		}
		var k Knife
		if err := s.decrypt(blob, &k); err != nil {
			return nil, err
		}
		k.ID = id
		out = append(out, k)
	}
	return out, rows.Err()
}

func (s *Store) GetKnife(id int64) (*Knife, error) {
	var blob []byte
	err := s.db.QueryRow(`SELECT data FROM knives WHERE id = ?`, id).Scan(&blob)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var k Knife
	if err := s.decrypt(blob, &k); err != nil {
		return nil, err
	}
	k.ID = id
	return &k, nil
}

func (s *Store) UpdateKnife(k *Knife) error {
	k.UpdatedAt = nowStamp()
	blob, err := s.encrypt(k)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE knives SET data = ? WHERE id = ?`, blob, k.ID)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func (s *Store) DeleteKnife(id int64) error {
	res, err := s.db.Exec(`DELETE FROM knives WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return checkAffected(res)
}
