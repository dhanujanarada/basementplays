-- +goose Up
-- The Go service owns everything under the `ticketing` schema. The migration runner
-- creates the schema before goose starts so its version table can live inside it;
-- this migration records that ownership in the migration history.
CREATE SCHEMA IF NOT EXISTS ticketing;

-- +goose Down
-- Keep the schema: goose's own version table lives in ticketing.goose_db_version.
SELECT 1;
