-- +goose Up
CREATE SCHEMA IF NOT EXISTS key_store;

-- +goose Down
DROP SCHEMA IF EXISTS key_store;
