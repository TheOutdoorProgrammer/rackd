-- +goose Up
-- Accessory↔firearm becomes many-to-many: an accessory with quantity > 1 can be
-- mounted on multiple guns. Mirrors the firearm_ammo join table.
CREATE TABLE firearm_accessory (
    firearm_id   INTEGER NOT NULL REFERENCES firearms(id)    ON DELETE CASCADE,
    accessory_id INTEGER NOT NULL REFERENCES accessories(id) ON DELETE CASCADE,
    data         BLOB,
    PRIMARY KEY (firearm_id, accessory_id)
);

-- Carry existing single assignments over from the old accessories.firearm_id column.
INSERT INTO firearm_accessory (firearm_id, accessory_id)
    SELECT firearm_id, id FROM accessories WHERE firearm_id IS NOT NULL;

-- accessories.firearm_id is now superseded by firearm_accessory and left in place
-- (always NULL going forward). Dropping it would require a full table rebuild in
-- SQLite; we don't do destructive DDL on the live encrypted vault for cosmetics.
-- A later migration can drop it if desired.

-- +goose Down
DROP TABLE firearm_accessory;
