package db

import (
	"database/sql"
	"errors"
)

func (s *Store) CreateAttachment(a *Attachment) error {
	a.CreatedAt = nowStamp()
	blob, err := s.encrypt(a) // StoredPath/ThumbPath are json:"-" → not sealed (they're columns)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(
		`INSERT INTO attachments (owner_type, owner_id, kind, stored_path, thumb_path, data)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		a.OwnerType, a.OwnerID, a.Kind, a.StoredPath, a.ThumbPath, blob,
	)
	if err != nil {
		return err
	}
	a.ID, err = res.LastInsertId()
	return err
}

func (s *Store) ListAttachments(ownerType string, ownerID int64) ([]Attachment, error) {
	rows, err := s.db.Query(
		`SELECT id, stored_path, thumb_path, data FROM attachments
		 WHERE owner_type = ? AND owner_id = ? ORDER BY id`,
		ownerType, ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Attachment{}
	for rows.Next() {
		a, err := s.scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		a.OwnerType = ownerType
		a.OwnerID = ownerID
		out = append(out, *a)
	}
	return out, rows.Err()
}

func (s *Store) GetAttachment(id int64) (*Attachment, error) {
	a, err := s.scanAttachment(s.db.QueryRow(
		`SELECT id, owner_type, owner_id, stored_path, thumb_path, data FROM attachments WHERE id = ?`, id,
	), true)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) DeleteAttachment(id int64) error {
	res, err := s.db.Exec(`DELETE FROM attachments WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

// scanAttachment decrypts a row. When withOwner is set the SELECT also carried
// owner_type/owner_id (used by GetAttachment, which has no owner context).
func (s *Store) scanAttachment(sc scanner, withOwner ...bool) (*Attachment, error) {
	var (
		id        int64
		ownerType string
		ownerID   int64
		stored    string
		thumb     sql.NullString
		blob      []byte
	)
	var err error
	if len(withOwner) > 0 && withOwner[0] {
		err = sc.Scan(&id, &ownerType, &ownerID, &stored, &thumb, &blob)
	} else {
		err = sc.Scan(&id, &stored, &thumb, &blob)
	}
	if err != nil {
		return nil, err
	}
	var a Attachment
	if err := s.decrypt(blob, &a); err != nil {
		return nil, err
	}
	a.ID = id
	a.StoredPath = stored
	if thumb.Valid {
		a.ThumbPath = thumb.String
	}
	if len(withOwner) > 0 && withOwner[0] {
		a.OwnerType = ownerType
		a.OwnerID = ownerID
	}
	return &a, nil
}
