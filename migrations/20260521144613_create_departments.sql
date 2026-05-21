-- +goose Up
CREATE TABLE departments (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(200) NOT NULL,
    parent_id  INTEGER REFERENCES departments(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_department_name_parent UNIQUE NULLS NOT DISTINCT (parent_id, name)
);

-- +goose Down
DROP TABLE IF EXISTS departments;
