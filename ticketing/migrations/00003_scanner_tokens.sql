-- +goose Up
-- Per-show scanner credentials for the optional admit endpoint. Only the
-- SHA-256 hash of a scanner token is stored; the plaintext is returned once when
-- an internal caller creates or rotates it.
ALTER TABLE ticketing.shows
    ADD COLUMN scanner_token_hash   text,
    ADD COLUMN scanner_token_set_at timestamptz,
    ADD CONSTRAINT shows_scanner_token_hash_not_blank CHECK (
        scanner_token_hash IS NULL OR btrim(scanner_token_hash) <> ''
    );

-- Partial unique index: allows many shows without a scanner token while making
-- the hash lookup unambiguous.
CREATE UNIQUE INDEX shows_scanner_token_hash_idx
    ON ticketing.shows (scanner_token_hash)
    WHERE scanner_token_hash IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS ticketing.shows_scanner_token_hash_idx;
ALTER TABLE ticketing.shows
    DROP CONSTRAINT IF EXISTS shows_scanner_token_hash_not_blank,
    DROP COLUMN IF EXISTS scanner_token_set_at,
    DROP COLUMN IF EXISTS scanner_token_hash;
