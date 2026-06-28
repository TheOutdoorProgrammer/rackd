-- +goose Up
-- Long-lived cache for GunSpec.io API responses. Entries never expire on their
-- own (firearm specs are static) and are cleared manually, to stay under the
-- free tier's tight rate limit. Data is public catalog info, so it is stored as
-- plaintext (not vault-encrypted like user inventory).
CREATE TABLE gunspec_cache (
    cache_key  TEXT PRIMARY KEY,
    data       TEXT NOT NULL,
    fetched_at TEXT NOT NULL
);

-- +goose Down
DROP TABLE gunspec_cache;
