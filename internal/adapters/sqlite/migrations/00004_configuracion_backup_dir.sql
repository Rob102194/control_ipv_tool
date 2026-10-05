-- +goose Up
ALTER TABLE configuracion ADD COLUMN backup_dir VARCHAR(500) NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE configuracion DROP COLUMN backup_dir;
