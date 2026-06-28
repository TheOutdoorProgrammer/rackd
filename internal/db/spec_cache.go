package db

import (
	"database/sql"
	"errors"
)

// SpecCacheGet returns the cached response for key, if present.
func (s *Store) SpecCacheGet(key string) ([]byte, bool, error) {
	var data []byte
	err := s.db.QueryRow(`SELECT data FROM spec_cache WHERE cache_key = ?`, key).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

// SpecCachePut stores (or replaces) a cached response.
func (s *Store) SpecCachePut(key string, data []byte) error {
	_, err := s.db.Exec(
		`INSERT INTO spec_cache (cache_key, data, fetched_at) VALUES (?, ?, ?)
		 ON CONFLICT(cache_key) DO UPDATE SET data = excluded.data, fetched_at = excluded.fetched_at`,
		key, data, nowStamp(),
	)
	return err
}

// SpecCacheCount returns how many responses are cached.
func (s *Store) SpecCacheCount() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM spec_cache`).Scan(&n)
	return n, err
}

// SpecCacheClear deletes every cached response and returns how many were removed.
func (s *Store) SpecCacheClear() (int, error) {
	res, err := s.db.Exec(`DELETE FROM spec_cache`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
