-- +goose Up
-- +goose StatementBegin
ALTER TABLE positions ADD COLUMN avg_cost_base REAL NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE positions DROP COLUMN avg_cost_base;
-- +goose StatementEnd
