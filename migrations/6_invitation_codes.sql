-- +goose Up
CREATE TABLE invitation_codes(
    code uuid PRIMARY KEY,
    project_id BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE invitation_codes;