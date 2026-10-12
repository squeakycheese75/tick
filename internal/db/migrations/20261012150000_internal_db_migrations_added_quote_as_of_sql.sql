-- +goose Up
-- +goose StatementBegin
ALTER TABLE price_cache ADD COLUMN as_of TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE price_cache DROP COLUMN as_of;
-- +goose StatementEnd
