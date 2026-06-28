-- +goose Up
-- Generalize the API response cache (was gunspec-specific; now keyed by source
-- prefix, e.g. "wiki:search:" / "wiki:page:").
ALTER TABLE gunspec_cache RENAME TO spec_cache;

-- +goose Down
ALTER TABLE spec_cache RENAME TO gunspec_cache;
