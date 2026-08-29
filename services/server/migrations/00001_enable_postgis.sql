-- +goose Up
CREATE EXTENSION IF NOT EXISTS postgis;

-- +goose Down
-- PostGIS is provided by the database image and intentionally retained on rollback.
SELECT 1;
