-- +goose Up
CREATE EXTENSION IF NOT EXISTS postgis;

-- +goose Down
-- The database image owns PostGIS, so rollback intentionally retains it.
SELECT 1;
