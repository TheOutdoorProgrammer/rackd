-- +goose Up
-- vault_meta holds the envelope-encryption material (not user content).
CREATE TABLE vault_meta (
    id            INTEGER PRIMARY KEY CHECK (id = 1),
    salt          BLOB    NOT NULL,
    argon_memory  INTEGER NOT NULL,
    argon_time    INTEGER NOT NULL,
    argon_threads INTEGER NOT NULL,
    wrapped_dek   BLOB    NOT NULL
);

-- Each item is stored as a single AES-256-GCM-encrypted JSON blob (`data`).
-- Only surrogate keys and relationship references stay in the clear so the
-- database can relate rows; everything that identifies an item is encrypted.

CREATE TABLE firearms (
    id   INTEGER PRIMARY KEY,
    data BLOB NOT NULL
);

CREATE TABLE ammo (
    id   INTEGER PRIMARY KEY,
    data BLOB NOT NULL
);

CREATE TABLE knives (
    id   INTEGER PRIMARY KEY,
    data BLOB NOT NULL
);

CREATE TABLE accessories (
    id         INTEGER PRIMARY KEY,
    firearm_id INTEGER REFERENCES firearms(id) ON DELETE SET NULL,
    data       BLOB NOT NULL
);

CREATE TABLE firearm_ammo (
    firearm_id INTEGER NOT NULL REFERENCES firearms(id) ON DELETE CASCADE,
    ammo_id    INTEGER NOT NULL REFERENCES ammo(id) ON DELETE CASCADE,
    data       BLOB,
    PRIMARY KEY (firearm_id, ammo_id)
);

CREATE TABLE attachments (
    id          INTEGER PRIMARY KEY,
    owner_type  TEXT    NOT NULL, -- firearms | ammo | knives | accessories
    owner_id    INTEGER NOT NULL,
    kind        TEXT    NOT NULL, -- photo
    stored_path TEXT    NOT NULL, -- opaque on-disk name; file bytes are encrypted
    thumb_path  TEXT,
    data        BLOB    NOT NULL  -- encrypted metadata (filename, content type, size)
);

CREATE INDEX idx_attachments_owner ON attachments(owner_type, owner_id);
CREATE INDEX idx_accessories_firearm ON accessories(firearm_id);

-- +goose Down
DROP TABLE attachments;
DROP TABLE firearm_ammo;
DROP TABLE accessories;
DROP TABLE knives;
DROP TABLE ammo;
DROP TABLE firearms;
DROP TABLE vault_meta;
